package admin

import (
	"context"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type AdminProfileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 当前登录管理员信息
func NewAdminProfileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminProfileLogic {
	return &AdminProfileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminProfileLogic) AdminProfile(req *types.AdminProfileReq) (resp *types.AdminProfileResp, err error) {
	admin, err := l.svcCtx.Dao.FindAdminById(l.ctx, ctxdata.GetUidFromCtx(l.ctx))
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, xerr.NewErrMsg("管理员不存在")
	}
	return &types.AdminProfileResp{Admin: conv.AdminInfo(admin)}, nil
}
