package insight

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"looklook/app/aiagent/cmd/api/internal/biz"
	"looklook/app/aiagent/cmd/api/internal/dao"

	"github.com/zeromicro/go-zero/core/logx"
)

// PatrolResult 一次巡检的结果概要（前端「值守状态」与巡检报告都用它）。
type PatrolResult struct {
	Batch     string      `json:"batch"`
	StartedAt int64       `json:"startedAt"`
	CostMs    int64       `json:"costMs"`
	Scanned   Scanned     `json:"scanned"`
	Findings  int         `json:"findings"`
	Created   int         `json:"created"`
	Updated   int         `json:"updated"`
	Cleared   int64       `json:"cleared"`
	Pending   int64       `json:"pending"`
	Rules     []RuleStat  `json:"rules"`
	Alerts    []dao.Alert `json:"alerts"`
}

type Scanned struct {
	Orders    int `json:"orders"`
	Homestays int `json:"homestays"`
	Users     int `json:"users"`
}

type RuleStat struct {
	Rule  string `json:"rule"`
	Title string `json:"title"`
	Count int    `json:"count"`
}

// finding 是一条规则命中的结果（还没落库）。
type finding struct {
	Key         string
	Rule        string
	Level       string
	Category    string
	Title       string
	Detail      string
	Suggestion  string
	ActionTool  string
	ActionArgs  map[string]any
	ActionText  string
	RelatedType string
	RelatedID   string
}

// RunPatrol 跑一次异常巡检。
//
// 采集用只读接口（一次拉够，避免分页遗漏），规则是确定性的 —— 巡检不该依赖大模型，
// 否则告警会时有时无、无法解释；大模型只在日报里负责「说人话」。
func (s *Service) RunPatrol(ctx context.Context, trigger string) (*PatrolResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	started := time.Now()
	batch := started.Format("20060102-1504")
	result := &PatrolResult{Batch: batch, StartedAt: started.Unix()}

	if trigger == "" {
		trigger = "manual"
	}

	token, err := s.serviceToken()
	if err != nil {
		return nil, err
	}

	// ---------- 1) 采集 ----------
	orders, err := s.biz.AdminOrderList(ctx, token, map[string]any{"page": 1, "pageSize": 200})
	if err != nil {
		return nil, fmt.Errorf("巡检取订单数据失败：%w", err)
	}
	homestays, err := s.biz.AdminHomestayList(ctx, token, map[string]any{"page": 1, "pageSize": 200})
	if err != nil {
		return nil, fmt.Errorf("巡检取房源数据失败：%w", err)
	}
	users, err := s.biz.AdminUserList(ctx, token, map[string]any{"page": 1, "pageSize": 200})
	if err != nil {
		return nil, fmt.Errorf("巡检取用户数据失败：%w", err)
	}

	orderList := orders.List
	homestayList := homestays.List
	userList := users.List
	result.Scanned = Scanned{Orders: len(orderList), Homestays: len(homestayList), Users: len(userList)}

	// ---------- 2) 规则 ----------
	now := time.Now().Unix()
	cfg := s.cfg.Cron
	all := make([]finding, 0, 16)
	all = append(all, rulePendingPayTimeout(orderList, now, cfg.PendingPayMin, cfg.MaxAlertsPerRule)...)
	all = append(all, ruleOverdueUnused(orderList, now, cfg.MaxAlertsPerRule)...)
	all = append(all, ruleOversold(orderList, cfg.MaxAlertsPerRule)...)
	all = append(all, ruleLowScore(homestayList, cfg.LowScore, cfg.MinCommentCount, cfg.MaxAlertsPerRule)...)
	all = append(all, rulePriceAboveMarket(homestayList, cfg.MaxAlertsPerRule)...)
	all = append(all, ruleHighCancelUser(orderList, userList, cfg.HighCancelRate, cfg.MinOrderForCancel, cfg.MaxAlertsPerRule)...)
	if f := ruleIdleHomestay(homestayList, orderList, now, cfg.IdleHomestayDays); f != nil {
		all = append(all, *f)
	}
	result.Findings = len(all)

	// ---------- 3) 落库（幂等） ----------
	stat := map[string]int{}
	keys := make([]string, 0, len(all))
	for _, f := range all {
		created, err := s.dao.UpsertAlert(ctx, dao.Alert{
			AlertKey:    f.Key,
			Rule:        f.Rule,
			Level:       f.Level,
			Category:    f.Category,
			Title:       f.Title,
			Detail:      f.Detail,
			Suggestion:  f.Suggestion,
			ActionTool:  f.ActionTool,
			ActionArgs:  jsonNull(f.ActionArgs),
			ActionText:  f.ActionText,
			RelatedType: f.RelatedType,
			RelatedId:   f.RelatedID,
			Status:      dao.AlertPending,
			Batch:       batch,
		})
		if err != nil {
			logx.WithContext(ctx).Errorf("upsert alert err: %v", err)
			continue
		}
		keys = append(keys, f.Key)
		stat[f.Rule]++
		if created {
			result.Created++
		} else {
			result.Updated++
		}
	}

	// ---------- 4) 已经消失的异常自动关闭 ----------
	cleared, err := s.dao.ClearMissingAlerts(ctx, keys)
	if err != nil {
		logx.WithContext(ctx).Errorf("clear alerts err: %v", err)
	}
	result.Cleared = cleared

	// ---------- 5) 汇总 ----------
	pending, err := s.dao.CountAlerts(ctx, dao.AlertPending)
	if err != nil {
		logx.WithContext(ctx).Errorf("count alerts err: %v", err)
	}
	result.Pending = pending
	result.CostMs = time.Since(started).Milliseconds()

	rules := make([]RuleStat, 0, len(stat))
	for rule, n := range stat {
		rules = append(rules, RuleStat{Rule: rule, Title: RuleTitle(rule), Count: n})
	}
	sort.Slice(rules, func(i, j int) bool { return rules[i].Count > rules[j].Count })
	result.Rules = rules

	alerts, err := s.dao.ListAlerts(ctx, dao.AlertPending, "", 50)
	if err != nil {
		logx.WithContext(ctx).Errorf("list alerts err: %v", err)
	}
	result.Alerts = alerts

	s.lastPatrolAt = started
	s.lastPatrolInfo = *result

	logx.WithContext(ctx).Infof("patrol done batch=%s trigger=%s findings=%d created=%d cleared=%d cost=%dms",
		batch, trigger, result.Findings, result.Created, cleared, result.CostMs)
	return result, nil
}

