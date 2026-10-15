// Package llm 抽象大模型调用。
//
// 两个实现：
//   - openai：任何 OpenAI 兼容网关（DeepSeek / 通义 / vLLM / Ollama），支持流式与 function calling
//   - mock：本地规则规划器。没有 API Key 时也能完整跑通「ReAct 工具调用 + RAG 抽取式回答」，
//     保证项目开箱即用、可被自动化测试覆盖
package llm

import (
	"context"
	"encoding/json"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Message 兼容 OpenAI 的对话消息结构
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallId string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	Id       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolSpec 描述一个可被模型调用的工具（JSON Schema）
type ToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// MarshalJSON 包装成 OpenAI 规范要求的 {"type":"function","function":{...}}。
// 真实网关（OpenAI / DeepSeek）要求 tools[i] 必须带 type 字段，
// 直接下发扁平的 name/description/parameters 会被判为参数非法（400/422）。
// 内部代码仍按扁平字段使用，只在序列化时加信封。
func (s ToolSpec) MarshalJSON() ([]byte, error) {
	type function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters"`
	}
	return json.Marshal(struct {
		Type     string   `json:"type"`
		Function function `json:"function"`
	}{
		Type: "function",
		Function: function{
			Name:        s.Name,
			Description: s.Description,
			Parameters:  s.Parameters,
		},
	})
}

type Request struct {
	Messages    []Message
	Tools       []ToolSpec
	Temperature float64
	MaxTokens   int
	// Scene 供 mock provider 做场景化意图识别
	Scene string
	// Knowledge 是 RAG 检索到的知识片段，mock provider 用它做抽取式回答
	Knowledge []string
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
}

type Response struct {
	Content   string
	ToolCalls []ToolCall
	Usage     Usage
	Model     string
}

// Provider 大模型提供方。onDelta 用于流式吐出正文增量（工具调用阶段通常为空）。
type Provider interface {
	Name() string
	ChatStream(ctx context.Context, req Request, onDelta func(string) error) (*Response, error)
}
