package llm

import (
	"context"
	"regexp"
	"strconv"
	"strings"
)

// MockProvider 是一个「规则规划器」：在没有大模型 API Key 的环境下，
// 依然能完整地跑通 Agent 的 ReAct 流程——意图识别 → 生成工具调用 →
// 读取工具结果 → 组织回答；知识类问题则用 RAG 召回的片段做抽取式回答。
//
// 它对外的行为（产出 tool_calls / 流式正文 / usage）与真实模型完全一致，
// 因此上层 runner、SSE 协议、前端交互都不需要为它写任何分支。
type MockProvider struct{}

func NewMock() *MockProvider { return &MockProvider{} }

func (m *MockProvider) Name() string { return "mock:rule-planner" }

var (
	// 订单号形如 HSO2026101516515335601734：允许 0~4 位字母前缀 + 14 位以上数字
	snRe      = regexp.MustCompile(`(?i)\b[A-Z]{0,4}\d{14,32}\b`)
	idRe      = regexp.MustCompile(`(?i)(?:id|编号|民宿|房源|用户|订单)\s*[:：#]?\s*(\d{1,10})`)
	daysRe    = regexp.MustCompile(`(7|14|30|九十|三十|七|十四)\s*天`)
	knownCity = []string{"北京", "上海", "杭州", "成都", "厦门", "丽江", "三亚", "大理"}
)

func (m *MockProvider) ChatStream(ctx context.Context, req Request, onDelta func(string) error) (*Response, error) {
	last := lastMessage(req.Messages)

	// 阶段二：已经拿到工具结果 → 组织最终回答
	if last != nil && last.Role == RoleTool {
		answer := m.compose(req, observations(req.Messages))
		return m.stream(answer, onDelta), nil
	}

	// 阶段一：意图识别 → 生成工具调用
	question := ""
	if last != nil && last.Role == RoleUser {
		question = last.Content
	}

	tools := map[string]bool{}
	for _, t := range req.Tools {
		tools[t.Name] = true
	}

	if call, ask := m.plan(req.Scene, question, tools); call != nil {
		return &Response{
			Model:     m.Name(),
			ToolCalls: []ToolCall{*call},
			Usage:     Usage{PromptTokens: len([]rune(question))/2 + 20},
		}, nil
	} else if ask != "" {
		return m.stream(ask, onDelta), nil
	}

	// 没有命中任何工具 → 用 RAG 召回的知识作答
	return m.stream(m.answerFromKnowledge(req, question), onDelta), nil
}

// ---------- 规划 ----------

func (m *MockProvider) plan(scene, q string, tools map[string]bool) (*ToolCall, string) {
	if strings.TrimSpace(q) == "" {
		return nil, "你想了解点什么？可以直接问「上海有没有适合 2 人的民宿」，或者「帮我看看我的订单」。"
	}
	if scene == "admin" {
		return m.planAdmin(q, tools)
	}
	return m.planCustomer(q, tools)
}

