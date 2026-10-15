package order

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type OrderDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 订单详情
func NewOrderDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderDetailLogic {
	return &OrderDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderDetailLogic) OrderDetail(req *types.OrderDetailReq) (resp *types.OrderDetailResp, err error) {
	if req.Sn == "" {
		return nil, xerr.NewErrMsg("订单号不能为空")
	}

	o, err := l.svcCtx.Dao.FindOrderBySn(l.ctx, req.Sn)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, xerr.NewErrMsg("订单不存在")
	}

	mobiles, err := l.svcCtx.Dao.UserMobiles(l.ctx, []int64{o.UserId})
	if err != nil {
		return nil, err
	}
	o.UserMobile = mobiles[o.UserId]

	return &types.OrderDetailResp{Order: conv.OrderItem(*o)}, nil
}
