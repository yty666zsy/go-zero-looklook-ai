// Package tools 是 Agent 的工具注册表。
//
// 每个工具 = 一个业务能力 + 一份 JSON Schema + 一段人类可读的摘要生成逻辑。
// 约定：
//   - Read 工具直接执行；
//   - Write 工具带 Write=true，runner 不会直接执行，而是生成「待确认动作」，
//     等用户在前端点确认后才落地（见 internal/agent/runner.go）。
package tools

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"looklook/app/aiagent/cmd/api/internal/biz"
	"looklook/app/aiagent/cmd/api/internal/llm"
)

// Env 是一次工具调用所处的上下文（谁在用、用哪个 token 调业务接口）。
type Env struct {
	Scene          string
	OwnerID        int64
	Token          string
	ConversationID int64
}

type Param struct {
	Type        string
	Description string
	Required    bool
	Enum        []string
}

// Card 是给前端渲染的结构化结果，模型只看 Observation。
type Card struct {
	Kind  string `json:"kind"`
	Title string `json:"title,omitempty"`
	Items any    `json:"items,omitempty"`
}

type Result struct {
	Observation string // 给模型的紧凑文本
	Summary     string // 步骤摘要（前端时间线展示）
	Cards       []Card // 结构化卡片
}

type Tool struct {
	Name        string
	Scene       string
	Write       bool
	Description string
	Params      map[string]Param
	// Summarize 生成「待确认动作」的中文确认文案（写操作必需）
	Summarize func(args map[string]any) string
	Run       func(ctx context.Context, env *Env, args map[string]any) (Result, error)
}

type Registry struct {
	order []string
	tools map[string]*Tool
	biz   *biz.Client
}

func NewRegistry(b *biz.Client) *Registry {
	r := &Registry{tools: map[string]*Tool{}, biz: b}
	registerCustomer(r, b)
	registerAdmin(r, b)
	return r
}

func (r *Registry) add(t *Tool) {
	key := key(t.Scene, t.Name)
	r.tools[key] = t
	r.order = append(r.order, key)
}

func (r *Registry) Get(scene, name string) (*Tool, bool) {
	t, ok := r.tools[key(scene, name)]
	return t, ok
}

// Specs 返回该场景下允许模型使用的工具（JSON Schema），并保证顺序稳定。
func (r *Registry) Specs(scene string) []llm.ToolSpec {
	keys := make([]string, 0, len(r.order))
	for _, k := range r.order {
		if strings.HasPrefix(k, scene+"|") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	specs := make([]llm.ToolSpec, 0, len(keys))
	for _, k := range keys {
		t := r.tools[k]
		specs = append(specs, llm.ToolSpec{
			Name:        t.Name,
			Description: t.Description,
			Parameters: map[string]any{
				"type":       "object",
				"properties": properties(t.Params),
				"required":   required(t.Params),
			},
		})
	}
	return specs
}

func (r *Registry) Names(scene string) []string {
	specs := r.Specs(scene)
	out := make([]string, 0, len(specs))
	for _, s := range specs {
		out = append(out, s.Name)
	}
	return out
}

func key(scene, name string) string { return scene + "|" + name }

func properties(params map[string]Param) map[string]any {
	out := make(map[string]any, len(params))
	for k, p := range params {
		item := map[string]any{"type": p.Type, "description": p.Description}
		if len(p.Enum) > 0 {
			item["enum"] = p.Enum
		}
		out[k] = item
	}
	return out
}

func required(params map[string]Param) []string {
	out := make([]string, 0, len(params))
	for k, p := range params {
		if p.Required {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

// ---------- 参数读取 ----------

func strArg(args map[string]any, k string) string {
	if v, ok := args[k]; ok {
		switch t := v.(type) {
		case string:
			return strings.TrimSpace(t)
		case float64:
			return strconv.FormatFloat(t, 'f', -1, 64)
		case bool:
			return strconv.FormatBool(t)
		}
	}
	return ""
}

func intArg(args map[string]any, k string) (int64, bool) {
	v, ok := args[k]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case string:
		n, err := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		if err != nil {
			return 0, false
		}
		return n, true
	}
	return 0, false
}

func floatArg(args map[string]any, k string) (float64, bool) {
	v, ok := args[k]
	if !ok || v == nil {
		return 0, false
	}
	switch t := v.(type) {
	case float64:
		return t, true
	case int64:
		return float64(t), true
	case int:
		return float64(t), true
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(t), 64)
		if err != nil {
			return 0, false
		}
		return f, true
	}
	return 0, false
}

// ---------- 公共格式化 ----------

var tradeStateText = map[int64]string{
	-1: "已取消",
	0:  "待支付",
	1:  "未使用（已支付）",
	2:  "已使用",
	3:  "已退款",
	4:  "已过期",
}

func stateText(s int64) string {
	if t, ok := tradeStateText[s]; ok {
		return t
	}
	return fmt.Sprintf("未知状态(%d)", s)
}

func money(v float64) string { return fmt.Sprintf("¥%.2f", v) }

// 人数、价格等数值格式化
func itoa(v int64) string { return strconv.FormatInt(v, 10) }
