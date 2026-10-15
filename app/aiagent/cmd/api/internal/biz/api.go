package biz

import "context"

// ---------- 数据结构（与现有 api 返回一一对应，价格均为「元」） ----------

type Homestay struct {
	Id                  int64    `json:"id"`
	Title               string   `json:"title"`
	SubTitle            string   `json:"subTitle"`
	Cover               string   `json:"cover"`
	Info                string   `json:"info"`
	PeopleNum           int64    `json:"peopleNum"`
	RowState            int64    `json:"rowState"`
	FoodInfo            string   `json:"foodInfo"`
	FoodPrice           float64  `json:"foodPrice"`
	HomestayPrice       float64  `json:"homestayPrice"`
	MarketHomestayPrice float64  `json:"marketHomestayPrice"`
	City                string   `json:"city"`
	Address             string   `json:"address"`
	Tags                []string `json:"tags"`
	Score               float64  `json:"score"`
	CommentCount        int64    `json:"commentCount"`
}

type ScoreStat struct {
	Score    float64 `json:"score"`
	Clean    float64 `json:"clean"`
	Location float64 `json:"location"`
	Service  float64 `json:"service"`
	Count    int64   `json:"count"`
}

type HomestayDetail struct {
	Homestay  Homestay  `json:"homestay"`
	ScoreStat ScoreStat `json:"scoreStat"`
}

type OrderView struct {
	Id              int64   `json:"id"`
	Sn              string  `json:"sn"`
	Title           string  `json:"title"`
	HomestayId      int64   `json:"homestayId"`
	Cover           string  `json:"cover"`
	OrderTotalPrice float64 `json:"orderTotalPrice"`
	CreateTime      int64   `json:"createTime"`
	TradeState      int64   `json:"tradeState"`
	LiveStartDate   int64   `json:"liveStartDate"`
	LiveEndDate     int64   `json:"liveEndDate"`
}

type OrderDetail struct {
	Sn              string  `json:"sn"`
	UserId          int64   `json:"userId"`
	HomestayId      int64   `json:"homestayId"`
	Title           string  `json:"title"`
	Cover           string  `json:"cover"`
	OrderTotalPrice float64 `json:"orderTotalPrice"`
	CreateTime      int64   `json:"createTime"`
	TradeState      int64   `json:"tradeState"`
	LiveStartDate   int64   `json:"liveStartDate"`
	LiveEndDate     int64   `json:"liveEndDate"`
	LivePeopleNum   int64   `json:"livePeopleNum"`
	NeedFood        int64   `json:"needFood"`
	FoodTotalPrice  float64 `json:"foodTotalPrice"`
	HomestayTotalPrice float64 `json:"homestayTotalPrice"`
	Remark          string  `json:"remark"`
	PayTime         int64   `json:"payTime"`
	PayType         string  `json:"payType"`
}

// ---------- 管理端数据结构 ----------

type Overview struct {
	TodayOrderCount     int64   `json:"todayOrderCount"`
	TodayGmv            float64 `json:"todayGmv"`
	YesterdayOrderCount int64   `json:"yesterdayOrderCount"`
	YesterdayGmv        float64 `json:"yesterdayGmv"`
	TotalOrderCount     int64   `json:"totalOrderCount"`
	TotalGmv            float64 `json:"totalGmv"`
	UserCount           int64   `json:"userCount"`
	HomestayCount       int64   `json:"homestayCount"`
	BusinessCount       int64   `json:"businessCount"`
	CommentCount        int64   `json:"commentCount"`
	WaitPayCount        int64   `json:"waitPayCount"`
	PaidCount           int64   `json:"paidCount"`
	CancelCount         int64   `json:"cancelCount"`
}

type TrendPoint struct {
	Date       string  `json:"date"`
	OrderCount int64   `json:"orderCount"`
	Gmv        float64 `json:"gmv"`
}

type CityStat struct {
	City  string  `json:"city"`
	Count int64   `json:"count"`
	Gmv   float64 `json:"gmv"`
}

type StateStat struct {
	TradeState int64 `json:"tradeState"`
	Count      int64 `json:"count"`
}

