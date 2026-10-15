package insight

import (
	"context"
	"fmt"
	"strings"
	"time"

	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/llm"
	"looklook/app/aiagent/cmd/api/internal/tools"

	"github.com/zeromicro/go-zero/core/logx"
)

// 日报固定采集这几个维度（只读工具），顺序即正文顺序。
func reportPlan() []planItem {
	return []planItem{
		{tool: "dashboard_overview"},
		{tool: "dashboard_trend", args: map[string]any{"days": int64(7)}},
		{tool: "order_stat"},
		{tool: "dashboard_city"},
	}
}

// GenerateReport 生成（或重算）一份经营日报。
//
//	reportType   daily / weekly
//	date         归属日期 YYYY-MM-DD，空则取今天
//	generatedBy  cron（定时）/ manual（运营手动点）
//	force        false 时当天已有日报就直接返回，不重复跑
func (s *Service) GenerateReport(ctx context.Context, reportType, date, generatedBy string, force bool) (*dao.Report, bool, error) {
	if reportType == "" {
		reportType = dao.ReportDaily
	}
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if generatedBy == "" {
		generatedBy = "manual"
	}

	if !force {
		old, err := s.dao.GetReportByDate(ctx, reportType, date)
		if err != nil {
			return nil, false, err
		}
		if old != nil {
			return old, false, nil
		}
	}

	started := time.Now()
	token, err := s.serviceToken()
	if err != nil {
		return nil, false, err
	}

	// 1) 采集：跑一组只读工具（每一步都留审计）
	c := s.runTools(ctx, token, reportPlan())

	// 2) 汇总当前未处理的异常，日报里要有「风险提示」
	alerts, err := s.dao.ListAlerts(ctx, dao.AlertPending, "", 8)
	if err != nil {
		logx.WithContext(ctx).Errorf("list alerts err: %v", err)
	}
	pending, err := s.dao.CountAlerts(ctx, dao.AlertPending)
	if err != nil {
		logx.WithContext(ctx).Errorf("count alerts err: %v", err)
	}

	// 3) 组稿
	title, summary, content, highlights := composeReport(date, c, alerts, pending)

	// 4) 接了真模型就让它润色一版（mock 模式保持确定性文案）
	model := ""
	if s.llmPolished() {
		if polished, m := s.polish(ctx, content); polished != "" {
			content = polished
			model = m
		}
	}

	// 5) 落库
	metrics := c.firstCardWithKind("metrics")
	id, err := s.dao.UpsertReport(ctx, dao.ReportWrite{
		ReportType:  reportType,
		ReportDate:  date,
		Title:       title,
		Summary:     summary,
		Content:     content,
		Metrics:     metrics,
		Cards:       c.cards,
		Steps:       c.steps,
		Highlights:  highlights,
		Model:       model,
		GeneratedBy: generatedBy,
		LatencyMs:   time.Since(started).Milliseconds(),
	})
	if err != nil {
		return nil, false, err
	}

	s.mu.Lock()
	s.lastReportDate = date
	s.lastReportAt = time.Now()
	s.mu.Unlock()

	rep, err := s.dao.GetReport(ctx, id)
	if err != nil {
		return nil, false, err
	}
	logx.WithContext(ctx).Infof("report generated date=%s by=%s cost=%dms", date, generatedBy, time.Since(started).Milliseconds())
	return rep, true, nil
}

// LatestReportDate 最近一次成功生成的日报日期（给「值守状态」用）。
func (s *Service) LatestReportDate() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastReportDate
}

// ---------- 组稿 ----------

