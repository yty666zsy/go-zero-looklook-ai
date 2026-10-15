package homestay

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type HomestayUpdateStateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 上架/下架
func NewHomestayUpdateStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HomestayUpdateStateLogic {
	return &HomestayUpdateStateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HomestayUpdateStateLogic) HomestayUpdateState(req *types.HomestayUpdateStateReq) (resp *types.HomestayUpdateStateResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("id 不能为空")
	}
	if req.RowState != 0 && req.RowState != 1 {
		return nil, xerr.NewErrMsg("rowState 只能是 0（下架）或 1（上架）")
	}

	h, err := l.svcCtx.Dao.FindHomestay(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, xerr.NewErrMsg("民宿不存在")
	}
	if h.RowState == req.RowState {
		return &types.HomestayUpdateStateResp{Id: req.Id, RowState: req.RowState}, nil
	}

	if err := l.svcCtx.Dao.UpdateHomestayState(l.ctx, req.Id, req.RowState); err != nil {
		return nil, err
	}
	// C 端强缓存，改完立刻清一次，保证前台马上生效
	l.svcCtx.Dao.FlushCache(l.ctx)

	return &types.HomestayUpdateStateResp{Id: req.Id, RowState: req.RowState}, nil
}