type OrderStat struct {
	All     int64 `json:"all"`
	WaitPay int64 `json:"waitPay"`
	WaitUse int64 `json:"waitUse"`
	Used    int64 `json:"used"`
	Cancel  int64 `json:"cancel"`
	Refund  int64 `json:"refund"`
	Expired int64 `json:"expired"`
}

type AdminOrderItem struct {
	Id              int64   `json:"id"`
	Sn              string  `json:"sn"`
	UserId          int64   `json:"userId"`
	UserMobile      string  `json:"userMobile"`
	HomestayId      int64   `json:"homestayId"`
	Title           string  `json:"title"`
	LiveStartDate   int64   `json:"liveStartDate"`
	LiveEndDate     int64   `json:"liveEndDate"`
	LivePeopleNum   int64   `json:"livePeopleNum"`
	OrderTotalPrice float64 `json:"orderTotalPrice"`
	TradeState      int64   `json:"tradeState"`
	Remark          string  `json:"remark"`
	CreateTime      int64   `json:"createTime"`
}

type AdminOrderPaged struct {
	List     []AdminOrderItem `json:"list"`
	Page     int64            `json:"page"`
	PageSize int64            `json:"pageSize"`
	Total    int64            `json:"total"`
	SumPrice float64          `json:"sumPrice"`
}

type AdminUserItem struct {
	Id         int64  `json:"id"`
	Mobile     string `json:"mobile"`
	Nickname   string `json:"nickname"`
	Sex        int64  `json:"sex"`
	Status     int64  `json:"status"`
	OrderCount int64  `json:"orderCount"`
	CreateTime int64  `json:"createTime"`
}

type AdminUserPaged struct {
	List     []AdminUserItem `json:"list"`
	Page     int64           `json:"page"`
	PageSize int64           `json:"pageSize"`
	Total    int64           `json:"total"`
}

type AdminHomestayItem struct {
	Id                  int64   `json:"id"`
	Title               string  `json:"title"`
	City                string  `json:"city"`
	Address             string  `json:"address"`
	Tags                string  `json:"tags"`
	PeopleNum           int64   `json:"peopleNum"`
	HomestayPrice       float64 `json:"homestayPrice"`
	MarketHomestayPrice float64 `json:"marketHomestayPrice"`
	RowState            int64   `json:"rowState"`
	Score               float64 `json:"score"`
	CommentCount        int64   `json:"commentCount"`
	BusinessTitle       string  `json:"businessTitle"`
	CreateTime          int64   `json:"createTime"`
}

type AdminHomestayPaged struct {
	List     []AdminHomestayItem `json:"list"`
	Page     int64               `json:"page"`
	PageSize int64               `json:"pageSize"`
	Total    int64               `json:"total"`
}

// ---------- C 端 ----------

func (c *Client) SearchHomestay(ctx context.Context, token string, req map[string]any) (list []Homestay, total int64, err error) {
	var resp struct {
		List  []Homestay `json:"list"`
		Total int64      `json:"total"`
	}
	if err = c.post(ctx, c.travelBase, "/travel/v1/homestay/search", token, req, &resp); err != nil {
		return nil, 0, err
	}
	return resp.List, resp.Total, nil
}

