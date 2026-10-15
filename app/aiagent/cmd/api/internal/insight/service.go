// Package insight 是「主动型 Agent」：不等用户提问，服务自己按节奏跑。
//
// 两块能力：
//
//	report  经营日报   —— 无人值守跑一组只读工具，产出结论、指标卡片与执行步骤
//	patrol  异常巡检   —— 确定性规则扫一遍业务数据，产出可一键处置的告警
//
// 设计取舍：
//   - 采集只用只读接口，且复用现有工具/业务客户端，不重复实现业务逻辑；
//   - 涉及写操作时只生成 ai_action_token 待确认动作，仍需人工在前端点确认，
//     与对话式 Agent 走完全相同的安全通道；
//   - 无人值守调用以内部服务账号（ai-agent）身份签发 JWT，
//     所以 admin-api 侧看到的是一条可追溯的实名调用，而不是匿名特权。
package insight

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"looklook/app/aiagent/cmd/api/internal/biz"
	"looklook/app/aiagent/cmd/api/internal/config"
	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/llm"
	"looklook/app/aiagent/cmd/api/internal/tools"
	"looklook/pkg/ctxdata"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
)

type Service struct {
	cfg   config.Config
	dao   *dao.Dao
	biz   *biz.Client
	tools *tools.Registry
	llm   llm.Provider

	// 定时任务串行执行，避免巡检重入把同一条告警刷成多份
	mu sync.Mutex

	lastPatrolAt   time.Time
	lastPatrolInfo PatrolResult
	lastReportDate string
	lastReportAt   time.Time
}

func New(cfg config.Config, d *dao.Dao, b *biz.Client, t *tools.Registry, p llm.Provider) *Service {
	return &Service{cfg: cfg, dao: d, biz: b, tools: t, llm: p}
}

// LastPatrol 最近一次巡检的概要（给前端显示「值守状态」）。
func (s *Service) LastPatrol() PatrolResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastPatrolInfo
}

// llmPolished 只有真正接了大模型才做「润色」，mock 模式下保持确定性文案。
func (s *Service) llmPolished() bool {
	if s.llm == nil {
		return false
	}
	return !strings.HasPrefix(s.llm.Name(), "mock")
}

// uid 内部服务账号 id。
func (s *Service) uid() int64 {
	if s.cfg.Cron.ServiceUid > 0 {
		return s.cfg.Cron.ServiceUid
	}
	return 1
}

// serviceToken 以内部服务账号签发管理端 JWT。
//
// 无人值守任务没有「发起人」，但也不应该绕过权限模型，
// 所以用预置的 role=service 账号签发：调用会落到 admin-api 的审计里，
// 能追溯到具体账号，而不是一个匿名特权通道。
func (s *Service) serviceToken() (string, error) {
	secret := s.cfg.AdminJwtAuth.AccessSecret
	if secret == "" {
		return "", fmt.Errorf("未配置 AdminJwtAuth.AccessSecret")
	}
	now := time.Now().Unix()
	ttl := s.cfg.AdminJwtAuth.AccessExpire
	if ttl <= 0 {
		ttl = 3600
	}
	claims := jwt.MapClaims{
		ctxdata.CtxKeyJwtUserId: s.uid(),
		"iat":                   now,
		"exp":                   now + ttl,
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// callTool 以服务身份调用一个只读工具，并记一条审计。
// 写工具不在这里直接执行（要人工确认），所以直接拒绝。
func (s *Service) callTool(ctx context.Context, token, name string, args map[string]any) (tools.Result, error) {
	tool, ok := s.tools.Get(dao.SceneAdmin, name)
	if !ok {
		return tools.Result{}, fmt.Errorf("工具不存在：%s", name)
	}
	if tool.Write {
		return tools.Result{}, fmt.Errorf("无人值守任务不允许直接执行写操作：%s", name)
	}
	if args == nil {
		args = map[string]any{}
	}

	timeout := time.Duration(s.cfg.Agent.ToolTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	env := &tools.Env{Scene: dao.SceneAdmin, OwnerID: s.uid(), Token: token}
	start := time.Now()
	res, err := tool.Run(cctx, env, args)
	cost := time.Since(start).Milliseconds()

	status := "ok"
	obs := res.Observation
	if err != nil {
		status = "fail"
		obs = err.Error()
	}
	_ = s.dao.AppendAudit(ctx, 0, dao.SceneAdmin, s.uid(), name, args, obs, status, cost)
	return res, err
}

// ---------- 执行步骤（日报会落库，前端画时间线） ----------

type Step struct {
	Index   int    `json:"index"`
	Type    string `json:"type"`
	Tool    string `json:"tool"`
	Args    any    `json:"args,omitempty"`
	Status  string `json:"status"`
	Summary string `json:"summary,omitempty"`
	CostMs  int64  `json:"costMs,omitempty"`
	Error   string `json:"error,omitempty"`
}

// collector 累积一次采集的步骤、各工具的结论与卡片。
type collector struct {
	steps  []Step
	cards  []tools.Card
	byTool map[string]tools.Result
}

// obs 取某个工具的结论文本（工具失败时返回空串）。
func (c *collector) obs(tool string) string {
	return strings.TrimSpace(c.byTool[tool].Observation)
}

type planItem struct {
	tool string
	args map[string]any
}

// runTools 按固定顺序跑一组只读工具；单个工具失败不影响整体（记录到步骤里）。
func (s *Service) runTools(ctx context.Context, token string, plan []planItem) *collector {
	c := &collector{byTool: map[string]tools.Result{}}
	for _, p := range plan {
		args := p.args
		if args == nil {
			args = map[string]any{}
		}
		start := time.Now()
		res, err := s.callTool(ctx, token, p.tool, args)
		cost := time.Since(start).Milliseconds()

		st := Step{Index: len(c.steps) + 1, Type: "tool", Tool: p.tool, Args: args, CostMs: cost}
		if err != nil {
			st.Status = "fail"
			st.Error = err.Error()
			st.Summary = "调用失败"
			c.steps = append(c.steps, st)
			logx.WithContext(ctx).Errorf("insight tool %s err: %v", p.tool, err)
			continue
		}
		st.Status = "done"
		st.Summary = res.Summary
		c.steps = append(c.steps, st)
		c.byTool[p.tool] = res
		c.cards = append(c.cards, res.Cards...)
	}
	return c
}

// firstCardWithKind 取指定 kind 的卡片（日报用它复用对话里的卡片渲染）。
func (c *collector) firstCardWithKind(kind string) []tools.Card {
	out := make([]tools.Card, 0, 1)
	for _, card := range c.cards {
		if card.Kind == kind {
			out = append(out, card)
		}
	}
	return out
}
