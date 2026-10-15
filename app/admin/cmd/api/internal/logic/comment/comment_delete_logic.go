package comment

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"looklook/pkg/xerr"
)

type CommentDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 删除评价
func NewCommentDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CommentDeleteLogic {
	return &CommentDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CommentDeleteLogic) CommentDelete(req *types.CommentDeleteReq) (resp *types.CommentDeleteResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("id 不能为空")
	}

	homestayId, err := l.svcCtx.Dao.SoftDeleteComment(l.ctx, req.Id)
	if err == sqlx.ErrNotFound {
		return nil, xerr.NewErrMsg("评价不存在或已删除")
	}
	if err != nil {
		return nil, err
	}

	// 回写民宿评价数 + 清 C 端缓存
	if err := l.svcCtx.Dao.SyncHomestayCommentCount(l.ctx, []int64{homestayId}); err != nil {
		l.Errorf("sync homestay comment count err: %v", err)
	}
	l.svcCtx.Dao.FlushCache(l.ctx)

	return &types.CommentDeleteResp{Id: req.Id}, nil
}