func (m *MockProvider) planCustomer(q string, tools map[string]bool) (*ToolCall, string) {
	sn := findSn(q)

	switch {
	case has(tools, "cancel_order") && containsAny(q, "取消", "退订", "不要了", "退掉"):
		if sn != "" {
			return call("cancel_order", map[string]any{"sn": sn}), ""
		}
		// 「取消订单有什么规则」这类是在问政策，不是要执行写操作 → 交给知识库
		if containsAny(q, "规则", "政策", "多久", "怎么", "如何", "能不能", "可以", "吗", "说明", "条件") {
			break
		}
		return nil, "可以帮你取消待支付的订单。请把订单号发我，或者到「我的订单」里选中那笔订单，我再帮你取消。"

	case has(tools, "order_detail") && sn != "":
		return call("order_detail", map[string]any{"sn": sn}), ""

	case has(tools, "my_orders") && containsAny(q, "我的订单", "订单列表", "查订单", "订单状态", "我订", "历史订单", "订单"):
		args := map[string]any{}
		if containsAny(q, "待支付", "未付款", "没付款") {
			args["state"] = "wait_pay"
		} else if containsAny(q, "已支付", "未使用", "待入住") {
			args["state"] = "paid"
		} else if containsAny(q, "已完成", "已使用") {
			args["state"] = "used"
		}
		return call("my_orders", args), ""

	case has(tools, "homestay_detail") && containsAny(q, "详情", "介绍", "怎么样", "好不好", "在哪", "设施", "房型"):
		if id, ok := findId(q); ok {
			return call("homestay_detail", map[string]any{"homestayId": id}), ""
		}
	}

	// 找房意图：城市 / 人数 / 价格 / 关键词
	city := findCity(q)
	people := findPeople(q)
	minP, maxP := findPriceRange(q)
	keyword := findKeyword(q)
	if has(tools, "search_homestay") && (city != "" || people > 0 || maxP > 0 || keyword != "" ||
		containsAny(q, "推荐", "找", "住", "有没有", "想住", "民宿", "房源")) {
		args := map[string]any{}
		if city != "" {
			args["city"] = city
		}
		if people > 0 {
			args["peopleNum"] = people
		}
		if minP > 0 {
			args["minPrice"] = minP
		}
		if maxP > 0 {
			args["maxPrice"] = maxP
		}
		if keyword != "" {
			args["keyword"] = keyword
		}
		return call("search_homestay", args), ""
	}

	return nil, ""
}

func (m *MockProvider) planAdmin(q string, tools map[string]bool) (*ToolCall, string) {
	sn := findSn(q)

	switch {
	case has(tools, "force_cancel_order") && containsAny(q, "强制取消", "取消订单", "退单", "取消这笔"):
		if sn == "" {
			return nil, "强制取消需要先给出订单号（例如「把订单 20240101123456789012 强制取消，理由：民宿无法接待」）。我会先生成待确认动作，你确认后才会执行。"
		}
		args := map[string]any{"sn": sn}
		args["reason"] = findReason(q)
		return call("force_cancel_order", args), ""

	case has(tools, "update_user_status") && containsAny(q, "禁用", "解禁", "封禁", "封号", "恢复"):
		id, ok := findId(q)
		if !ok {
			return nil, "请告诉我要操作的用户 ID，例如「禁用用户 3」。"
		}
		status := 0
		if containsAny(q, "解禁", "启用", "恢复") {
			status = 1
		}
		return call("update_user_status", map[string]any{"userId": id, "status": status}), ""

	case has(tools, "update_homestay_price") && containsAny(q, "改价", "调价", "改价格"):
		id, ok := findId(q)
		if !ok {
			return nil, "请告诉我要改价的房源 ID 和目标价格，例如「把房源 3 的售价改成 688」。"
		}
		price := findMoney(q)
		if price <= 0 {
			return nil, "还需要一个目标价格，例如「把房源 3 的售价改成 688」。"
		}
		return call("update_homestay_price", map[string]any{"homestayId": id, "homestayPrice": price}), ""

	case has(tools, "update_homestay_state") && containsAny(q, "下架", "上架", "停售", "恢复售卖"):
		id, ok := findId(q)
		if !ok {
			return nil, "请告诉我要操作哪套房源，例如「下架房源 3」。"
		}
		rowState := 1
		if containsAny(q, "下架", "停售") {
			rowState = 0
		}
		return call("update_homestay_state", map[string]any{"homestayId": id, "rowState": rowState}), ""

	case has(tools, "dashboard_trend") && containsAny(q, "趋势", "走势", "折线", "最近", "近几天", "近 7", "近7", "近 14", "近14", "近 30", "近30"):
		return call("dashboard_trend", map[string]any{"days": findDays(q)}), ""

	case has(tools, "dashboard_city") && containsAny(q, "城市分布", "城市", "供给"):
		return call("dashboard_city", map[string]any{}), ""

	case has(tools, "order_stat") && containsAny(q, "订单统计", "状态占比", "各状态", "订单占比", "订单状态"):
		return call("order_stat", map[string]any{}), ""

	case has(tools, "dashboard_overview") && containsAny(q, "看板", "今日", "今天", "成交额", "gmv", "GMV", "概览", "经营", "数据", "营收", "多少单", "订单量"):
		return call("dashboard_overview", map[string]any{}), ""

	case has(tools, "order_detail") && sn != "":
		return call("order_detail", map[string]any{"sn": sn}), ""

	case has(tools, "order_list") && containsAny(q, "订单列表", "查订单", "最近下单", "交易", "订单"):
		args := map[string]any{}
		if sn != "" {
			args["keyword"] = sn
		}
		if containsAny(q, "待支付") {
			args["tradeState"] = 0
		} else if containsAny(q, "已取消") {
			args["tradeState"] = -1
		}
		city := findCity(q)
		if city != "" {
			args["keyword"] = city
		}
		return call("order_list", args), ""

	case has(tools, "user_list") && containsAny(q, "用户", "会员", "客户"):
		return call("user_list", map[string]any{}), ""

	case has(tools, "homestay_list") && containsAny(q, "民宿", "房源", "列表"):
		args := map[string]any{}
		if city := findCity(q); city != "" {
			args["city"] = city
		}
		return call("homestay_list", args), ""
	}

	return nil, ""
}

