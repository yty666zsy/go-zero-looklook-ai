package tools

import (
	"context"
	"fmt"
	"strings"

	"looklook/app/aiagent/cmd/api/internal/biz"
)

// ---------- 前端结构化卡片 ----------

type HomestayCard struct {
	Id           int64    `json:"id"`
	Title        string   `json:"title"`
	Cover        string   `json:"cover"`
	City         string   `json:"city"`
	Address      string   `json:"address"`
	Tags         []string `json:"tags"`
	PeopleNum    int64    `json:"peopleNum"`
	Price        float64  `json:"price"`
	Score        float64  `json:"score"`
	CommentCount int64    `json:"commentCount"`
}

type OrderCard struct {
	Sn             string  `json:"sn"`
	Title          string  `json:"title"`
	Cover          string  `json:"cover"`
	TradeState     int64   `json:"tradeState"`
	TradeStateText string  `json:"tradeStateText"`
	Total          float64 `json:"total"`
	LiveStartDate  int64   `json:"liveStartDate"`
	LiveEndDate    int64   `json:"liveEndDate"`
}

type MetricItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Hint  string `json:"hint,omitempty"`
}

type TrendItem struct {
	Date       string  `json:"date"`
	OrderCount int64   `json:"orderCount"`
	Gmv        float64 `json:"gmv"`
}

// ---------- C 端工具 ----------

