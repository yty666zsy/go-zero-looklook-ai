package auth

import (
	"context"
	"time"

	"looklook/app/admin/cmd/api/internal/conv"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/ctxdata"
	"looklook/pkg/tool"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"looklook/pkg/xerr"
)

type AdminLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// 管理员登录
func NewAdminLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminLoginLogic {
	return &AdminLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminLoginLogic) AdminLogin(req *types.AdminLoginReq) (resp *types.AdminLoginResp, err error) {
	if req.Username == "" || req.Password == "" {
		return nil, xerr.NewErrMsg("用户名和密码不能为空")
	}

	admin, err := l.svcCtx.Dao.FindAdminByUsername(l.ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if admin == nil || admin.Password != tool.Md5ByString(req.Password) {
		return nil, xerr.NewErrMsg("用户名或密码错误")
	}
	if admin.Status != 1 {
		return nil, xerr.NewErrMsg("该账号已被禁用")
	}

	now := time.Now().Unix()
	accessExpire := l.svcCtx.Config.JwtAuth.AccessExpire
	token, err := l.jwtToken(l.svcCtx.Config.JwtAuth.AccessSecret, now, accessExpire, admin.Id)
	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.Dao.TouchAdminLoginTime(l.ctx, admin.Id, time.Now()); err != nil {
		l.Errorf("touch admin login time err: %v", err)
	}

	return &types.AdminLoginResp{
		AccessToken:  token,
		AccessExpire: now + accessExpire,
		RefreshAfter: now + accessExpire/2,
		Admin:        conv.AdminInfo(admin),
	}, nil
}

func (l *AdminLoginLogic) jwtToken(secretKey string, iat, seconds, adminId int64) (string, error) {
	claims := make(jwt.MapClaims)
	claims["exp"] = iat + seconds
	claims["iat"] = iat
	claims[ctxdata.CtxKeyJwtUserId] = adminId
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims = claims
	return token.SignedString([]byte(secretKey))
}
