// Package agent 是 Agent 运行时：ReAct 循环 + 工具调度 + 事件流。
package agent

// Event 是推给前端的流式事件（SSE 的 data 部分）。
//
// 事件类型：
//
//	meta    会话已就绪，携带 conversationId
//	step    一次工具调用/思考步骤的开始与结束（前端画时间线）
//	delta   正文增量（打字机效果）
//	retract 撤回此前发过的正文（模型在调工具前的过渡语，应丢弃）
//	card    结构化卡片（房源/订单/指标/趋势）
//	action  待确认的写操作（前端弹确认框）
//	done    本轮结束，携带 usage / messageId
//	error   出错
type Event struct {
	Type           string `json:"type"`
	ConversationId int64  `json:"conversationId,omitempty"`
	MessageId      int64  `json:"messageId,omitempty"`
	Step           *Step  `json:"step,omitempty"`
	Text           string `json:"text,omitempty"`
	Card           *Card  `json:"card,omitempty"`
	Action         *Action `json:"action,omitempty"`
	Model          string `json:"model,omitempty"`
	PromptTokens   int    `json:"promptTokens,omitempty"`
	CompletionTokens int  `json:"completionTokens,omitempty"`
	Message        string `json:"message,omitempty"`
}

type Step struct {
	Index   int    `json:"index"`
	Type    string `json:"type"`             // thought | tool
	Tool    string `json:"tool,omitempty"`   // 工具名
	Args    any    `json:"args,omitempty"`   // 入参
	Status  string `json:"status"`           // running | done | fail | pending
	Summary string `json:"summary,omitempty"`
	CostMs  int64  `json:"costMs,omitempty"`
	Error   string `json:"error,omitempty"`
}

type Card struct {
	Kind  string `json:"kind"`
	Title string `json:"title,omitempty"`
	Items any    `json:"items,omitempty"`
}

// Action 待确认的写操作
type Action struct {
	Token       string `json:"token"`
	Tool        string `json:"tool"`
	Args        any    `json:"args,omitempty"`
	ConfirmText string `json:"confirmText"`
	ExpireIn    int64  `json:"expireIn"` // 秒
}

// EmitFunc 由传输层（SSE）实现
type EmitFunc func(Event) error
