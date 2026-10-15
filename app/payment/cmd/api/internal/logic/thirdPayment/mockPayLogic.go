package thirdPayment

import (
	"context"
	"fmt"
	"time"

	"looklook/app/order/cmd/rpc/order"
	"looklook/app/payment/cmd/api/internal/svc"
	"looklook/app/payment/cmd/api/internal/types"
	"looklook/app/payment/cmd/rpc/payment"
	"looklook/app/payment/model"
	"looklook/pkg/ctxdata"
	"looklook/pkg/xerr"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

// MockPayMode 模拟支付的支付方式标识
const MockPayMode = "MOCK_PAY"

type MockPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMockPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) MockPayLogic {
	return MockPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MockPay 开发环境模拟支付，完整走一遍"创建流水 -> 支付成功回调 -> 订单状态变更"的链路，
// 只是跳过真实微信支付下单环节，方便 H5/浏览器联调。线上（Mode != dev）直接拒绝。
func (l *MockPayLogic) MockPay(req types.MockPayReq) (*types.MockPayResp, error) {

	if l.svcCtx.Config.Mode != "dev" {
		return nil, errors.Wrapf(xerr.NewErrMsg("当前环境不支持模拟支付"), "mock pay not allowed in mode : %s", l.svcCtx.Config.Mode)
	}

	userId := ctxdata.GetUidFromCtx(l.ctx)

	// 1、校验订单归属与状态
	detail, err := l.svcCtx.OrderRpc.HomestayOrderDetail(l.ctx, &order.HomestayOrderDetailReq{
		Sn: req.OrderSn,
	})
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("订单不存在"), "mock pay get order detail fail err : %v , sn : %s", err, req.OrderSn)
	}
	if detail.HomestayOrder == nil || detail.HomestayOrder.Id == 0 {
		return nil, errors.Wrapf(xerr.NewErrMsg("订单不存在"), "mock pay order not found sn : %s", req.OrderSn)
	}
	if detail.HomestayOrder.UserId != userId {
		return nil, errors.Wrapf(xerr.NewErrMsg("无权支付该订单"), "mock pay order not belong to user sn : %s , userId : %d", req.OrderSn, userId)
	}
	if detail.HomestayOrder.TradeState != model.ThirdPaymentPayTradeStateWait {
		return nil, errors.Wrapf(xerr.NewErrMsg("该订单当前状态不可支付"), "mock pay order state invalid sn : %s , state : %d", req.OrderSn, detail.HomestayOrder.TradeState)
	}

	// 2、创建本地支付流水
	createResp, err := l.svcCtx.PaymentRpc.CreatePayment(l.ctx, &payment.CreatePaymentReq{
		UserId:      userId,
		PayModel:    MockPayMode,
		PayTotal:    detail.HomestayOrder.OrderTotalPrice,
		OrderSn:     req.OrderSn,
		ServiceType: model.ThirdPaymentServiceTypeHomestayOrder,
	})
	if err != nil || createResp.Sn == "" {
		return nil, errors.Wrapf(xerr.NewErrMsg("创建支付流水失败"), "mock pay create payment fail err : %v , sn : %s", err, req.OrderSn)
	}

	// 3、把流水置为支付成功（这一步内部会通过 kq 通知 order-mq 推进订单状态）
	now := time.Now().Unix()
	if _, err := l.svcCtx.PaymentRpc.UpdateTradeState(l.ctx, &payment.UpdateTradeStateReq{
		Sn:             createResp.Sn,
		TradeState:     "SUCCESS",
		TransactionId:  fmt.Sprintf("MOCK%s", createResp.Sn),
		TradeType:      "MOCK",
		TradeStateDesc: "模拟支付成功",
		PayStatus:      model.ThirdPaymentPayTradeStateSuccess,
		PayTime:        now,
	}); err != nil {
		return nil, errors.Wrapf(xerr.NewErrMsg("支付流水更新失败"), "mock pay update trade state fail err : %v , sn : %s", err, req.OrderSn)
	}

	// 4、兜底：kafka 未启动时不依赖 mq，直接推进订单状态（幂等）
	if _, err := l.svcCtx.OrderRpc.UpdateHomestayOrderTradeState(l.ctx, &order.UpdateHomestayOrderTradeStateReq{
		Sn:         req.OrderSn,
		TradeState: 1, //未使用（已支付）
	}); err != nil {
		logx.WithContext(l.ctx).Errorf("mock pay update order trade state fail err : %v , sn : %s", err, req.OrderSn)
	}

	return &types.MockPayResp{
		OrderSn:    req.OrderSn,
		TradeState: 1,
	}, nil
}
