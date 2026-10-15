package tools

import (
	"context"
	"fmt"
	"math"
	"strings"

	"looklook/app/aiagent/cmd/api/internal/biz"
)

func registerAdmin(r *Registry, b *biz.Client) {
	r.add(&Tool{
		Name:        "dashboard_overview",
		Scene:       "admin",
		Description: "查询经营看板核心指标：今日/昨日订单数与成交额、累计订单与累计成交额、用户数、民宿数、评价数、待支付/已支付/已取消数量。用户问经营数据、今日成交额、GMV 时使用。",
		Params:      map[string]Param{},
		Run: func(ctx context.Context, env *Env, _ map[string]any) (Result, error) {
			o, err := b.AdminOverview(ctx, env.Token)
			if err != nil {
				return Result{}, err
			}
			sb := &strings.Builder{}
			fmt.Fprintf(sb, "今日订单 %d 单（环比 %s），今日成交额 %s（环比 %s）\n",
				o.TodayOrderCount, pct(o.TodayOrderCount, o.YesterdayOrderCount),
				money(o.TodayGmv), pctFloat(o.TodayGmv, o.YesterdayGmv))
			fmt.Fprintf(sb, "累计订单 %d 单，累计成交额 %s\n", o.TotalOrderCount, money(o.TotalGmv))
			fmt.Fprintf(sb, "其中：待支付 %d 单、已支付 %d 单、已取消 %d 单\n", o.WaitPayCount, o.PaidCount, o.CancelCount)
			fmt.Fprintf(sb, "用户 %d 位、民宿 %d 套、店铺 %d 家、评价 %d 条\n",
				o.UserCount, o.HomestayCount, o.BusinessCount, o.CommentCount)

			items := []MetricItem{
				{Label: "今日订单", Value: itoa(o.TodayOrderCount) + " 单", Hint: "环比 " + pct(o.TodayOrderCount, o.YesterdayOrderCount)},
				{Label: "今日成交额", Value: money(o.TodayGmv), Hint: "环比 " + pctFloat(o.TodayGmv, o.YesterdayGmv)},
				{Label: "累计订单", Value: itoa(o.TotalOrderCount) + " 单", Hint: "待支付 " + itoa(o.WaitPayCount) + " 单"},
				{Label: "累计成交额", Value: money(o.TotalGmv), Hint: "已支付 " + itoa(o.PaidCount) + " 单"},
				{Label: "用户 / 民宿", Value: itoa(o.UserCount) + " / " + itoa(o.HomestayCount)},
				{Label: "评价", Value: itoa(o.CommentCount) + " 条"},
			}
			return Result{
				Observation: sb.String(),
				Summary:     "查询经营看板概览",
				Cards:       []Card{{Kind: "metrics", Title: "经营概览", Items: items}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "dashboard_trend",
		Scene:       "admin",
		Description: "查询最近 N 天的下单趋势（每日订单数与成交额）。用户问趋势、走势、最近几天表现时使用。",
		Params: map[string]Param{
			"days": {Type: "integer", Description: "统计天数", Enum: []string{"7", "14", "30"}},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			days := intArgOr(args, "days", 7)
			if days != 7 && days != 14 && days != 30 {
				days = 7
			}
			list, err := b.AdminTrend(ctx, env.Token, days)
			if err != nil {
				return Result{}, err
			}
			if len(list) == 0 {
				return Result{Observation: "最近没有订单数据。", Summary: "查询趋势：无数据"}, nil
			}
			var sumOrder int64
			var sumGmv float64
			var peak TrendItem
			items := make([]TrendItem, 0, len(list))
			for _, p := range list {
				sumOrder += p.OrderCount
				sumGmv += p.Gmv
				if p.OrderCount > peak.OrderCount {
					peak = TrendItem{Date: p.Date, OrderCount: p.OrderCount, Gmv: p.Gmv}
				}
				items = append(items, TrendItem{Date: p.Date, OrderCount: p.OrderCount, Gmv: p.Gmv})
			}
			obs := fmt.Sprintf("最近 %d 天：合计 %d 单、成交额 %s，日均 %.1f 单；峰值在 %s（%d 单，%s）。",
				days, sumOrder, money(sumGmv), float64(sumOrder)/float64(len(list)),
				peak.Date, peak.OrderCount, money(peak.Gmv))
			return Result{
				Observation: obs,
				Summary:     fmt.Sprintf("查询近 %d 天下单趋势", days),
				Cards:       []Card{{Kind: "trend", Title: fmt.Sprintf("近 %d 天下单趋势", days), Items: items}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "dashboard_city",
		Scene:       "admin",
		Description: "查询房源的城市分布（各城市房源数与成交额）。用户问城市分布、哪个城市房源多时使用。",
		Params:      map[string]Param{},
		Run: func(ctx context.Context, env *Env, _ map[string]any) (Result, error) {
			list, err := b.AdminCity(ctx, env.Token)
			if err != nil {
				return Result{}, err
			}
			if len(list) == 0 {
				return Result{Observation: "暂无城市数据。", Summary: "查询城市分布：无数据"}, nil
			}
			sb := &strings.Builder{}
			fmt.Fprintf(sb, "共 %d 个城市有房源：\n", len(list))
			items := make([]MetricItem, 0, len(list))
			for _, c := range list {
				fmt.Fprintf(sb, "- %s：%d 套房源，成交额 %s\n", c.City, c.Count, money(c.Gmv))
				items = append(items, MetricItem{Label: c.City, Value: itoa(c.Count) + " 套", Hint: "成交额 " + money(c.Gmv)})
			}
			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("查询城市分布：%d 个城市", len(list)),
				Cards:       []Card{{Kind: "metrics", Title: "城市分布", Items: items}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "order_stat",
		Scene:       "admin",
		Description: "查询订单状态统计（各状态订单数量）。用户问订单统计、状态占比时使用。",
		Params:      map[string]Param{},
		Run: func(ctx context.Context, env *Env, _ map[string]any) (Result, error) {
			s, err := b.AdminOrderStat(ctx, env.Token)
			if err != nil {
				return Result{}, err
			}
			obs := fmt.Sprintf("订单合计 %d 单：待支付 %d、未使用 %d、已使用 %d、已取消 %d、已退款 %d、已过期 %d。",
				s.All, s.WaitPay, s.WaitUse, s.Used, s.Cancel, s.Refund, s.Expired)
			items := []MetricItem{
				{Label: "全部", Value: itoa(s.All) + " 单"},
				{Label: "待支付", Value: itoa(s.WaitPay) + " 单"},
				{Label: "未使用", Value: itoa(s.WaitUse) + " 单"},
				{Label: "已使用", Value: itoa(s.Used) + " 单"},
				{Label: "已取消", Value: itoa(s.Cancel) + " 单"},
				{Label: "已退款", Value: itoa(s.Refund) + " 单"},
				{Label: "已过期", Value: itoa(s.Expired) + " 单"},
			}
			return Result{
				Observation: obs,
				Summary:     "查询订单状态统计",
				Cards:       []Card{{Kind: "metrics", Title: "订单状态统计", Items: items}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "order_list",
		Scene:       "admin",
		Description: "分页查询订单列表，支持按订单号/手机号关键字和状态筛选。用户要查订单列表、最近下单、某用户的订单时使用。",
		Params: map[string]Param{
			"keyword":    {Type: "string", Description: "订单号或用户手机号"},
			"tradeState": {Type: "integer", Description: "订单状态：-1 已取消 / 0 待支付 / 1 未使用 / 2 已使用 / 3 已退款 / 4 已过期"},
			"page":       {Type: "integer", Description: "页码，默认 1"},
			"pageSize":   {Type: "integer", Description: "每页条数，默认 5，最大 10"},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			req := map[string]any{"page": intArgOr(args, "page", 1), "pageSize": intArgOr(args, "pageSize", 5)}
			if v := strArg(args, "keyword"); v != "" {
				req["keyword"] = v
			}
			if v, ok := intArg(args, "tradeState"); ok {
				req["tradeState"] = v
			}
			p, err := b.AdminOrderList(ctx, env.Token, req)
			if err != nil {
				return Result{}, err
			}
			if len(p.List) == 0 {
				return Result{Observation: "没有符合条件的订单。", Summary: "查询订单列表：0 条"}, nil
			}
			sb := &strings.Builder{}
			fmt.Fprintf(sb, "共 %d 条订单（本页 %d 条，合计金额 %s）：\n", p.Total, len(p.List), money(p.SumPrice))
			items := make([]biz.AdminOrderItem, 0, len(p.List))
			for i, o := range p.List {
				fmt.Fprintf(sb, "%d) 单号 %s | 用户 %s | %s | %s | %s | 入住 %s ~ %s\n",
					i+1, o.Sn, orDefault(o.UserMobile, itoa(o.UserId)), o.Title,
					stateText(o.TradeState), money(o.OrderTotalPrice),
					dateText(o.LiveStartDate), dateText(o.LiveEndDate))
				items = append(items, o)
			}
			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("查询订单列表：共 %d 条", p.Total),
				Cards:       []Card{{Kind: "admin_orders", Title: "订单列表", Items: items}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "order_detail",
		Scene:       "admin",
		Description: "管理端按订单号查询订单详情（含下单用户手机号与状态）。用户给出订单号要查详情时使用。",
		Params: map[string]Param{
			"sn": {Type: "string", Description: "订单号", Required: true},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			sn := strArg(args, "sn")
			if sn == "" {
				return Result{}, fmt.Errorf("缺少订单号")
			}
			o, err := b.AdminOrderDetail(ctx, env.Token, sn)
			if err != nil {
				return Result{}, err
			}
			obs := fmt.Sprintf("订单 %s：用户 %s（id=%d），房源 %s（id=%d），状态 %s，金额 %s，入住 %s ~ %s 共 %d 人，下单时间 %s，备注：%s",
				o.Sn, orDefault(o.UserMobile, "-"), o.UserId, o.Title, o.HomestayId,
				stateText(o.TradeState), money(o.OrderTotalPrice),
				dateText(o.LiveStartDate), dateText(o.LiveEndDate), o.LivePeopleNum,
				dateTimeText(o.CreateTime), orDefault(o.Remark, "无"))
			return Result{
				Observation: obs,
				Summary:     fmt.Sprintf("查询订单 %s 详情", o.Sn),
				Cards:       []Card{{Kind: "admin_orders", Title: "订单详情", Items: []biz.AdminOrderItem{*o}}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "user_list",
		Scene:       "admin",
		Description: "分页查询用户列表（含手机号、昵称、状态、下单数）。用户要查用户、会员、某个手机号时使用。",
		Params: map[string]Param{
			"keyword":  {Type: "string", Description: "手机号或昵称关键字"},
			"page":     {Type: "integer", Description: "页码，默认 1"},
			"pageSize": {Type: "integer", Description: "每页条数，默认 5，最大 10"},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			req := map[string]any{"page": intArgOr(args, "page", 1), "pageSize": intArgOr(args, "pageSize", 5)}
			if v := strArg(args, "keyword"); v != "" {
				req["keyword"] = v
			}
			p, err := b.AdminUserList(ctx, env.Token, req)
			if err != nil {
				return Result{}, err
			}
			if len(p.List) == 0 {
				return Result{Observation: "没有符合条件的用户。", Summary: "查询用户：0 条"}, nil
			}
			sb := &strings.Builder{}
			fmt.Fprintf(sb, "共 %d 位用户（本页 %d 位）：\n", p.Total, len(p.List))
			items := make([]biz.AdminUserItem, 0, len(p.List))
			for i, u := range p.List {
				status := "正常"
				if u.Status != 1 {
					status = "已禁用"
				}
				fmt.Fprintf(sb, "%d) [id=%d] %s %s（%s），下单 %d 笔，注册时间 %s\n",
					i+1, u.Id, u.Nickname, u.Mobile, status, u.OrderCount, dateText(u.CreateTime))
				items = append(items, u)
			}
			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("查询用户：共 %d 位", p.Total),
				Cards:       []Card{{Kind: "admin_users", Title: "用户列表", Items: items}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "homestay_list",
		Scene:       "admin",
		Description: "分页查询民宿/房源列表，支持城市、上架状态与关键字筛选。用户要查房源、看某城市有哪些民宿时使用。",
		Params: map[string]Param{
			"keyword":  {Type: "string", Description: "标题关键字"},
			"city":     {Type: "string", Description: "城市"},
			"rowState": {Type: "integer", Description: "0 已下架 / 1 在售"},
			"page":     {Type: "integer", Description: "页码，默认 1"},
			"pageSize": {Type: "integer", Description: "每页条数，默认 5，最大 10"},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			req := map[string]any{"page": intArgOr(args, "page", 1), "pageSize": intArgOr(args, "pageSize", 5)}
			if v := strArg(args, "keyword"); v != "" {
				req["keyword"] = v
			}
			if v := strArg(args, "city"); v != "" {
				req["city"] = v
			}
			if v, ok := intArg(args, "rowState"); ok {
				req["rowState"] = v
			}
			p, err := b.AdminHomestayList(ctx, env.Token, req)
			if err != nil {
				return Result{}, err
			}
			if len(p.List) == 0 {
				return Result{Observation: "没有符合条件的房源。", Summary: "查询房源：0 条"}, nil
			}
			sb := &strings.Builder{}
			fmt.Fprintf(sb, "共 %d 套房源（本页 %d 套）：\n", p.Total, len(p.List))
			items := make([]biz.AdminHomestayItem, 0, len(p.List))
			for i, h := range p.List {
				state := "在售"
				if h.RowState == 0 {
					state = "已下架"
				}
				fmt.Fprintf(sb, "%d) [id=%d] %s（%s %s）%s/晚，门市价 %s，可住 %d 人，%s，评分 %.1f\n",
					i+1, h.Id, h.Title, h.City, h.Address, money(h.HomestayPrice),
					money(h.MarketHomestayPrice), h.PeopleNum, state, h.Score)
				items = append(items, h)
			}
			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("查询房源：共 %d 套", p.Total),
				Cards:       []Card{{Kind: "admin_homestays", Title: "房源列表", Items: items}},
			}, nil
		},
	})

	// ---------- 写操作（需二次确认） ----------

	r.add(&Tool{
		Name:        "force_cancel_order",
		Scene:       "admin",
		Write:       true,
		Description: "强制取消订单。已支付订单会同步把支付流水置为已退款并回推 C 端状态。属于高风险写操作，必须由管理员确认。",
		Params: map[string]Param{
			"sn":     {Type: "string", Description: "订单号", Required: true},
			"reason": {Type: "string", Description: "取消原因（会记录到审计）"},
		},
		Summarize: func(args map[string]any) string {
			return fmt.Sprintf("强制取消订单 %s（原因：%s）。已支付订单将同步退款并通知 C 端，操作不可逆。",
				strArg(args, "sn"), orDefault(strArg(args, "reason"), "未填写"))
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			sn := strArg(args, "sn")
			if sn == "" {
				return Result{}, fmt.Errorf("缺少订单号")
			}
			reason := orDefault(strArg(args, "reason"), "运营后台强制取消")
			if err := b.AdminForceCancel(ctx, env.Token, sn, reason); err != nil {
				return Result{}, err
			}
			return Result{
				Observation: fmt.Sprintf("订单 %s 已强制取消，状态已同步到 C 端。", sn),
				Summary:     fmt.Sprintf("强制取消订单 %s 成功", sn),
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "update_user_status",
		Scene:       "admin",
		Write:       true,
		Description: "禁用或启用用户。禁用后该用户无法登录 C 端。属于写操作，必须由管理员确认。",
		Params: map[string]Param{
			"userId": {Type: "integer", Description: "用户 ID", Required: true},
			"status": {Type: "integer", Description: "1 启用 / 0 禁用", Required: true, Enum: []string{"0", "1"}},
		},
		Summarize: func(args map[string]any) string {
			id, _ := intArg(args, "userId")
			s, _ := intArg(args, "status")
			if s == 1 {
				return fmt.Sprintf("启用用户 #%d，该用户将恢复 C 端登录。", id)
			}
			return fmt.Sprintf("禁用用户 #%d，该用户将无法登录 C 端。", id)
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			id, ok := intArg(args, "userId")
			if !ok || id <= 0 {
				return Result{}, fmt.Errorf("缺少 userId")
			}
			s, ok := intArg(args, "status")
			if !ok {
				return Result{}, fmt.Errorf("缺少 status")
			}
			if err := b.AdminUserUpdateStatus(ctx, env.Token, id, s); err != nil {
				return Result{}, err
			}
			word := "禁用"
			if s == 1 {
				word = "启用"
			}
			return Result{
				Observation: fmt.Sprintf("用户 #%d 已%s。", id, word),
				Summary:     fmt.Sprintf("%s用户 #%d 成功", word, id),
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "update_homestay_state",
		Scene:       "admin",
		Write:       true,
		Description: "上架或下架民宿。下架后 C 端搜索与首页不再展示。属于写操作，必须由管理员确认。",
		Params: map[string]Param{
			"homestayId": {Type: "integer", Description: "民宿 ID", Required: true},
			"rowState":   {Type: "integer", Description: "1 上架 / 0 下架", Required: true, Enum: []string{"0", "1"}},
		},
		Summarize: func(args map[string]any) string {
			id, _ := intArg(args, "homestayId")
			s, _ := intArg(args, "rowState")
			if s == 1 {
				return fmt.Sprintf("上架房源 #%d，C 端将恢复可预订。", id)
			}
			return fmt.Sprintf("下架房源 #%d，C 端将不再展示该房源（历史订单不受影响）。", id)
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			id, ok := intArg(args, "homestayId")
			if !ok || id <= 0 {
				return Result{}, fmt.Errorf("缺少 homestayId")
			}
			s, ok := intArg(args, "rowState")
			if !ok {
				return Result{}, fmt.Errorf("缺少 rowState")
			}
			if err := b.AdminHomestayUpdateState(ctx, env.Token, id, s); err != nil {
				return Result{}, err
			}
			word := "下架"
			if s == 1 {
				word = "上架"
			}
			return Result{
				Observation: fmt.Sprintf("房源 #%d 已%s，C 端已同步。", id, word),
				Summary:     fmt.Sprintf("%s房源 #%d 成功", word, id),
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "update_homestay_price",
		Scene:       "admin",
		Write:       true,
		Description: "修改民宿售价。售价不能高于门市价（市场价），否则会被拒绝。属于写操作，必须由管理员确认。",
		Params: map[string]Param{
			"homestayId":    {Type: "integer", Description: "民宿 ID", Required: true},
			"homestayPrice": {Type: "number", Description: "目标售价（元/晚）", Required: true},
		},
		Summarize: func(args map[string]any) string {
			id, _ := intArg(args, "homestayId")
			p, _ := floatArg(args, "homestayPrice")
			return fmt.Sprintf("把房源 #%d 的售价改为 %s/晚（若高于门市价会被系统拒绝）。", id, money(p))
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			id, ok := intArg(args, "homestayId")
			if !ok || id <= 0 {
				return Result{}, fmt.Errorf("缺少 homestayId")
			}
			price, ok := floatArg(args, "homestayPrice")
			if !ok || price <= 0 {
				return Result{}, fmt.Errorf("缺少 homestayPrice")
			}
			if err := b.AdminHomestayUpdatePrice(ctx, env.Token, id, price); err != nil {
				return Result{}, err
			}
			return Result{
				Observation: fmt.Sprintf("房源 #%d 的售价已改为 %s/晚。", id, money(price)),
				Summary:     fmt.Sprintf("房源 #%d 改价成功", id),
			}, nil
		},
	})
}

// ---------- 环比 ----------

func pct(cur, prev int64) string { return pctFloat(float64(cur), float64(prev)) }

func pctFloat(cur, prev float64) string {
	if prev == 0 {
		if cur == 0 {
			return "0.0%"
		}
		return "+100.0%"
	}
	v := (cur - prev) / math.Abs(prev) * 100
	sign := ""
	if v > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.1f%%", sign, v)
}
