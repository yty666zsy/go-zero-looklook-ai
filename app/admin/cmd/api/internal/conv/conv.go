// Package conv 负责 dao 行结构 → api types 的转换。
package conv

import (
	"encoding/json"
	"math"

	"looklook/app/admin/cmd/api/internal/dao"
	"looklook/app/admin/cmd/api/internal/types"
)

func AdminInfo(a *dao.AdminUser) types.AdminInfo {
	var last int64
	if !a.LastLoginTime.IsZero() {
		if ts := a.LastLoginTime.Unix(); ts > 0 {
			last = ts
		}
	}
	return types.AdminInfo{
		Id:            a.Id,
		Username:      a.Username,
		Nickname:      a.Nickname,
		Role:          a.Role,
		LastLoginTime: last,
	}
}

func HomestayItem(r dao.HomestayRow) types.HomestayItem {
	return types.HomestayItem{
		Id:                  r.Id,
		Title:               r.Title,
		SubTitle:            r.SubTitle,
		Cover:               r.Cover,
		City:                r.City,
		Address:             r.Address,
		Tags:                r.Tags,
		PeopleNum:           r.PeopleNum,
		HomestayPrice:       round2(r.HomestayPrice),
		MarketHomestayPrice: round2(r.MarketHomestayPrice),
		FoodInfo:            r.FoodInfo,
		FoodPrice:           round2(r.FoodPrice),
		RowState:            r.RowState,
		Score:               round1(r.Score),
		CommentCount:        r.CommentCount,
		BusinessId:          r.BusinessId,
		BusinessTitle:       r.BusinessTitle,
		CreateTime:          r.CreateTime,
	}
}

func OrderItem(r dao.OrderRow) types.OrderItem {
	return types.OrderItem{
		Id:              r.Id,
		Sn:              r.Sn,
		UserId:          r.UserId,
		UserMobile:      r.UserMobile,
		HomestayId:      r.HomestayId,
		Title:           r.Title,
		SubTitle:        r.SubTitle,
		Cover:           r.Cover,
		LiveStartDate:   r.LiveStartDate,
		LiveEndDate:     r.LiveEndDate,
		LivePeopleNum:   r.LivePeopleNum,
		NeedFood:        r.NeedFood,
		HomestayPrice:   round2(r.HomestayPrice),
		FoodPrice:       round2(r.FoodPrice),
		OrderTotalPrice: round2(r.OrderTotalPrice),
		TradeState:      r.TradeState,
		TradeCode:       r.TradeCode,
		Remark:          r.Remark,
		CreateTime:      r.CreateTime,
	}
}

func UserItem(r dao.UserRow) types.UserItem {
	return types.UserItem{
		Id:         r.Id,
		Mobile:     r.Mobile,
		Nickname:   r.Nickname,
		Avatar:     r.Avatar,
		Sex:        r.Sex,
		Info:       r.Info,
		Status:     r.Status,
		OrderCount: r.OrderCount,
		CreateTime: r.CreateTime,
	}
}

// CommentItem 把 DB 里的 star（JSON 数组字符串，如 [5,4.5,5]）解析成数组，并算出综合评分。
func CommentItem(r dao.CommentRow) types.CommentItem {
	star := make([]float64, 0, 3)
	_ = json.Unmarshal([]byte(r.Star), &star)
	var score float64
	if len(star) > 0 {
		var sum float64
		for _, s := range star {
			sum += s
		}
		score = sum / float64(len(star))
	}
	return types.CommentItem{
		Id:            r.Id,
		HomestayId:    r.HomestayId,
		HomestayTitle: r.HomestayTitle,
		UserId:        r.UserId,
		Nickname:      r.Nickname,
		Avatar:        r.Avatar,
		Content:       r.Content,
		Star:          star,
		Score:         round1(score),
		CreateTime:    r.CreateTime,
	}
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// Round1 / Round2 供 logic 层做金额、评分的小数位收敛。
func Round1(v float64) float64 { return round1(v) }

func Round2(v float64) float64 { return round2(v) }

// NormalizePage 归一化分页参数并回显给前端。
func NormalizePage(pageNo, pageSize int64) (int64, int64) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return pageNo, pageSize
}

// OrderStateText 订单状态的中文描述。
func OrderStateText(state int64) string {
	switch state {
	case -1:
		return "已取消"
	case 0:
		return "待支付"
	case 1:
		return "未使用"
	case 2:
		return "已使用"
	case 3:
		return "已退款"
	case 4:
		return "已过期"
	default:
		return "未知"
	}
}
