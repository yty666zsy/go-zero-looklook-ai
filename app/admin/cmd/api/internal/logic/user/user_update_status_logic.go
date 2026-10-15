package user

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type UserUpdateStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 启用/禁用用户
func NewUserUpdateStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UserUpdateStatusLogic {
	return &UserUpdateStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UserUpdateStatusLogic) UserUpdateStatus(req *types.UserUpdateStatusReq) (resp *types.UserUpdateStatusResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("id 不能为空")
	}
	if req.Status != 0 && req.Status != 1 {
		return nil, xerr.NewErrMsg("status 只能是 0（禁用）或 1（正常）")
	}

	if err := l.svcCtx.Dao.UpdateUserStatus(l.ctx, req.Id, req.Status); err != nil {
		return nil, err
	}
	// user 在 C 端有缓存（cache:looklookUsercenter:user:*）
	l.svcCtx.Dao.FlushCache(l.ctx)

	return &types.UserUpdateStatusResp{Id: req.Id, Status: req.Status}, nil
}