// ---------- 回答 ----------

func observations(msgs []Message) []string {
	// 取最后一次用户提问之后的全部工具结果
	start := 0
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == RoleUser {
			start = i + 1
			break
		}
	}
	out := make([]string, 0, 2)
	for _, m := range msgs[start:] {
		if m.Role != RoleTool || m.Content == "" {
			continue
		}
		out = append(out, m.Content)
	}
	return out
}

func (m *MockProvider) compose(req Request, obs []string) string {
	if len(obs) == 0 {
		return "没有查到相关数据，换个条件再试试？"
	}
	var b strings.Builder
	if req.Scene == "admin" {
		b.WriteString("已为你查询到以下经营数据：\n\n")
	} else {
		b.WriteString("帮你查到了这些信息：\n\n")
	}
	for _, o := range obs {
		b.WriteString(trim(o, 1200))
		b.WriteString("\n\n")
	}
	if req.Scene == "admin" {
		b.WriteString("需要我继续下钻某条数据吗？也可以直接让我调整房源、用户或订单，我会先生成待确认动作。")
	} else {
		b.WriteString("需要我进一步筛选，或者帮你下单/查订单，随时说～")
	}
	return b.String()
}

func (m *MockProvider) answerFromKnowledge(req Request, q string) string {
	if len(req.Knowledge) == 0 {
		if req.Scene == "admin" {
			return "我暂时没有检索到对应的运营资料。可以换个问法（例如「今日成交额」「订单状态占比」「改价规范」），或者让我直接查数据。"
		}
		return "抱歉，我没找到相关信息。你可以换个说法（例如「上海适合 2 人的民宿」「取消订单规则」），或者告诉我订单号，我帮你查订单。"
	}
	var b strings.Builder
	b.WriteString(trim(req.Knowledge[0], 900))
	if len(req.Knowledge) > 1 {
		b.WriteString("\n\n补充：")
		b.WriteString(trim(req.Knowledge[1], 400))
	}
	return b.String()
}

// ---------- 流式 ----------

func (m *MockProvider) stream(text string, onDelta func(string) error) *Response {
	resp := &Response{Model: m.Name(), Content: text, Usage: Usage{PromptTokens: 40, CompletionTokens: len([]rune(text)) / 2}}
	if onDelta == nil {
		return resp
	}
	r := []rune(text)
	const chunk = 12
	for i := 0; i < len(r); i += chunk {
		end := i + chunk
		if end > len(r) {
			end = len(r)
		}
		if err := onDelta(string(r[i:end])); err != nil {
			return resp
		}
	}
	return resp
}

// ---------- 解析小工具 ----------

