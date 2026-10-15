package homestay

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type HomestayDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 民宿详情
func NewHomestayDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomestayDetailLogic {
	return &HomestayDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HomestayDetailLogic) HomestayDetail(req *types.HomestayDetailReq) (resp *types.HomestayDetailResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("id 不能为空")
	}
	h, err := l.svcCtx.Dao.FindHomestay(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, xerr.NewErrMsg("民宿不存在")
	}
	return &types.HomestayDetailResp{Homestay: conv.HomestayItem(*h)}, nil
}
