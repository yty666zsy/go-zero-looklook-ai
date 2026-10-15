package homestayComment

import (
	"context"

	"looklook/app/travel/cmd/api/internal/conv"
	"looklook/app/travel/cmd/api/internal/svc"
	"looklook/app/travel/cmd/api/internal/types"
	"looklook/pkg/xerr"

	"github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type CommentListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentListLogic(ctx context.Context, svcCtx *svc.ServiceContext) CommentListLogic {
	return CommentListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CommentList 民宿评价分页列表
func (l *CommentListLogic) CommentList(req types.CommentListReq) (*types.CommentListResp, error) {

	if req.HomestayId <= 0 {
		return &types.CommentListResp{List: []types.HomestayComment{}}, nil
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}

	builder := l.svcCtx.HomestayCommentModel.SelectBuilder().Where(squirrel.Eq{"homestay_id": req.HomestayId})
	list, total, err := l.svcCtx.HomestayCommentModel.FindPageListByPageWithTotal(l.ctx, builder, page, pageSize, "id desc")
	if err != nil {
		return nil, errors.Wrapf(xerr.NewErrCode(xerr.DB_ERROR), "commentList db err req : %+v , err : %v", req, err)
	}

	return &types.CommentListResp{
		List:  conv.CommentListToType(list),
		Total: total,
	}, nil
}
