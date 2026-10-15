package dashboard

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 核心指标
func NewDashboardOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardOverviewLogic {
	return &DashboardOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DashboardOverviewLogic) DashboardOverview(req *types.DashboardOverviewReq) (resp *types.DashboardOverviewResp, err error) {
	o, err := l.svcCtx.Dao.DashboardOverview(l.ctx)
	if err != nil {
		return nil, err
	}
	return &types.DashboardOverviewResp{
		TodayOrderCount:     o.TodayOrderCount,
		TodayGmv:            conv.Round2(o.TodayGmv),
		YesterdayOrderCount: o.YesterdayOrderCount,
		YesterdayGmv:        conv.Round2(o.YesterdayGmv),
		TotalOrderCount:     o.TotalOrderCount,
		TotalGmv:            conv.Round2(o.TotalGmv),
		UserCount:           o.UserCount,
		HomestayCount:       o.HomestayCount,
		BusinessCount:       o.BusinessCount,
		CommentCount:        o.CommentCount,
		WaitPayCount:        o.WaitPayCount,
		PaidCount:           o.PaidCount,
		CancelCount:         o.CancelCount,
	}, nil
}
