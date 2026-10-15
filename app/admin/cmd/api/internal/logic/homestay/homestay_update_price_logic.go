package homestay

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type HomestayUpdatePriceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 调整价格
func NewHomestayUpdatePriceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomestayUpdatePriceLogic {
	return &HomestayUpdatePriceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HomestayUpdatePriceLogic) HomestayUpdatePrice(req *types.HomestayUpdatePriceReq) (resp *types.HomestayUpdatePriceResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("id 不能为空")
	}
	if req.HomestayPrice <= 0 || req.HomestayPrice > 1000000 {
		return nil, xerr.NewErrMsg("价格需在 0 ~ 1000000 元之间")
	}

	marketPrice := req.MarketHomestayPrice
	if marketPrice <= 0 {
		marketPrice = req.HomestayPrice
	}
	if marketPrice < req.HomestayPrice {
		return nil, xerr.NewErrMsg("市场价不能低于售价")
	}

	h, err := l.svcCtx.Dao.FindHomestay(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, xerr.NewErrMsg("民宿不存在")
	}

	if err := l.svcCtx.Dao.UpdateHomestayPrice(l.ctx, req.Id, req.HomestayPrice, marketPrice); err != nil {
		return nil, err
	}
	l.svcCtx.Dao.FlushCache(l.ctx)

	return &types.HomestayUpdatePriceResp{Id: req.Id}, nil
}
