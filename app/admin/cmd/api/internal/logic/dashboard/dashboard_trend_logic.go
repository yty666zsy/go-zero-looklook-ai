package dashboard

import (
	"context"
	"time"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardTrendLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 近 N 日下单趋势
func NewDashboardTrendLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardTrendLogic {
	return &DashboardTrendLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DashboardTrendLogic) DashboardTrend(req *types.DashboardTrendReq) (resp *types.DashboardTrendResp, err error) {
	days := req.Days
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}

	rows, err := l.svcCtx.Dao.OrderTrend(l.ctx, days)
	if err != nil {
		return nil, err
	}
	byDate := make(map[string]types.TrendPoint, len(rows))
	for _, r := range rows {
		byDate[r.Date] = types.TrendPoint{
			Date:       r.Date,
			OrderCount: r.OrderCount,
			Gmv:        conv.Round2(r.Gmv),
		}
	}

	// 补齐没有订单的日期，前端画图不用再处理空洞
	list := make([]types.TrendPoint, 0, days)
	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		day := now.AddDate(0, 0, -int(i)).Format("2006-01-02")
		if p, ok := byDate[day]; ok {
			list = append(list, p)
			continue
		}
		list = append(list, types.TrendPoint{Date: day})
	}
	return &types.DashboardTrendResp{List: list}, nil
}
