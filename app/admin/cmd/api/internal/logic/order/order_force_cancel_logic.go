package order

import (
	"context"
	"fmt"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/dao"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"looklook/pkg/xerr"
)

type OrderForceCancelLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 强制取消订单
func NewOrderForceCancelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderForceCancelLogic {
	return &OrderForceCancelLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OrderForceCancelLogic) OrderForceCancel(req *types.OrderForceCancelReq) (resp *types.OrderForceCancelResp, err error) {
	if req.Sn == "" {
		return nil, xerr.NewErrMsg("订单号不能为空")
	}

	reason := req.Reason
	if reason == "" {
		reason = "后台强制取消"
	}

	affected, oldState, err := l.svcCtx.Dao.ForceCancelOrder(l.ctx, req.Sn, reason)
	if err == sqlx.ErrNotFound {
		return nil, xerr.NewErrMsg("订单不存在")
	}
	if err != nil {
		return nil, err
	}
	if affected == 0 {
		return nil, xerr.NewErrMsg(fmt.Sprintf("订单当前状态为「%s」，不能取消", conv.OrderStateText(oldState)))
	}

	l.svcCtx.Dao.FlushCache(l.ctx)
	l.Infof("admin force cancel order sn=%s, from=%d", req.Sn, oldState)

	return &types.OrderForceCancelResp{
		Sn:         req.Sn,
		TradeState: dao.OrderStateCancel,
	}, nil
}
