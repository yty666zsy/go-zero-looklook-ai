package agent

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"looklook/app/aiagent/cmd/api/internal/config"
	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/llm"
	"looklook/app/aiagent/cmd/api/internal/rag"
	"looklook/app/aiagent/cmd/api/internal/tools"

	"github.com/zeromicro/go-zero/core/logx"
)

type Runner struct {
	cfg   config.Config
	dao   *dao.Dao
	llm   llm.Provider
	tools *tools.Registry
	kb    *rag.KB
}

func NewRunner(cfg config.Config, d *dao.Dao, p llm.Provider, t *tools.Registry, kb *rag.KB) *Runner {
	return &Runner{cfg: cfg, dao: d, llm: p, tools: t, kb: kb}
}

// RunReq 一次对话请求
type RunReq struct {
	Scene          string
	OwnerID        int64
	Token          string // 透传给业务接口的 JWT
	ConversationID int64
	Input          string
}

// Run 执行一轮对话：ReAct 循环，边跑边通过 emit 推事件。
func (r *Runner) Run(ctx context.Context, in RunReq, emit EmitFunc) error {
	started := time.Now()
	input := strings.TrimSpace(in.Input)
	if input == "" {
		return emit(Event{Type: "error", Message: "内容不能为空"})
	}

	// 1) 会话
	convId, err := r.ensureConversation(ctx, in, input)
	if err != nil {
		return emit(Event{Type: "error", Message: "创建会话失败：" + err.Error()})
	}
	if err := emit(Event{Type: "meta", ConversationId: convId}); err != nil {
		return err
	}

	// 2) RAG：检索知识库（政策/FAQ/运营口径），注入系统提示词
	knowledge := r.retrieve(input)

	// 3) 组装上下文
	history, err := r.dao.ListMessages(ctx, convId, int64(r.cfg.Agent.MaxHistory))
	if err != nil {
		logx.WithContext(ctx).Errorf("load history err: %v", err)
	}
	msgs := make([]llm.Message, 0, len(history)+2)
	msgs = append(msgs, llm.Message{Role: llm.RoleSystem, Content: systemPrompt(in.Scene, knowledge)})
	for _, h := range history {
		if h.Role != llm.RoleUser && h.Role != llm.RoleAssistant {
			continue
		}
		if strings.TrimSpace(h.Content) == "" {
			continue
		}
		msgs = append(msgs, llm.Message{Role: h.Role, Content: h.Content})
	}
	msgs = append(msgs, llm.Message{Role: llm.RoleUser, Content: input})

	if _, err := r.dao.AppendMessage(ctx, dao.MessageWrite{
		ConversationId: convId, Scene: in.Scene, OwnerId: in.OwnerID,
		Role: llm.RoleUser, Content: input,
	}); err != nil {
		logx.WithContext(ctx).Errorf("save user message err: %v", err)
	}

	// 4) ReAct 循环
	specs := r.tools.Specs(in.Scene)
	env := &tools.Env{Scene: in.Scene, OwnerID: in.OwnerID, Token: in.Token, ConversationID: convId}

	var (
		steps     []Step
		cards     []Card
		finalText strings.Builder
		usage     llm.Usage
		model     string
		pending   *Action
		stepSeq   int
	)

	for step := 1; step <= r.cfg.Agent.MaxSteps; step++ {
		var roundText strings.Builder
		resp, err := r.llm.ChatStream(ctx, llm.Request{
			Messages:  msgs,
			Tools:     specs,
			Scene:     in.Scene,
			Knowledge: knowledge,
		}, func(delta string) error {
			roundText.WriteString(delta)
			finalText.WriteString(delta)
			return emit(Event{Type: "delta", Text: delta})
		})
		if err != nil {
			_ = emit(Event{Type: "error", Message: "模型调用失败：" + err.Error()})
			break
		}
		if resp.Model != "" {
			model = resp.Model
		}
		usage.PromptTokens += resp.Usage.PromptTokens
		usage.CompletionTokens += resp.Usage.CompletionTokens

		// 这一轮要调工具时，模型常先吐一句过渡语（如 "I'll look up that order first."）。
		// 正文增量已经实时发给前端了，这里补一个 retract 事件让前端撤回这段，
		// 只保留「工具执行时间线 + 最终回答」，落库正文同去。
		if len(resp.ToolCalls) > 0 && roundText.Len() > 0 {
			preamble := roundText.String()
			if err := emit(Event{Type: "retract", Text: preamble}); err != nil {
				return err
			}
			keep := strings.TrimSuffix(finalText.String(), preamble)
			finalText.Reset()
			finalText.WriteString(keep)
		}

		if len(resp.ToolCalls) == 0 {
			break
		}

		msgs = append(msgs, llm.Message{Role: llm.RoleAssistant, Content: resp.Content, ToolCalls: resp.ToolCalls})

		stop := false
		for _, tc := range resp.ToolCalls {
			stepSeq++
			tool, ok := r.tools.Get(in.Scene, tc.Function.Name)
			if !ok {
				msgs = append(msgs, llm.Message{
					Role: llm.RoleTool, ToolCallId: tc.Id,
					Content: "工具 " + tc.Function.Name + " 在当前场景不可用，请不要调用它。",
				})
				continue
			}
			args := parseArgs(tc.Function.Arguments)

			// 写操作：不执行，生成待确认动作
			if tool.Write {
				summary := tool.Summarize(args)
				token := newToken()
				expireIn := r.cfg.Agent.ActionTTLSec
				if expireIn <= 0 {
					expireIn = 300
				}
				if err := r.dao.CreateActionToken(ctx, dao.ActionToken{
					Token: token, Scene: in.Scene, OwnerId: in.OwnerID, ConversationId: convId,
					Tool: tool.Name, Args: sqlNull(args), Summary: summary,
					ExpireTime: time.Now().Add(time.Duration(expireIn) * time.Second),
				}); err != nil {
					logx.WithContext(ctx).Errorf("create action token err: %v", err)
				}
				_ = r.dao.AppendAudit(ctx, convId, in.Scene, in.OwnerID, tool.Name, args, summary, "pending", 0)

				pending = &Action{Token: token, Tool: tool.Name, Args: args, ConfirmText: summary, ExpireIn: expireIn}
				steps = append(steps, Step{Index: stepSeq, Type: "tool", Tool: tool.Name, Args: args, Status: "pending", Summary: summary})
				if err := emit(Event{Type: "step", Step: &steps[len(steps)-1]}); err != nil {
					return err
				}
				if err := emit(Event{Type: "action", Action: pending}); err != nil {
					return err
				}
				msgs = append(msgs, llm.Message{
					Role: llm.RoleTool, ToolCallId: tc.Id,
					Content: "已生成待确认动作（尚未执行）：" + summary + "。请提示用户点击确认。",
				})
				stop = true
				continue
			}

			// 读操作：直接执行
			running := Step{Index: stepSeq, Type: "tool", Tool: tool.Name, Args: args, Status: "running"}
			if err := emit(Event{Type: "step", Step: &running}); err != nil {
				return err
			}
			costStart := time.Now()
			res, err := r.callTool(ctx, tool, env, args)
			cost := time.Since(costStart).Milliseconds()

			done := Step{Index: stepSeq, Type: "tool", Tool: tool.Name, Args: args, CostMs: cost}
			if err != nil {
				done.Status = "fail"
				done.Error = err.Error()
				done.Summary = "调用失败"
				_ = r.dao.AppendAudit(ctx, convId, in.Scene, in.OwnerID, tool.Name, args, err.Error(), "fail", cost)
				msgs = append(msgs, llm.Message{
					Role: llm.RoleTool, ToolCallId: tc.Id,
					Content: "工具执行失败：" + err.Error(),
				})
			} else {
				done.Status = "done"
				done.Summary = res.Summary
				_ = r.dao.AppendAudit(ctx, convId, in.Scene, in.OwnerID, tool.Name, args, res.Observation, "ok", cost)
				msgs = append(msgs, llm.Message{Role: llm.RoleTool, ToolCallId: tc.Id, Content: res.Observation})
				for i := range res.Cards {
					c := Card{Kind: res.Cards[i].Kind, Title: res.Cards[i].Title, Items: res.Cards[i].Items}
					cards = append(cards, c)
					if err := emit(Event{Type: "card", Card: &c}); err != nil {
						return err
					}
				}
			}
			steps = append(steps, done)
			if err := emit(Event{Type: "step", Step: &steps[len(steps)-1]}); err != nil {
				return err
			}
		}
		if stop {
			break
		}
	}

	// 5) 兜底文案
	text := finalText.String()
	if pending != nil && strings.TrimSpace(text) == "" {
		text = "我已准备好下面这个操作，但需要你确认后才会执行：\n\n- " + pending.ConfirmText
		_ = emit(Event{Type: "delta", Text: text})
	}
	if strings.TrimSpace(text) == "" {
		text = "抱歉，我这轮没有拿到有效结果，换个说法再试试？"
		_ = emit(Event{Type: "delta", Text: text})
	}

	// 6) 落库助手消息
	msgId, err := r.dao.AppendMessage(ctx, dao.MessageWrite{
		ConversationId: convId, Scene: in.Scene, OwnerId: in.OwnerID,
		Role: llm.RoleAssistant, Content: text,
		Steps: steps, Cards: cards, Model: model,
		PromptTokens:     int64(usage.PromptTokens),
		CompletionTokens: int64(usage.CompletionTokens),
		LatencyMs:        time.Since(started).Milliseconds(),
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("save assistant message err: %v", err)
	}
	if err := r.dao.TouchConversation(ctx, convId, text); err != nil {
		logx.WithContext(ctx).Errorf("touch conversation err: %v", err)
	}

	return emit(Event{
		Type:             "done",
		ConversationId:   convId,
		MessageId:        msgId,
		Model:            model,
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
	})
}

// ConfirmAction 执行/取消一个待确认的写操作。
func (r *Runner) ConfirmAction(ctx context.Context, scene string, ownerId int64, authToken, token string, confirm bool) ([]tools.Card, string, error) {
	act, err := r.dao.GetActionToken(ctx, token, scene, ownerId)
	if err != nil {
		return nil, "", err
	}
	if act == nil {
		return nil, "", fmt.Errorf("待确认动作不存在或不属于当前账号")
	}
	if act.Status != dao.ActionStatusPending {
		return nil, "", fmt.Errorf("该动作已处理过，请重新发起")
	}
	if time.Now().After(act.ExpireTime) {
		_, _ = r.dao.MarkActionToken(ctx, token, dao.ActionStatusExpired)
		return nil, "", fmt.Errorf("确认已过期，请重新发起操作")
	}

	tool, ok := r.tools.Get(scene, act.Tool)
	if !ok || !tool.Write {
		return nil, "", fmt.Errorf("动作不可执行：%s", act.Tool)
	}
	args := parseArgs(act.Args.String)

	if !confirm {
		if _, err := r.dao.MarkActionToken(ctx, token, dao.ActionStatusRejected); err != nil {
			return nil, "", err
		}
		_ = r.dao.AppendAudit(ctx, act.ConversationId, scene, ownerId, act.Tool, args, "用户已取消", "rejected", 0)
		return nil, "已取消该操作，未对系统做任何改动。", nil
	}

	res, err := r.callTool(ctx, tool, &tools.Env{Scene: scene, OwnerID: ownerId, Token: authToken, ConversationID: act.ConversationId}, args)
	if err != nil {
		_ = r.dao.AppendAudit(ctx, act.ConversationId, scene, ownerId, act.Tool, args, err.Error(), "fail", 0)
		return nil, "", err
	}
	if _, err := r.dao.MarkActionToken(ctx, token, dao.ActionStatusDone); err != nil {
		return nil, "", err
	}
	_ = r.dao.AppendAudit(ctx, act.ConversationId, scene, ownerId, act.Tool, args, res.Observation, "ok", 0)

	// 把执行结果写回会话，保证多轮上下文一致
	if act.ConversationId > 0 {
		_, _ = r.dao.AppendMessage(ctx, dao.MessageWrite{
			ConversationId: act.ConversationId, Scene: scene, OwnerId: ownerId,
			Role: llm.RoleAssistant, Content: "【操作已执行】" + res.Summary + "\n\n" + res.Observation,
		})
		_ = r.dao.TouchConversation(ctx, act.ConversationId, res.Summary)
	}
	return res.Cards, res.Summary + "。" + res.Observation, nil
}

// ---------- 内部 ----------

func (r *Runner) callTool(ctx context.Context, tool *tools.Tool, env *tools.Env, args map[string]any) (tools.Result, error) {
	timeout := time.Duration(r.cfg.Agent.ToolTimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	type out struct {
		res tools.Result
		err error
	}
	ch := make(chan out, 1)
	go func() {
		defer func() {
			if rec := recover(); rec != nil {
				ch <- out{err: fmt.Errorf("工具内部异常: %v", rec)}
			}
		}()
		res, err := tool.Run(cctx, env, args)
		ch <- out{res: res, err: err}
	}()
	select {
	case o := <-ch:
		return o.res, o.err
	case <-cctx.Done():
		return tools.Result{}, fmt.Errorf("工具执行超时（%s）", timeout)
	}
}

func (r *Runner) ensureConversation(ctx context.Context, in RunReq, input string) (int64, error) {
	if in.ConversationID > 0 {
		conv, err := r.dao.GetConversation(ctx, in.ConversationID, in.Scene, in.OwnerID)
		if err != nil {
			return 0, err
		}
		if conv != nil {
			return conv.Id, nil
		}
	}
	return r.dao.CreateConversation(ctx, in.Scene, in.OwnerID, title(input))
}

func (r *Runner) retrieve(query string) []string {
	if r.kb == nil {
		return nil
	}
	hits := r.kb.Search(query, r.cfg.Rag.TopK, r.cfg.Rag.MinScore)
	out := make([]string, 0, len(hits))
	for _, h := range hits {
		out = append(out, "《"+h.Doc.Title+"》"+clip(h.Doc.Content, r.cfg.Rag.MaxDocChars))
	}
	return out
}

func parseArgs(s string) map[string]any {
	s = strings.TrimSpace(s)
	if s == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return map[string]any{}
	}
	return m
}

func sqlNull(v any) sql.NullString {
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" || string(b) == "{}" {
		return sql.NullString{}
	}
	return sql.NullString{String: string(b), Valid: true}
}

func newToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func title(s string) string { return clip(strings.TrimSpace(s), 40) }

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
