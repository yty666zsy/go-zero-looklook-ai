package dashboard

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardCityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 城市分布
func NewDashboardCityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardCityLogic {
	return &DashboardCityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DashboardCityLogic) DashboardCity(req *types.DashboardCityReq) (resp *types.DashboardCityResp, err error) {
	rows, err := l.svcCtx.Dao.CityStats(l.ctx)
	if err != nil {
		return nil, err
	}
	list := make([]types.CityStat, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.CityStat{
			City:  r.City,
			Count: r.Count,
			Gmv:   conv.Round2(r.Gmv),
		})
	}
	return &types.DashboardCityResp{List: list}, nil
}