func registerCustomer(r *Registry, b *biz.Client) {
	r.add(&Tool{
		Name:        "search_homestay",
		Scene:       "c",
		Description: "按城市、关键词、价格区间、入住人数搜索可预订的民宿，返回房源列表。用户想找房、要推荐、提到某个城市或标签时使用。",
		Params: map[string]Param{
			"keyword":   {Type: "string", Description: "关键词，如 海景、私汤温泉、近地铁"},
			"city":      {Type: "string", Description: "城市，如 上海、杭州、丽江"},
			"minPrice":  {Type: "number", Description: "最低价（元/晚）"},
			"maxPrice":  {Type: "number", Description: "最高价（元/晚）"},
			"peopleNum": {Type: "integer", Description: "入住人数"},
			"sortBy":    {Type: "string", Description: "排序：priceAsc 价格升序 / scoreDesc 评分降序", Enum: []string{"priceAsc", "scoreDesc"}},
			"pageSize":  {Type: "integer", Description: "返回条数，默认 5，最大 10"},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			req := map[string]any{"page": 1, "pageSize": intArgOr(args, "pageSize", 5)}
			if v := strArg(args, "keyword"); v != "" {
				req["keyword"] = v
			}
			if v := strArg(args, "city"); v != "" {
				req["city"] = v
			}
			if v, ok := floatArg(args, "minPrice"); ok && v > 0 {
				req["minPrice"] = v
			}
			if v, ok := floatArg(args, "maxPrice"); ok && v > 0 {
				req["maxPrice"] = v
			}
			if v, ok := intArg(args, "peopleNum"); ok && v > 0 {
				req["peopleNum"] = v
			}
			if v := strArg(args, "sortBy"); v != "" {
				req["sortBy"] = v
			}

			list, total, err := b.SearchHomestay(ctx, env.Token, req)
			if err != nil {
				return Result{}, err
			}
			if len(list) == 0 {
				return Result{
					Observation: "没有符合条件的民宿。建议放宽价格或换个城市重试。",
					Summary:     "搜索民宿：0 条",
				}, nil
			}

			var sb strings.Builder
			fmt.Fprintf(&sb, "共找到 %d 家民宿，前 %d 家如下：\n", total, len(list))
			cards := make([]HomestayCard, 0, len(list))
			for i, h := range list {
				fmt.Fprintf(&sb, "%d) [id=%d] %s（%s）%s/晚，可住 %d 人，评分 %.1f，%d 条评价，标签：%s\n",
					i+1, h.Id, h.Title, h.City, money(h.HomestayPrice), h.PeopleNum, h.Score,
					h.CommentCount, strings.Join(h.Tags, "/"))
				cards = append(cards, HomestayCard{
					Id: h.Id, Title: h.Title, Cover: h.Cover, City: h.City, Address: h.Address,
					Tags: h.Tags, PeopleNum: h.PeopleNum, Price: h.HomestayPrice,
					Score: h.Score, CommentCount: h.CommentCount,
				})
			}

			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("搜索民宿：命中 %d 家", total),
				Cards:       []Card{{Kind: "homestay_list", Title: "为你找到的民宿", Items: cards}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "homestay_detail",
		Scene:       "c",
		Description: "查询某个民宿的详情：介绍、地址、价格、容纳人数、餐食、三维评分（卫生/位置/服务）。用户问到具体房源、或需要确认细节时使用。",
		Params: map[string]Param{
			"homestayId": {Type: "integer", Description: "民宿 ID", Required: true},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			id, ok := intArg(args, "homestayId")
			if !ok || id <= 0 {
				return Result{}, fmt.Errorf("缺少 homestayId")
			}
			d, err := b.HomestayDetail(ctx, env.Token, id)
			if err != nil {
				return Result{}, err
			}
			h := d.Homestay
			sb := &strings.Builder{}
			fmt.Fprintf(sb, "[id=%d] %s（%s）\n地址：%s %s\n价格：%s/晚（门市价 %s）\n可住 %d 人；餐食：%s（%s/位）\n标签：%s\n评分：%.1f（卫生 %.1f / 位置 %.1f / 服务 %.1f，共 %d 条评价）\n介绍：%s\n",
				h.Id, h.Title, h.City, h.City, h.Address, money(h.HomestayPrice), money(h.MarketHomestayPrice),
				h.PeopleNum, orDefault(h.FoodInfo, "无"), money(h.FoodPrice), strings.Join(h.Tags, "/"),
				d.ScoreStat.Score, d.ScoreStat.Clean, d.ScoreStat.Location, d.ScoreStat.Service,
				d.ScoreStat.Count, trimText(h.Info, 200))

			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("查询房源 #%d 详情", h.Id),
				Cards: []Card{{Kind: "homestay_list", Title: h.Title, Items: []HomestayCard{{
					Id: h.Id, Title: h.Title, Cover: h.Cover, City: h.City, Address: h.Address,
					Tags: h.Tags, PeopleNum: h.PeopleNum, Price: h.HomestayPrice,
					Score: h.Score, CommentCount: h.CommentCount,
				}}}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "my_orders",
		Scene:       "c",
		Description: "查询当前用户的订单列表。用户问「我的订单」「订单状态」「有没有待支付的订单」时使用。",
		Params: map[string]Param{
			"state":    {Type: "string", Description: "筛选：all 全部 / wait_pay 待支付 / paid 有效（未使用与已使用）", Enum: []string{"all", "wait_pay", "paid", "used"}},
			"pageSize": {Type: "integer", Description: "返回条数，默认 5"},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			tradeState := int64(-99)
			switch strArg(args, "state") {
			case "wait_pay":
				tradeState = 0
			case "paid", "used":
				tradeState = 1
			}
			list, err := b.MyOrders(ctx, env.Token, 0, intArgOr(args, "pageSize", 5), tradeState)
			if err != nil {
				return Result{}, err
			}
			if len(list) == 0 {
				return Result{Observation: "当前筛选条件下没有订单。", Summary: "查询订单：0 条"}, nil
			}

			sb := &strings.Builder{}
			fmt.Fprintf(sb, "共 %d 笔订单：\n", len(list))
			cards := make([]OrderCard, 0, len(list))
			for i, o := range list {
				fmt.Fprintf(sb, "%d) 单号 %s | %s | %s | %s | 入住 %s ~ %s\n",
					i+1, o.Sn, o.Title, stateText(o.TradeState), money(o.OrderTotalPrice),
					dateText(o.LiveStartDate), dateText(o.LiveEndDate))
				cards = append(cards, OrderCard{
					Sn: o.Sn, Title: o.Title, Cover: o.Cover,
					TradeState: o.TradeState, TradeStateText: stateText(o.TradeState),
					Total: o.OrderTotalPrice, LiveStartDate: o.LiveStartDate, LiveEndDate: o.LiveEndDate,
				})
			}
			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("查询订单：%d 条", len(list)),
				Cards:       []Card{{Kind: "order_list", Title: "我的订单", Items: cards}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "order_detail",
		Scene:       "c",
		Description: "按订单号查询订单详情（状态、金额、入住日期、人数、支付时间）。用户给出订单号时使用。",
		Params: map[string]Param{
			"sn": {Type: "string", Description: "订单号", Required: true},
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			sn := strArg(args, "sn")
			if sn == "" {
				return Result{}, fmt.Errorf("缺少订单号")
			}
			o, err := b.OrderDetail(ctx, env.Token, sn)
			if err != nil {
				return Result{}, err
			}
			pay := "未支付"
			if o.PayTime > 0 {
				pay = "已支付（" + dateTimeText(o.PayTime) + "）"
			}
			sb := &strings.Builder{}
			fmt.Fprintf(sb, "订单 %s\n房源：%s（id=%d）\n状态：%s\n金额：%s（房费 %s + 餐食 %s）\n入住：%s ~ %s，%d 人\n支付：%s\n下单时间：%s\n备注：%s\n",
				o.Sn, o.Title, o.HomestayId, stateText(o.TradeState), money(o.OrderTotalPrice),
				money(o.HomestayTotalPrice), money(o.FoodTotalPrice),
				dateText(o.LiveStartDate), dateText(o.LiveEndDate), o.LivePeopleNum,
				pay, dateTimeText(o.CreateTime), orDefault(o.Remark, "无"))

			return Result{
				Observation: sb.String(),
				Summary:     fmt.Sprintf("查询订单 %s", o.Sn),
				Cards: []Card{{Kind: "order_list", Title: "订单详情", Items: []OrderCard{{
					Sn: o.Sn, Title: o.Title, Cover: o.Cover,
					TradeState: o.TradeState, TradeStateText: stateText(o.TradeState),
					Total: o.OrderTotalPrice, LiveStartDate: o.LiveStartDate, LiveEndDate: o.LiveEndDate,
				}}}},
			}, nil
		},
	})

	r.add(&Tool{
		Name:        "cancel_order",
		Scene:       "c",
		Write:       true,
		Description: "取消一笔「待支付」的订单。属于写操作，必须由用户确认后再执行。",
		Params: map[string]Param{
			"sn": {Type: "string", Description: "订单号", Required: true},
		},
		Summarize: func(args map[string]any) string {
			return fmt.Sprintf("取消订单 %s（仅待支付订单可取消，取消后房源立即释放）", strArg(args, "sn"))
		},
		Run: func(ctx context.Context, env *Env, args map[string]any) (Result, error) {
			sn := strArg(args, "sn")
			if sn == "" {
				return Result{}, fmt.Errorf("缺少订单号")
			}
			if err := b.CancelOrder(ctx, env.Token, sn); err != nil {
				return Result{}, err
			}
			return Result{
				Observation: fmt.Sprintf("订单 %s 已取消，状态变为「已取消」，房源已释放。", sn),
				Summary:     fmt.Sprintf("取消订单 %s 成功", sn),
			}, nil
		},
	})
}
