package homestay

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HomestayListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 民宿列表
func NewHomestayListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomestayListLogic {
	return &HomestayListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HomestayListLogic) HomestayList(req *types.HomestayListReq) (resp *types.HomestayListResp, err error) {
	page, pageSize := conv.NormalizePage(req.Page, req.PageSize)

	rows, total, err := l.svcCtx.Dao.ListHomestays(l.ctx, page, pageSize, req.Keyword, req.City, req.RowState)
	if err != nil {
		return nil, err
	}

	list := make([]types.HomestayItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, conv.HomestayItem(r))
	}
	return &types.HomestayListResp{
		List:     list,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
