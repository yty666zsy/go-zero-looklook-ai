package svc

import (
	"context"
	"os"
	"strings"

	"looklook/app/aiagent/cmd/api/internal/agent"
	"looklook/app/aiagent/cmd/api/internal/biz"
	"looklook/app/aiagent/cmd/api/internal/config"
	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/insight"
	"looklook/app/aiagent/cmd/api/internal/llm"
	"looklook/app/aiagent/cmd/api/internal/rag"
	"looklook/app/aiagent/cmd/api/internal/tools"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config  config.Config
	Dao     *dao.Dao
	Biz     *biz.Client
	LLM     llm.Provider
	Tools   *tools.Registry
	KB      *rag.KB
	Runner  *agent.Runner
	Insight *insight.Service
}

func NewServiceContext(c config.Config) *ServiceContext {
	d := dao.New(sqlx.NewMysql(c.AiDB.DataSource))
	b := biz.New(c.TravelAPI.BaseUrl, c.OrderAPI.BaseUrl, c.AdminAPI.BaseUrl, maxTimeout(c))
	kb := rag.NewKB()

	var provider llm.Provider
	apiKey := c.LLM.ApiKey
	if apiKey == "" {
		// 密钥优先从环境变量注入，仓库里不放任何密钥
		apiKey = firstEnv("AI_AGENT_API_KEY", "DEEPSEEK_API_KEY", "OPENAI_API_KEY")
	}
	if c.LLM.Provider == "openai" && apiKey != "" {
		provider = llm.NewOpenAI(c.LLM.BaseUrl, apiKey, c.LLM.Model, c.LLM.Temperature, c.LLM.MaxTokens, c.LLM.TimeoutMs)
	} else {
		if c.LLM.Provider == "openai" {
			logx.Info("LLM ApiKey 未配置（可设环境变量 AI_AGENT_API_KEY），已降级为 mock 模式")
		}
		// 没有 API Key 时退化为本地规则规划器，保证服务开箱可用
		provider = llm.NewMock()
	}

	registry := tools.NewRegistry(b)
	ctx := &ServiceContext{
		Config: c,
		Dao:    d,
		Biz:    b,
		LLM:    provider,
		Tools:  registry,
		KB:     kb,
	}
	ctx.Runner = agent.NewRunner(c, d, provider, registry, kb)
	ctx.Insight = insight.New(c, d, b, registry, provider)
	ctx.ReloadKnowledge(context.Background())
	return ctx
}

// ReloadKnowledge 重建内存中的 RAG 索引（启动时 + 知识库变更时调用）。
func (s *ServiceContext) ReloadKnowledge(ctx context.Context) {
	rows, err := s.Dao.ListKnowledge(ctx)
	if err != nil {
		logx.WithContext(ctx).Errorf("load knowledge err: %v", err)
		return
	}
	docs := make([]rag.Doc, 0, len(rows))
	for _, r := range rows {
		docs = append(docs, rag.Doc{
			Id: r.Id, Scene: r.Scene, Category: r.Category,
			Title: r.Title, Content: r.Content, Keywords: r.Keywords,
		})
	}
	s.KB.Build(docs)
	logx.Infof("knowledge index built: %d docs, provider=%s", len(docs), s.LLM.Name())
}

// firstEnv 返回第一个非空的环境变量值（密钥只走环境变量，不落仓库）。
func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func maxTimeout(c config.Config) int64 {
	t := c.AdminAPI.Timeout
	if c.TravelAPI.Timeout > t {
		t = c.TravelAPI.Timeout
	}
	if c.OrderAPI.Timeout > t {
		t = c.OrderAPI.Timeout
	}
	return t
}
