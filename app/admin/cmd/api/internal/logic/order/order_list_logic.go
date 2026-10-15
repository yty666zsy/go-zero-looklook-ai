package order

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type OrderListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 订单列表
func NewOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderListLogic {
	return &OrderListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderListLogic) OrderList(req *types.OrderListReq) (resp *types.OrderListResp, err error) {
	page, pageSize := conv.NormalizePage(req.Page, req.PageSize)

	rows, total, sum, err := l.svcCtx.Dao.ListOrders(l.ctx, page, pageSize, req.TradeState, req.Keyword)
	if err != nil {
		return nil, err
	}

	// 手机号在 usercenter 库，批量补一次
	userIds := make([]int64, 0, len(rows))
	for _, r := range rows {
		userIds = append(userIds, r.UserId)
	}
	mobiles, err := l.svcCtx.Dao.UserMobiles(l.ctx, userIds)
	if err != nil {
		return nil, err
	}

	list := make([]types.OrderItem, 0, len(rows))
	for _, r := range rows {
		r.UserMobile = mobiles[r.UserId]
		list = append(list, conv.OrderItem(r))
	}
	return &types.OrderListResp{
		List:     list,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
		SumPrice: sum,
	}, nil
}