func (c *Client) HomestayDetail(ctx context.Context, token string, id int64) (*HomestayDetail, error) {
	var resp HomestayDetail
	if err := c.post(ctx, c.travelBase, "/travel/v1/homestay/homestayDetail", token, map[string]any{"id": id}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) MyOrders(ctx context.Context, token string, lastId, pageSize, tradeState int64) ([]OrderView, error) {
	var resp struct {
		List []OrderView `json:"list"`
	}
	req := map[string]any{"lastId": lastId, "pageSize": pageSize, "tradeState": tradeState}
	if err := c.post(ctx, c.orderBase, "/order/v1/homestayOrder/userHomestayOrderList", token, req, &resp); err != nil {
		return nil, err
	}
	return resp.List, nil
}

func (c *Client) OrderDetail(ctx context.Context, token, sn string) (*OrderDetail, error) {
	var resp OrderDetail
	if err := c.post(ctx, c.orderBase, "/order/v1/homestayOrder/userHomestayOrderDetail", token, map[string]any{"sn": sn}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) CancelOrder(ctx context.Context, token, sn string) error {
	var resp struct {
		Sn         string `json:"sn"`
		TradeState int64  `json:"tradeState"`
	}
	return c.post(ctx, c.orderBase, "/order/v1/homestayOrder/cancelHomestayOrder", token, map[string]any{"sn": sn}, &resp)
}

// ---------- 管理端 ----------

func (c *Client) AdminOverview(ctx context.Context, token string) (*Overview, error) {
	var resp Overview
	if err := c.post(ctx, c.adminBase, "/admin/v1/dashboard/overview", token, map[string]any{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) AdminTrend(ctx context.Context, token string, days int64) ([]TrendPoint, error) {
	var resp struct {
		List []TrendPoint `json:"list"`
	}
	if err := c.post(ctx, c.adminBase, "/admin/v1/dashboard/trend", token, map[string]any{"days": days}, &resp); err != nil {
		return nil, err
	}
	return resp.List, nil
}

func (c *Client) AdminCity(ctx context.Context, token string) ([]CityStat, error) {
	var resp struct {
		List []CityStat `json:"list"`
	}
	if err := c.post(ctx, c.adminBase, "/admin/v1/dashboard/city", token, map[string]any{}, &resp); err != nil {
		return nil, err
	}
	return resp.List, nil
}

func (c *Client) AdminOrderStat(ctx context.Context, token string) (*OrderStat, error) {
	var resp OrderStat
	if err := c.post(ctx, c.adminBase, "/admin/v1/order/stat", token, map[string]any{}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) AdminOrderList(ctx context.Context, token string, req map[string]any) (*AdminOrderPaged, error) {
	var resp AdminOrderPaged
	if err := c.post(ctx, c.adminBase, "/admin/v1/order/list", token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) AdminOrderDetail(ctx context.Context, token, sn string) (*AdminOrderItem, error) {
	var resp struct {
		Order AdminOrderItem `json:"order"`
	}
	if err := c.post(ctx, c.adminBase, "/admin/v1/order/detail", token, map[string]any{"sn": sn}, &resp); err != nil {
		return nil, err
	}
	return &resp.Order, nil
}

func (c *Client) AdminForceCancel(ctx context.Context, token, sn, reason string) error {
	var resp struct {
		Sn         string `json:"sn"`
		TradeState int64  `json:"tradeState"`
	}
	return c.post(ctx, c.adminBase, "/admin/v1/order/forceCancel", token, map[string]any{"sn": sn, "reason": reason}, &resp)
}

func (c *Client) AdminUserList(ctx context.Context, token string, req map[string]any) (*AdminUserPaged, error) {
	var resp AdminUserPaged
	if err := c.post(ctx, c.adminBase, "/admin/v1/user/list", token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) AdminUserUpdateStatus(ctx context.Context, token string, id, status int64) error {
	var resp struct {
		Id     int64 `json:"id"`
		Status int64 `json:"status"`
	}
	return c.post(ctx, c.adminBase, "/admin/v1/user/updateStatus", token, map[string]any{"id": id, "status": status}, &resp)
}

func (c *Client) AdminHomestayList(ctx context.Context, token string, req map[string]any) (*AdminHomestayPaged, error) {
	var resp AdminHomestayPaged
	if err := c.post(ctx, c.adminBase, "/admin/v1/homestay/list", token, req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *Client) AdminHomestayUpdateState(ctx context.Context, token string, id, rowState int64) error {
	var resp struct {
		Id       int64 `json:"id"`
		RowState int64 `json:"rowState"`
	}
	return c.post(ctx, c.adminBase, "/admin/v1/homestay/updateState", token, map[string]any{"id": id, "rowState": rowState}, &resp)
}

func (c *Client) AdminHomestayUpdatePrice(ctx context.Context, token string, id int64, price float64) error {
	var resp struct {
		Id int64 `json:"id"`
	}
	return c.post(ctx, c.adminBase, "/admin/v1/homestay/updatePrice", token, map[string]any{"id": id, "homestayPrice": price}, &resp)
}
