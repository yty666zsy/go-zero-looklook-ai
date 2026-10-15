package dashboard

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DashboardOrderStateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 订单状态分布
func NewDashboardOrderStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DashboardOrderStateLogic {
	return &DashboardOrderStateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DashboardOrderStateLogic) DashboardOrderState(req *types.DashboardStateReq) (resp *types.DashboardStateResp, err error) {
	rows, err := l.svcCtx.Dao.OrderStateDist(l.ctx)
	if err != nil {
		return nil, err
	}
	list := make([]types.StateStat, 0, len(rows))
	for _, r := range rows {
		list = append(list, types.StateStat{TradeState: r.TradeState, Count: r.Count})
	}
	return &types.DashboardStateResp{List: list}, nil
}
