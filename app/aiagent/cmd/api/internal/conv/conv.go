package conv

import (
	"encoding/json"

	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/insight"
	"looklook/app/aiagent/cmd/api/internal/tools"
	"looklook/app/aiagent/cmd/api/internal/types"
)

func Conversation(c dao.Conversation) types.ConversationItem {
	return types.ConversationItem{
		Id:         c.Id,
		Title:      c.Title,
		MsgCount:   c.MsgCount,
		LastMsg:    c.LastMsg,
		CreateTime: c.CreateTime.Unix(),
		UpdateTime: c.UpdateTime.Unix(),
	}
}

func Conversations(list []dao.Conversation) []types.ConversationItem {
	out := make([]types.ConversationItem, 0, len(list))
	for _, c := range list {
		out = append(out, Conversation(c))
	}
	return out
}

func Message(m dao.Message) types.MessageItem {
	item := types.MessageItem{
		Id:         m.Id,
		Role:       m.Role,
		Content:    m.Content,
		Model:      m.Model,
		CreateTime: m.CreateTime.Unix(),
	}
	if m.Steps.Valid && m.Steps.String != "" {
		var steps []types.StepItem
		if err := json.Unmarshal([]byte(m.Steps.String), &steps); err == nil {
			item.Steps = steps
		}
	}
	if m.Cards.Valid && m.Cards.String != "" {
		var cards []types.CardItem
		if err := json.Unmarshal([]byte(m.Cards.String), &cards); err == nil {
			item.Cards = cards
		}
	}
	return item
}

func Messages(list []dao.Message) []types.MessageItem {
	out := make([]types.MessageItem, 0, len(list))
	for _, m := range list {
		out = append(out, Message(m))
	}
	return out
}

func Cards(cards []tools.Card) []types.CardItem {
	out := make([]types.CardItem, 0, len(cards))
	for _, c := range cards {
		out = append(out, types.CardItem{Kind: c.Kind, Title: c.Title, Items: c.Items})
	}
	return out
}

func Knowledge(k dao.Knowledge) types.KnowledgeItem {
	return types.KnowledgeItem{
		Id:         k.Id,
		Scene:      k.Scene,
		Category:   k.Category,
		Title:      k.Title,
		Content:    k.Content,
		Keywords:   k.Keywords,
		Enabled:    k.Enabled,
		UpdateTime: k.UpdateTime.Unix(),
	}
}

// ---------- 经营日报 / 异常巡检 ----------

func Report(r dao.Report, withBody bool) types.ReportItem {
	item := types.ReportItem{
		Id:          r.Id,
		Type:        r.ReportType,
		Date:        r.ReportDate,
		Title:       r.Title,
		Summary:     r.Summary,
		GeneratedBy: r.GeneratedBy,
		Model:       r.Model,
		LatencyMs:   r.LatencyMs,
		CreateTime:  r.CreateTime.Unix(),
	}
	if r.Metrics.Valid && r.Metrics.String != "" {
		var cards []types.CardItem
		if err := json.Unmarshal([]byte(r.Metrics.String), &cards); err == nil {
			item.Metrics = cards
		}
	}
	if r.Highlights.Valid && r.Highlights.String != "" {
		var hs []string
		if err := json.Unmarshal([]byte(r.Highlights.String), &hs); err == nil {
			item.Highlights = hs
		}
	}
	if !withBody {
		// 列表页不返回正文与卡片，避免把大字段全查出来
		return item
	}
	item.Content = r.Content
	if r.Cards.Valid && r.Cards.String != "" {
		var cards []types.CardItem
		if err := json.Unmarshal([]byte(r.Cards.String), &cards); err == nil {
			item.Cards = cards
		}
	}
	if r.Steps.Valid && r.Steps.String != "" {
		var steps []types.StepItem
		if err := json.Unmarshal([]byte(r.Steps.String), &steps); err == nil {
			item.Steps = steps
		}
	}
	return item
}

func Reports(list []dao.Report) []types.ReportItem {
	out := make([]types.ReportItem, 0, len(list))
	for _, r := range list {
		out = append(out, Report(r, false))
	}
	return out
}

func Alert(a dao.Alert) types.AlertItem {
	item := types.AlertItem{
		Id:          a.Id,
		Rule:        a.Rule,
		RuleTitle:   insight.RuleTitle(a.Rule),
		Level:       a.Level,
		Category:    a.Category,
		Title:       a.Title,
		Detail:      a.Detail,
		Suggestion:  a.Suggestion,
		ActionTool:  a.ActionTool,
		ActionText:  a.ActionText,
		RelatedType: a.RelatedType,
		RelatedId:   a.RelatedId,
		Status:      a.Status,
		HitCount:    a.HitCount,
		LastHitTime: a.LastHitTime.Unix(),
		CreateTime:  a.CreateTime.Unix(),
	}
	if a.ActionArgs.Valid && a.ActionArgs.String != "" {
		var args map[string]any
		if err := json.Unmarshal([]byte(a.ActionArgs.String), &args); err == nil {
			item.ActionArgs = args
		}
	}
	return item
}

func Alerts(list []dao.Alert) []types.AlertItem {
	out := make([]types.AlertItem, 0, len(list))
	for _, a := range list {
		out = append(out, Alert(a))
	}
	return out
}

func Patrol(p *insight.PatrolResult) *types.PatrolRunResp {
	rules := make([]types.PatrolRule, 0, len(p.Rules))
	for _, r := range p.Rules {
		rules = append(rules, types.PatrolRule{Rule: r.Rule, Title: insight.RuleTitle(r.Rule), Count: r.Count})
	}
	return &types.PatrolRunResp{
		Batch:     p.Batch,
		StartedAt: p.StartedAt,
		CostMs:    p.CostMs,
		Scanned:   types.PatrolScanned{Orders: p.Scanned.Orders, Homestays: p.Scanned.Homestays, Users: p.Scanned.Users},
		Findings:  p.Findings,
		Created:   p.Created,
		Updated:   p.Updated,
		Cleared:   p.Cleared,
		Pending:   p.Pending,
		Rules:     rules,
		Alerts:    Alerts(p.Alerts),
	}
}
