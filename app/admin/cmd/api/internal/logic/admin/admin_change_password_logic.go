package admin

import (
	"context"

	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/ctxdata"
	"looklook/pkg/tool"

	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type AdminChangePasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 修改密码
func NewAdminChangePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminChangePasswordLogic {
	return &AdminChangePasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminChangePasswordLogic) AdminChangePassword(req *types.AdminChangePasswordReq) (resp *types.AdminChangePasswordResp, err error) {
	if len(req.NewPassword) < 6 {
		return nil, xerr.NewErrMsg("新密码至少 6 位")
	}

	uid := ctxdata.GetUidFromCtx(l.ctx)
	admin, err := l.svcCtx.Dao.FindAdminById(l.ctx, uid)
	if err != nil {
		return nil, err
	}
	if admin == nil {
		return nil, xerr.NewErrMsg("管理员不存在")
	}
	if admin.Password != tool.Md5ByString(req.OldPassword) {
		return nil, xerr.NewErrMsg("原密码不正确")
	}
	if err := l.svcCtx.Dao.UpdateAdminPassword(l.ctx, uid, tool.Md5ByString(req.NewPassword)); err != nil {
		return nil, err
	}
	return &types.AdminChangePasswordResp{Ok: true}, nil
}
