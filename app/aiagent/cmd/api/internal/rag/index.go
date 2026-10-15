// Package rag 是轻量级的检索增强（RAG）实现。
//
// 取舍：知识库规模在千条量级，引入向量库（Milvus/Qdrant）的运维成本
// 远大于收益，因此这里用「中文 bigram 分词 + BM25 + 标题/关键词加权」，
// 全部在内存里完成，零外部依赖；如果后续要接向量检索，只需替换 Index 的实现。
package rag

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

type Doc struct {
	Id       int64
	Scene    string
	Category string
	Title    string
	Content  string
	Keywords string
}

type Hit struct {
	Doc   Doc
	Score float64
}

const (
	k1 = 1.2
	b  = 0.75
)

type Index struct {
	docs      []Doc
	tokens    [][]string
	df        map[string]int
	docLen    []int
	avgDocLen float64
}

func Build(docs []Doc) *Index {
	idx := &Index{
		docs:   docs,
		tokens: make([][]string, 0, len(docs)),
		df:     make(map[string]int),
		docLen: make([]int, 0, len(docs)),
	}
	total := 0
	for _, d := range docs {
		// 标题重复 3 次、关键词重复 2 次，起到加权作用
		var sb strings.Builder
		sb.WriteString(d.Title)
		sb.WriteString(" ")
		sb.WriteString(d.Title)
		sb.WriteString(" ")
		sb.WriteString(d.Title)
		sb.WriteString(" ")
		sb.WriteString(strings.ReplaceAll(d.Keywords, ",", " "))
		sb.WriteString(" ")
		sb.WriteString(strings.ReplaceAll(d.Keywords, ",", " "))
		sb.WriteString(" ")
		sb.WriteString(d.Content)

		tk := tokenize(sb.String())
		idx.tokens = append(idx.tokens, tk)
		idx.docLen = append(idx.docLen, len(tk))
		total += len(tk)

		seen := make(map[string]struct{}, len(tk))
		for _, t := range tk {
			if _, ok := seen[t]; ok {
				continue
			}
			seen[t] = struct{}{}
			idx.df[t]++
		}
	}
	if len(docs) > 0 {
		idx.avgDocLen = float64(total) / float64(len(docs))
	}
	return idx
}

// Search 返回得分最高的 topK 条；低于 minScore 的会被过滤。
func (i *Index) Search(query string, topK int, minScore float64) []Hit {
	if i == nil || len(i.docs) == 0 {
		return nil
	}
	if topK <= 0 {
		topK = 4
	}
	qt := tokenize(query)
	if len(qt) == 0 {
		return nil
	}
	qset := make(map[string]struct{}, len(qt))
	for _, t := range qt {
		qset[t] = struct{}{}
	}

	hits := make([]Hit, 0, len(i.docs))
	N := float64(len(i.docs))
	for di, tk := range i.tokens {
		tf := make(map[string]int, len(tk))
		for _, t := range tk {
			tf[t]++
		}
		var score float64
		dl := float64(i.docLen[di])
		for q := range qset {
			f, ok := tf[q]
			if !ok {
				continue
			}
			n := float64(i.df[q])
			if n == 0 {
				continue
			}
			idf := math.Log(1 + (N-n+0.5)/(n+0.5))
			denom := float64(f) + k1*(1-b+b*dl/max(i.avgDocLen, 1))
			score += idf * (float64(f) * (k1 + 1)) / denom
		}
		if score > 0 {
			hits = append(hits, Hit{Doc: i.docs[di], Score: score})
		}
	}

	sort.Slice(hits, func(a, c int) bool { return hits[a].Score > hits[c].Score })
	if len(hits) > topK {
		hits = hits[:topK]
	}
	if minScore > 0 {
		out := hits[:0]
		for _, h := range hits {
			if h.Score >= minScore {
				out = append(out, h)
			}
		}
		hits = out
	}
	return hits
}

// tokenize 中文按「单字 + 相邻二字」切分（bigram 对中文检索效果稳定且无需词典），
// 英文/数字按连续字母数字串切分，其余字符忽略。
func tokenize(s string) []string {
	s = strings.ToLower(s)
	out := make([]string, 0, len(s))
	var latin strings.Builder

	flush := func() {
		if latin.Len() > 0 {
			out = append(out, latin.String())
			latin.Reset()
		}
	}

	var prevCJK rune
	for _, r := range s {
		switch {
		case unicode.Is(unicode.Han, r):
			flush()
			out = append(out, string(r))
			if prevCJK != 0 {
				out = append(out, string(prevCJK)+string(r))
			}
			prevCJK = r
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			prevCJK = 0
			latin.WriteRune(r)
		default:
			prevCJK = 0
			flush()
		}
	}
	flush()
	return out
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
