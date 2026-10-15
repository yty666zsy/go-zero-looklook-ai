package logic

import (
	"context"
	"looklook/app/payment/cmd/rpc/internal/svc"
	"looklook/app/payment/cmd/rpc/pb"
	"looklook/app/payment/model"
	"looklook/pkg/xerr"

	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetPaymentSuccessRefundByOrderSnLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPaymentSuccessRefundByOrderSnLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPaymentSuccessRefundByOrderSnLogic {
	return &GetPaymentSuccessRefundByOrderSnLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetPaymentSuccessRefundByOrderSnLogic) GetPaymentSuccessRefundByOrderSn(in *pb.GetPaymentSuccessRefundByOrderSnReq) (*pb.GetPaymentSuccessRefundByOrderSnResp, error) {

	// 注意：`trade_state` 存的是第三方返回的字符串（如 SUCCESS），
	// 本地交易状态在 `pay_status` 字段，因此这里按 pay_status 过滤。
	whereBuilder := l.svcCtx.ThirdPaymentModel.SelectBuilder().Where(
		"order_sn = ? and pay_status in (?, ?)",
		in.OrderSn, model.ThirdPaymentPayTradeStateSuccess, model.ThirdPaymentPayTradeStateRefund,
	)
	thirdPayments, err := l.svcCtx.ThirdPaymentModel.FindAll(l.ctx, whereBuilder, "id desc")
	if err != nil && err != model.ErrNotFound {
		return nil, errors.Wrapf(xerr.NewErrMsg("get payment record fail"), "get payment record fail FindOneByQuery  err : %v , orderSn:%s", err, in.OrderSn)
	}

	var resp pb.PaymentDetail
	if len(thirdPayments) > 0 {
		thirdPayment := thirdPayments[0]
		if thirdPayment != nil {
			_ = copier.Copy(&resp, thirdPayment)
			resp.CreateTime = thirdPayment.CreateTime.Unix()
			resp.UpdateTime = thirdPayment.UpdateTime.Unix()
			resp.PayTime = thirdPayment.PayTime.Unix()
		}
	}

	return &pb.GetPaymentSuccessRefundByOrderSnResp{
		PaymentDetail: &resp,
	}, nil
}
