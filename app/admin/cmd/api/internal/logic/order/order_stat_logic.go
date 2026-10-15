package order

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 订单状态统计
func NewOrderStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderStatLogic {
	return &OrderStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderStatLogic) OrderStat(req *types.OrderStatReq) (resp *types.OrderStatResp, err error) {
	rows, err := l.svcCtx.Dao.OrderStateDist(l.ctx)
	if err != nil {
		return nil, err
	}

	resp = &types.OrderStatResp{}
	for _, r := range rows {
		resp.All += r.Count
		switch r.TradeState {
		case -1:
			resp.Cancel = r.Count
		case 0:
			resp.WaitPay = r.Count
		case 1:
			resp.WaitUse = r.Count
		case 2:
			resp.Used = r.Count
		case 3:
			resp.Refund = r.Count
		case 4:
			resp.Expired = r.Count
		}
	}
	return resp, nil
}
