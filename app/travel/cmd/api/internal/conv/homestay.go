// Package conv 负责 model <-> api types 之间的转换。
// 这些转换在多个 logic 中复用，集中放一处避免重复。
package conv

import (
	"encoding/json"
	"strings"

	"looklook/app/travel/cmd/api/internal/types"
	"looklook/app/travel/model"
	"looklook/pkg/tool"
)

// SplitList 把逗号分隔的字符串切成切片，并去掉空白项
func SplitList(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// HomestayToType 把 model 转换成 api 层类型
func HomestayToType(m *model.Homestay) types.Homestay {
	if m == nil {
		return types.Homestay{}
	}

	images := SplitList(m.Banner)
	cover := ""
	if len(images) > 0 {
		cover = images[0]
	}

	return types.Homestay{
		Id:                  m.Id,
		Title:               m.Title,
		SubTitle:            m.SubTitle,
		Banner:              m.Banner,
		Cover:               cover,
		Images:              images,
		Info:                m.Info,
		PeopleNum:           m.PeopleNum,
		HomestayBusinessId:  m.HomestayBusinessId,
		UserId:              m.UserId,
		RowState:            m.RowState,
		RowType:             m.RowType,
		FoodInfo:            m.FoodInfo,
		FoodPrice:           tool.Fen2Yuan(m.FoodPrice),
		HomestayPrice:       tool.Fen2Yuan(m.HomestayPrice),
		MarketHomestayPrice: tool.Fen2Yuan(m.MarketHomestayPrice),
		City:                m.City,
		Address:             m.Address,
		Tags:                SplitList(m.Tags),
		Longitude:           m.Longitude,
		Latitude:            m.Latitude,
		Score:               m.Score,
		CommentCount:        m.CommentCount,
	}
}

// HomestayListToType 批量转换
func HomestayListToType(list []*model.Homestay) []types.Homestay {
	out := make([]types.Homestay, 0, len(list))
	for _, item := range list {
		out = append(out, HomestayToType(item))
	}
	return out
}

// CommentToType 评价转换，star 存的是 json 数组（卫生/位置/服务）
func CommentToType(m *model.HomestayComment) types.HomestayComment {
	if m == nil {
		return types.HomestayComment{}
	}

	var dims []float64
	if err := json.Unmarshal([]byte(m.Star), &dims); err != nil {
		dims = nil
	}

	var clean, location, service float64
	for i, v := range dims {
		switch i {
		case 0:
			clean = v
		case 1:
			location = v
		case 2:
			service = v
		}
	}
	var star float64
	if len(dims) > 0 {
		var sum float64
		for _, v := range dims {
			sum += v
		}
		star = sum / float64(len(dims))
	}

	return types.HomestayComment{
		Id:         m.Id,
		HomestayId: m.HomestayId,
		Content:    m.Content,
		Star:       star,
		Clean:      clean,
		Location:   location,
		Service:    service,
		UserId:     m.UserId,
		Nickname:   m.Nickname,
		Avatar:     m.Avatar,
		CreateTime: m.CreateTime.Unix(),
	}
}

// CommentListToType 批量转换
func CommentListToType(list []*model.HomestayComment) []types.HomestayComment {
	out := make([]types.HomestayComment, 0, len(list))
	for _, item := range list {
		out = append(out, CommentToType(item))
	}
	return out
}