// ---------- 规则实现 ----------

// R1 待支付超时：长时间未支付会一直占用房源库存，确认真不付款后应强制取消。
func rulePendingPayTimeout(orders []biz.AdminOrderItem, now int64, minMinutes int, limit int) []finding {
	if minMinutes <= 0 {
		minMinutes = 30
	}
	out := make([]finding, 0, 4)
	// 越早下单越紧急，先按时间排序
	sorted := append([]biz.AdminOrderItem(nil), orders...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].CreateTime < sorted[j].CreateTime })

	for _, o := range sorted {
		if o.TradeState != 0 || o.CreateTime <= 0 {
			continue
		}
		minutes := int((now - o.CreateTime) / 60)
		if minutes < minMinutes {
			continue
		}
		out = append(out, finding{
			Key:      "pending_pay_timeout:" + o.Sn,
			Rule:     "pending_pay_timeout",
			Level:    "warn",
			Category: "order",
			Title:    fmt.Sprintf("订单 %s 待支付已 %d 分钟", o.Sn, minutes),
			Detail: fmt.Sprintf("订单 %s（用户 %s，%s）下单已 %d 分钟仍未支付，金额 ¥%.2f，入住 %s ~ %s。该订单会一直占用房源库存，其他用户无法下单。",
				o.Sn, mobileOf(o), o.Title, minutes, o.OrderTotalPrice, dateText(o.LiveStartDate), dateText(o.LiveEndDate)),
			Suggestion:  "先联系用户确认是否还要预订；确认不付款后强制取消，释放房源库存。",
			ActionTool:  "force_cancel_order",
			ActionArgs:  map[string]any{"sn": o.Sn, "reason": "待支付超时，巡检建议取消"},
			ActionText:  fmt.Sprintf("强制取消订单 %s（用户 %s，¥%.2f），释放房源库存。", o.Sn, mobileOf(o), o.OrderTotalPrice),
			RelatedType: "order",
			RelatedID:   o.Sn,
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

// R2 已支付但离店超 24 小时仍未核销：民宿主漏核销，或用户根本没入住。
func ruleOverdueUnused(orders []biz.AdminOrderItem, now int64, limit int) []finding {
	out := make([]finding, 0, 4)
	for _, o := range orders {
		if o.TradeState != 1 || o.LiveEndDate <= 0 {
			continue
		}
		if now-o.LiveEndDate < 86400 {
			continue
		}
		days := int((now - o.LiveEndDate) / 86400)
		out = append(out, finding{
			Key:      "overdue_unused:" + o.Sn,
			Rule:     "overdue_unused",
			Level:    "warn",
			Category: "order",
			Title:    fmt.Sprintf("订单 %s 离店已 %d 天仍未核销", o.Sn, days),
			Detail: fmt.Sprintf("订单 %s（用户 %s，%s）入住 %s ~ %s，离店已过去 %d 天，状态仍是「未使用」。可能是民宿主忘记核销，也可能是用户未实际入住。",
				o.Sn, mobileOf(o), o.Title, dateText(o.LiveStartDate), dateText(o.LiveEndDate), days),
			Suggestion:  "联系民宿主核实：确实入住过就让其核销；未入住则与用户协商退款或按规则处理，避免结算金额与实际不符。",
			RelatedType: "order",
			RelatedID:   o.Sn,
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

// R3 房源日期超售：同一房源有两笔有效订单的入住区间重叠。
func ruleOversold(orders []biz.AdminOrderItem, limit int) []finding {
	byHomestay := map[int64][]biz.AdminOrderItem{}
	for _, o := range orders {
		// 只看还占着房源的订单：待支付 / 未使用 / 已使用
		if o.TradeState != 0 && o.TradeState != 1 && o.TradeState != 2 {
			continue
		}
		if o.LiveStartDate <= 0 || o.LiveEndDate <= 0 {
			continue
		}
		byHomestay[o.HomestayId] = append(byHomestay[o.HomestayId], o)
	}

	out := make([]finding, 0, 2)
	ids := make([]int64, 0, len(byHomestay))
	for id := range byHomestay {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	for _, id := range ids {
		list := byHomestay[id]
		sort.Slice(list, func(i, j int) bool { return list[i].LiveStartDate < list[j].LiveStartDate })
		for i := 0; i+1 < len(list); i++ {
			a, b := list[i], list[i+1]
			if a.LiveEndDate > b.LiveStartDate {
				out = append(out, finding{
					Key:      fmt.Sprintf("oversold:%d:%s:%s", id, a.Sn, b.Sn),
					Rule:     "oversold",
					Level:    "error",
					Category: "order",
					Title:    fmt.Sprintf("房源 #%d 出现日期冲突（%s 与 %s）", id, a.Sn, b.Sn),
					Detail: fmt.Sprintf("同一房源「%s」上有两笔有效订单的入住区间重叠：%s（%s ~ %s）与 %s（%s ~ %s）。若两位用户都到店将无法同时入住。",
						titleOfHomestay(a.Title), a.Sn, dateText(a.LiveStartDate), dateText(a.LiveEndDate),
						b.Sn, dateText(b.LiveStartDate), dateText(b.LiveEndDate)),
					Suggestion:  "立即人工介入：联系其中一位用户改期或退款，并检查下单链路是否有并发超卖。",
					RelatedType: "homestay",
					RelatedID:   fmt.Sprintf("%d", id),
				})
				break // 每个房源只报一条，避免组合爆炸
			}
		}
		if len(out) >= limit {
			break
		}
	}
	return out
}

// R4 在售房源长期零订单：聚合为一条，避免刷屏。
func ruleIdleHomestay(homestays []biz.AdminHomestayItem, orders []biz.AdminOrderItem, now int64, days int) *finding {
	if days <= 0 {
		days = 14
	}
	last := map[int64]int64{}
	for _, o := range orders {
		if o.TradeState == -1 { // 已取消不算有效需求
			continue
		}
		if o.CreateTime > last[o.HomestayId] {
			last[o.HomestayId] = o.CreateTime
		}
	}

	type idle struct {
		item biz.AdminHomestayItem
		last int64
	}
	list := make([]idle, 0, 8)
	for _, h := range homestays {
		if h.RowState != 1 {
			continue
		}
		t, ok := last[h.Id]
		if ok && now-t < int64(days)*86400 {
			continue
		}
		list = append(list, idle{item: h, last: t})
	}
	if len(list) == 0 {
		return nil
	}

	sort.Slice(list, func(i, j int) bool { return list[i].item.Id < list[j].item.Id })

	sb := &strings.Builder{}
	fmt.Fprintf(sb, "近 %d 天内有 %d 套在售房源没有产生任何有效订单：\n", days, len(list))
	for i, it := range list {
		if i >= 8 {
			fmt.Fprintf(sb, "…… 另有 %d 套。\n", len(list)-8)
			break
		}
		since := "从未有过订单"
		if it.last > 0 {
			since = "最后一单在 " + dateText(it.last)
		}
		fmt.Fprintf(sb, "· #%d %s（%s%s，¥%.2f/晚，评分 %.1f）：%s\n",
			it.item.Id, it.item.Title, it.item.City, it.item.Address, it.item.HomestayPrice, it.item.Score, since)
	}

	return &finding{
		Key:         "idle_homestay:agg",
		Rule:        "idle_homestay",
		Level:       "info",
		Category:    "homestay",
		Title:       fmt.Sprintf("%d 套在售房源近 %d 天零订单", len(list), days),
		Detail:      strings.TrimRight(sb.String(), "\n"),
		Suggestion:  "筛选出流量存在问题的房源：优先给这些房源做限时降价、首页曝光或标签优化；连续无改善的考虑下架，减少无效曝光。",
		RelatedType: "homestay",
	}
}

// R5 评分过低：评价样本足够且评分明显偏低，建议下架整改。
func ruleLowScore(homestays []biz.AdminHomestayItem, lowScore float64, minComments int64, limit int) []finding {
	if lowScore <= 0 {
		lowScore = 4.5
	}
	if minComments <= 0 {
		minComments = 5
	}
	out := make([]finding, 0, 4)
	for _, h := range homestays {
		if h.RowState != 1 || h.Score <= 0 {
			continue
		}
		if h.CommentCount < minComments || h.Score >= lowScore {
			continue
		}
		out = append(out, finding{
			Key:      fmt.Sprintf("low_score:%d", h.Id),
			Rule:     "low_score",
			Level:    "warn",
			Category: "homestay",
			Title:    fmt.Sprintf("房源 #%d 评分偏低（%.1f 分，%d 条评价）", h.Id, h.Score, h.CommentCount),
			Detail: fmt.Sprintf("「%s」（%s%s）当前评分 %.1f，低于 %.1f 分警戒线，已有 %d 条评价。低分房源会持续拉低列表页转化。",
				h.Title, h.City, h.Address, h.Score, lowScore, h.CommentCount),
			Suggestion:  "先下架整改（或联系民宿主改善卫生/服务后再上架），同时排查是否有集中差评可以申诉。",
			ActionTool:  "update_homestay_state",
			ActionArgs:  map[string]any{"homestayId": h.Id, "rowState": 0},
			ActionText:  fmt.Sprintf("下架房源 #%d「%s」（评分 %.1f），C 端将不再展示。", h.Id, h.Title, h.Score),
			RelatedType: "homestay",
			RelatedID:   fmt.Sprintf("%d", h.Id),
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

// R6 售价高于门市价：数据不一致，且现在的管理端会拒收这种改价。
func rulePriceAboveMarket(homestays []biz.AdminHomestayItem, limit int) []finding {
	out := make([]finding, 0, 4)
	for _, h := range homestays {
		if h.RowState != 1 || h.MarketHomestayPrice <= 0 {
			continue
		}
		if h.HomestayPrice <= h.MarketHomestayPrice {
			continue
		}
		out = append(out, finding{
			Key:      fmt.Sprintf("price_above_market:%d", h.Id),
			Rule:     "price_above_market",
			Level:    "error",
			Category: "homestay",
			Title:    fmt.Sprintf("房源 #%d 售价高于门市价", h.Id),
			Detail: fmt.Sprintf("「%s」售价 ¥%.2f 高于门市价 ¥%.2f，属于价格数据异常，容易引发用户投诉与比价质疑。",
				h.Title, h.HomestayPrice, h.MarketHomestayPrice),
			Suggestion:  "把售价改回不高于门市价的水平。",
			ActionTool:  "update_homestay_price",
			ActionArgs:  map[string]any{"homestayId": h.Id, "homestayPrice": h.MarketHomestayPrice},
			ActionText:  fmt.Sprintf("把房源 #%d「%s」售价从 ¥%.2f 调整为门市价 ¥%.2f。", h.Id, h.Title, h.HomestayPrice, h.MarketHomestayPrice),
			RelatedType: "homestay",
			RelatedID:   fmt.Sprintf("%d", h.Id),
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

// R7 高取消率用户：可能是恶意占房，也可能是体验问题需要回访。
func ruleHighCancelUser(orders []biz.AdminOrderItem, users []biz.AdminUserItem, rate float64, minOrders int64, limit int) []finding {
	if rate <= 0 {
		rate = 0.6
	}
	if minOrders <= 0 {
		minOrders = 3
	}

	type stat struct{ total, cancel int64 }
	byUser := map[int64]*stat{}
	for _, o := range orders {
		st := byUser[o.UserId]
		if st == nil {
			st = &stat{}
			byUser[o.UserId] = st
		}
		st.total++
		if o.TradeState == -1 {
			st.cancel++
		}
	}

	mobile := map[int64]string{}
	nick := map[int64]string{}
	for _, u := range users {
		mobile[u.Id] = u.Mobile
		nick[u.Id] = u.Nickname
	}

	ids := make([]int64, 0, len(byUser))
	for id := range byUser {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return byUser[ids[i]].cancel > byUser[ids[j]].cancel })

	out := make([]finding, 0, 4)
	for _, id := range ids {
		st := byUser[id]
		if st.total < minOrders {
			continue
		}
		r := float64(st.cancel) / float64(st.total)
		if r < rate {
			continue
		}
		name := mobile[id]
		if name == "" {
			name = fmt.Sprintf("用户 #%d", id)
		}
		out = append(out, finding{
			Key:      fmt.Sprintf("high_cancel_user:%d", id),
			Rule:     "high_cancel_user",
			Level:    "warn",
			Category: "user",
			Title:    fmt.Sprintf("%s 取消率 %.0f%%（%d/%d 单）", name, r*100, st.cancel, st.total),
			Detail: fmt.Sprintf("用户 #%d %s（昵称 %s）近期共 %d 笔订单，其中 %d 笔被取消，取消率 %.0f%%，明显高于正常水平。高频取消会长期占用房源库存。",
				id, mobile[id], nick[id], st.total, st.cancel, r*100),
			Suggestion:  "先回访确认原因（价格、行程变化或体验问题）；确认是恶意占房后临时禁用账号，禁用后无法登录 C 端。",
			ActionTool:  "update_user_status",
			ActionArgs:  map[string]any{"userId": id, "status": 0},
			ActionText:  fmt.Sprintf("禁用用户 #%d %s（取消率 %.0f%%），该用户将无法登录 C 端。", id, mobile[id], r*100),
			RelatedType: "user",
			RelatedID:   fmt.Sprintf("%d", id),
		})
		if len(out) >= limit {
			break
		}
	}
	return out
}

// ---------- 告警处置 ----------

// AlertAction 为一条告警生成「待确认动作」。
//
// 注意：这里同样不会直接执行写操作，只是把告警里预置的工具+参数
// 变成一次待确认动作，最终还是走 /action/confirm —— 与对话式 Agent 完全同一条安全通道。
func (s *Service) AlertAction(ctx context.Context, alertId, ownerId int64) (token, confirmText string, expireIn int64, err error) {
	alert, err := s.dao.GetAlert(ctx, alertId)
	if err != nil {
		return "", "", 0, err
	}
	if alert == nil {
		return "", "", 0, fmt.Errorf("告警不存在")
	}
	if alert.Status != dao.AlertPending {
		return "", "", 0, fmt.Errorf("该告警已处理，无需重复处置")
	}
	if alert.ActionTool == "" {
		return "", "", 0, fmt.Errorf("该告警没有可直接执行的处置动作，请人工处理")
	}
	tool, ok := s.tools.Get(dao.SceneAdmin, alert.ActionTool)
	if !ok || !tool.Write {
		return "", "", 0, fmt.Errorf("处置动作不可用：%s", alert.ActionTool)
	}

	args := map[string]any{}
	if alert.ActionArgs.Valid && alert.ActionArgs.String != "" {
		_ = json.Unmarshal([]byte(alert.ActionArgs.String), &args)
	}

	ttl := s.cfg.Agent.ActionTTLSec
	if ttl <= 0 {
		ttl = 300
	}
	token = newActionToken()
	if err := s.dao.CreateActionToken(ctx, dao.ActionToken{
		Token:      token,
		Scene:      dao.SceneAdmin,
		OwnerId:    ownerId,
		Tool:       alert.ActionTool,
		Args:       jsonNull(args),
		Summary:    alert.ActionText,
		ExpireTime: time.Now().Add(time.Duration(ttl) * time.Second),
	}); err != nil {
		return "", "", 0, err
	}
	_ = s.dao.AppendAudit(ctx, 0, dao.SceneAdmin, ownerId, alert.ActionTool, args, "来自告警 #"+fmt.Sprint(alertId)+"："+alert.Title, "pending", 0)

	return token, alert.ActionText, ttl, nil
}

// SetAlertStatus 人工把告警标记为已处置 / 已忽略 / 重新打开。
func (s *Service) SetAlertStatus(ctx context.Context, alertId, status int64) error {
	if status != dao.AlertPending && status != dao.AlertResolved && status != dao.AlertIgnored {
		return fmt.Errorf("不支持的状态：%d", status)
	}
	n, err := s.dao.SetAlertStatus(ctx, alertId, status)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("告警不存在或状态未变化")
	}
	return nil
}

// ---------- 小工具 ----------

func newActionToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func jsonNull(v any) sql.NullString {
	b, err := json.Marshal(v)
	if err != nil || string(b) == "null" || string(b) == "{}" || string(b) == "[]" {
		return sql.NullString{}
	}
	return sql.NullString{String: string(b), Valid: true}
}

func mobileOf(o biz.AdminOrderItem) string {
	if o.UserMobile == "" {
		return fmt.Sprintf("#%d", o.UserId)
	}
	return o.UserMobile
}

func titleOfHomestay(title string) string {
	if title == "" {
		return "未知房源"
	}
	return title
}

func dateText(unix int64) string {
	if unix <= 0 {
		return "-"
	}
	return time.Unix(unix, 0).Format("2006-01-02")
}

func RuleTitle(rule string) string {
	switch rule {
	case "pending_pay_timeout":
		return "待支付超时"
	case "overdue_unused":
		return "离店未核销"
	case "oversold":
		return "房源日期冲突"
	case "idle_homestay":
		return "在售房源零订单"
	case "low_score":
		return "房源评分偏低"
	case "price_above_market":
		return "售价高于门市价"
	case "high_cancel_user":
		return "用户取消率偏高"
	default:
		return rule
	}
}