func call(name string, args map[string]any) *ToolCall {
	b := &strings.Builder{}
	b.WriteString("{")
	first := true
	for k, v := range args {
		if !first {
			b.WriteString(",")
		}
		first = false
		b.WriteString(strconv.Quote(k))
		b.WriteString(":")
		switch t := v.(type) {
		case string:
			b.WriteString(strconv.Quote(t))
		case int:
			b.WriteString(strconv.Itoa(t))
		case int64:
			b.WriteString(strconv.FormatInt(t, 10))
		case float64:
			b.WriteString(strconv.FormatFloat(t, 'f', -1, 64))
		default:
			b.WriteString("null")
		}
	}
	b.WriteString("}")
	return &ToolCall{Id: "call_" + name, Type: "function", Function: FunctionCall{Name: name, Arguments: b.String()}}
}

func lastMessage(msgs []Message) *Message {
	if len(msgs) == 0 {
		return nil
	}
	return &msgs[len(msgs)-1]
}

// NameOf 已不需要，保留 lastMessage 供阶段判断
func has(tools map[string]bool, name string) bool { return tools[name] }

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}

func findSn(q string) string { return snRe.FindString(q) }

func findId(q string) (int64, bool) {
	m := idRe.FindStringSubmatch(q)
	if len(m) < 2 {
		return 0, false
	}
	v, err := strconv.ParseInt(m[1], 10, 64)
	if err != nil || v <= 0 {
		return 0, false
	}
	return v, true
}

func findCity(q string) string {
	for _, c := range knownCity {
		if strings.Contains(q, c) {
			return c
		}
	}
	return ""
}

func findPeople(q string) int {
	m := regexp.MustCompile(`(\d{1,2})\s*(?:人|位|个人)`).FindStringSubmatch(q)
	if len(m) < 2 {
		return 0
	}
	v, _ := strconv.Atoi(m[1])
	if v > 0 && v <= 20 {
		return v
	}
	return 0
}

func findMoney(q string) float64 {
	m := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*(?:元|块|rmb|￥)?`).FindAllStringSubmatch(q, -1)
	best := 0.0
	for _, g := range m {
		v, err := strconv.ParseFloat(g[1], 64)
		if err == nil && v > best {
			best = v
		}
	}
	return best
}

func findPriceRange(q string) (float64, float64) {
	// 「500 以内」「不超过 800」
	if m := regexp.MustCompile(`(?:不超过|低于|少于|预算|以内|以下)?\s*(\d{2,5})\s*(?:元|块)?\s*(?:以内|以下|内)`).FindStringSubmatch(q); len(m) >= 2 {
		if v, err := strconv.ParseFloat(m[1], 64); err == nil {
			return 0, v
		}
	}
	if m := regexp.MustCompile(`(\d{2,5})\s*[-~到至]\s*(\d{2,5})`).FindStringSubmatch(q); len(m) >= 3 {
		a, _ := strconv.ParseFloat(m[1], 64)
		b, _ := strconv.ParseFloat(m[2], 64)
		return a, b
	}
	return 0, 0
}

func findKeyword(q string) string {
	tags := []string{"海景", "江景", "湖景", "独栋", "小院", "私汤", "温泉", "亲子", "设计感", "地铁", "泳池", "复式", "山景", "古镇"}
	for _, t := range tags {
		if strings.Contains(q, t) {
			return t
		}
	}
	return ""
}

func findDays(q string) int {
	m := daysRe.FindStringSubmatch(q)
	if len(m) < 2 {
		return 7
	}
	switch m[1] {
	case "7", "七":
		return 7
	case "14", "十四":
		return 14
	case "30", "三十":
		return 30
	}
	return 7
}

func findReason(q string) string {
	for _, sep := range []string{"理由", "原因", "因为"} {
		if i := strings.Index(q, sep); i >= 0 {
			r := strings.Trim(q[i+len(sep):], "：: ，,。.")
			if r != "" {
				return trim(r, 100)
			}
		}
	}
	return "运营后台由 AI 助手代为取消"
}

func trim(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
