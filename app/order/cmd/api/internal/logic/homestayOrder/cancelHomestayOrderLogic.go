package homestayOrder

import (
	"context"

	"looklook/app/order/cmd/api/internal/svc"
	"looklook/app/order/cmd/api/internal/types"
	"looklook/app/order/cmd/rpc/order"
	"looklook/app/order/model"
	"looklook/pkg/ctxdata"
	"looklook/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type CancelHomestayOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelHomestayOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) CancelHomestayOrderLogic {
	return CancelHomestayOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CancelHomestayOrder 取消订单：只有本人的待支付订单可以取消
func (l *CancelHomestayOrderLogic) CancelHomestayOrder(req types.CancelHomestayOrderReq) (*types.CancelHomestayOrderResp, error) {

	userId := ctxdata.GetUidFromCtx(l.ctx)

	detail, err := l.svcCtx.OrderRpc.HomestayOrderDetail(l.ctx, &order.HomestayOrderDetailReq{
		Sn: req.Sn,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("订单不存在"), "cancel order get detail fail err : %v , sn : %s", err, req.Sn)
	}
	if detail.HomestayOrder == nil || detail.HomestayOrder.Id == 0 {
		return nil, errors.Wrapf(xerr.NewErrMsg("订单不存在"), "cancel order not found sn : %s", req.Sn)
	}
	if detail.HomestayOrder.UserId != userId {
		return nil, errors.Wrapf(xerr.NewErrMsg("无权操作该订单"), "cancel order not belong to user, sn : %s , userId : %d", req.Sn, userId)
	}
	if detail.HomestayOrder.TradeState != model.HomestayOrderTradeStateWaitPay {
		return nil, errors.Wrapf(xerr.NewErrMsg("只有待支付的订单可以取消"), "cancel order state invalid, sn : %s , state : %d", req.Sn, detail.HomestayOrder.TradeState)
	}

	_, err = l.svcCtx.OrderRpc.UpdateHomestayOrderTradeState(l.ctx, &order.UpdateHomestayOrderTradeStateReq{
		Sn:         req.Sn,
		TradeState: model.HomestayOrderTradeStateCancel,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("取消订单失败"), "cancel order update state fail err : %v , sn : %s", err, req.Sn)
	}

	return &types.CancelHomestayOrderResp{
		Sn:         req.Sn,
		TradeState: model.HomestayOrderTradeStateCancel,
	}, nil
}