func composeReport(date string, c *collector, alerts []dao.Alert, pending int64) (title, summary, content string, highlights []string) {
	title = date + " 经营日报"

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "%s · 经营日报\n\n", date)

	writeSection(sb, "一、核心指标", c.obs("dashboard_overview"))
	writeSection(sb, "二、近 7 天下单趋势", c.obs("dashboard_trend"))
	writeSection(sb, "三、订单状态分布", c.obs("order_stat"))
	writeSection(sb, "四、房源城市分布", c.obs("dashboard_city"))

	// 五、异常与风险
	sb.WriteString("五、异常与风险\n")
	if pending == 0 {
		sb.WriteString("本次巡检未发现需要处理的异常。\n")
	} else {
		fmt.Fprintf(sb, "当前有 %d 条待处理异常：\n", pending)
		for _, a := range alerts {
			fmt.Fprintf(sb, "· [%s] %s —— %s\n", levelText(a.Level), a.Title, a.Suggestion)
		}
		if int64(len(alerts)) < pending {
			fmt.Fprintf(sb, "· …… 另有 %d 条，请在「异常告警」页查看。\n", pending-int64(len(alerts)))
		}
	}
	sb.WriteString("\n")

	// 六、行动建议
	sb.WriteString("六、行动建议\n")
	for i, tip := range actionTips(c, alerts, pending) {
		fmt.Fprintf(sb, "%d. %s\n", i+1, tip)
	}

	content = strings.TrimRight(sb.String(), "\n")

	// 摘要与要点：直接取指标卡片里的数值，避免二次解析
	highlights = []string{}
	for _, label := range []string{"今日订单", "今日成交额", "累计订单", "累计成交额"} {
		if v := metricText(c, "dashboard_overview", label); v != "" {
			highlights = append(highlights, label+" "+v)
		}
	}
	highlights = append(highlights, fmt.Sprintf("待处理异常 %d 条", pending))

	head := "今日经营数据已汇总"
	if len(highlights) > 0 {
		head = strings.Join(highlights[:minInt(3, len(highlights))], "，")
	}
	summary = clipText(head, 480)

	return title, summary, content, highlights
}

func writeSection(sb *strings.Builder, head, body string) {
	sb.WriteString(head + "\n")
	if strings.TrimSpace(body) == "" {
		sb.WriteString("（该项数据获取失败）\n")
	} else {
		sb.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			sb.WriteString("\n")
		}
	}
	sb.WriteString("\n")
}

// actionTips 给出可执行的建议：优先推「有告警可处置」的，其次看趋势与待支付。
func actionTips(c *collector, alerts []dao.Alert, pending int64) []string {
	tips := make([]string, 0, 4)

	actionable := 0
	for _, a := range alerts {
		if a.ActionTool != "" && a.Status == dao.AlertPending {
			actionable++
		}
	}
	if actionable > 0 {
		tips = append(tips, fmt.Sprintf("打开「异常告警」，对 %d 条可直接处置的异常点击「一键处置」（仍需二次确认后才会执行）。", actionable))
	} else if pending > 0 {
		tips = append(tips, "打开「异常告警」处理待办事项。")
	}

	if v := metricText(c, "order_stat", "待支付"); v != "" {
		tips = append(tips, "关注待支付订单 "+v+"：长时间未支付的订单会一直占用房源库存，确认不付款后应强制取消以释放库存。")
	}

	if obs := c.obs("dashboard_trend"); obs != "" {
		tips = append(tips, "结合近 7 天趋势调整投放：连续低走的日期优先补充特价房源与首页曝光。")
	}

	if len(tips) == 0 {
		tips = append(tips, "经营平稳，暂无需要立即处理的事项。")
	}
	return tips
}

// metricText 从指标卡片里取某个指标的展示文本（含环比提示）。
func metricText(c *collector, tool, label string) string {
	res, ok := c.byTool[tool]
	if !ok {
		return ""
	}
	for _, card := range res.Cards {
		items, ok := card.Items.([]tools.MetricItem)
		if !ok {
			continue
		}
		for _, it := range items {
			if it.Label != label {
				continue
			}
			if it.Hint != "" {
				return it.Value + "（" + it.Hint + "）"
			}
			return it.Value
		}
	}
	return ""
}

// polish 让大模型基于确定性数据写一版更好读的日报。
// 明确要求「只能用给定数据」，避免模型编数字。
func (s *Service) polish(ctx context.Context, draft string) (string, string) {
	sys := "你是民宿预订平台「栖居」的资深运营分析师。请把下面这份自动采集的经营数据，" +
		"改写成一份简洁的中文经营日报：开头一句话结论，然后分点说明关键指标与变化，最后给出 2-3 条可执行的运营建议。" +
		"只能使用给定数据，不要编造任何数字，不要输出除小标题外的多余格式。"

	resp, err := s.llm.ChatStream(ctx, llm.Request{
		Messages: []llm.Message{
			{Role: llm.RoleSystem, Content: sys},
			{Role: llm.RoleUser, Content: draft},
		},
		Temperature: 0.3,
		MaxTokens:   900,
		Scene:       dao.SceneAdmin,
	}, func(string) error { return nil })
	if err != nil {
		logx.WithContext(ctx).Errorf("report polish err: %v", err)
		return "", ""
	}
	text := strings.TrimSpace(resp.Content)
	if len([]rune(text)) < 120 {
		return "", ""
	}
	return text, resp.Model
}

// ---------- 小工具 ----------

func levelText(level string) string {
	switch level {
	case "error":
		return "严重"
	case "warn":
		return "警告"
	default:
		return "提示"
	}
}

func clipText(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
