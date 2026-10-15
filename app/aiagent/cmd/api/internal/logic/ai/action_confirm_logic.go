package ai

import (
	"context"

	"looklook/app/aiagent/cmd/api/internal/conv"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/ctxdata"
	"looklook/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type ActionConfirmLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	scene  string
	// authToken 是调用方自己的 JWT：执行写操作时要透传给业务接口，权限模型才不会被绕过
	authToken string
}

func NewActionConfirmLogic(ctx context.Context, svcCtx *svc.ServiceContext, scene, authToken string) *ActionConfirmLogic {
	return &ActionConfirmLogic{
		Logger:    logx.WithContext(ctx),
		ctx:       ctx,
		svcCtx:    svcCtx,
		scene:     scene,
		authToken: authToken,
	}
}

// ActionConfirm 由用户在对话里点「确认执行 / 取消」后调用。
// Agent 永远不会自行执行写操作，必须经过这一步。
func (l *ActionConfirmLogic) ActionConfirm(req *types.ActionConfirmReq) (resp *types.ActionConfirmResp, err error) {
	if req.Token == "" {
		return nil, xerr.NewErrMsg("缺少确认令牌")
	}
	uid := ctxdata.GetUidFromCtx(l.ctx)
	cards, msg, err := l.svcCtx.Runner.ConfirmAction(l.ctx, l.scene, uid, l.authToken, req.Token, req.Confirm)
	if err != nil {
		return nil, xerr.NewErrMsg(err.Error())
	}
	return &types.ActionConfirmResp{
		Ok:      req.Confirm,
		Message: msg,
		Cards:   conv.Cards(cards),
	}, nil
}
