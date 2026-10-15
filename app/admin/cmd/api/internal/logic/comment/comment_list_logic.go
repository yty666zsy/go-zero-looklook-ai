package comment

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 评价列表
func NewCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentListLogic {
	return &CommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentListLogic) CommentList(req *types.CommentListReq) (resp *types.CommentListResp, err error) {
	page, pageSize := conv.NormalizePage(req.Page, req.PageSize)

	rows, total, err := l.svcCtx.Dao.ListComments(l.ctx, page, pageSize, req.HomestayId, req.Keyword)
	if err != nil {
		return nil, err
	}

	list := make([]types.CommentItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, conv.CommentItem(r))
	}
	return &types.CommentListResp{
		List:     list,
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}, nil
}
