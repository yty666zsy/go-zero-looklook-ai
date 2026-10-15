package rag

import "sync"

// KB 是知识库在内存中的索引，支持热更新（后台修改知识库后调用 Build 重建）。
type KB struct {
	mu  sync.RWMutex
	idx *Index
	n   int
}

func NewKB() *KB { return &KB{} }

func (k *KB) Build(docs []Doc) {
	idx := Build(docs)
	k.mu.Lock()
	k.idx = idx
	k.n = len(docs)
	k.mu.Unlock()
}

func (k *KB) Search(query string, topK int, minScore float64) []Hit {
	k.mu.RLock()
	idx := k.idx
	k.mu.RUnlock()
	if idx == nil {
		return nil
	}
	return idx.Search(query, topK, minScore)
}

func (k *KB) Size() int {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return k.n
}
