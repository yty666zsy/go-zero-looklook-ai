package ai

import (
	"context"

	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/ctxdata"
	"looklook/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConversationDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	scene  string
}

func NewConversationDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext, scene string) *ConversationDeleteLogic {
	return &ConversationDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		scene:  scene,
	}
}

func (l *ConversationDeleteLogic) ConversationDelete(req *types.ConversationDeleteReq) (resp *types.ConversationDeleteResp, err error) {
	if req.ConversationId <= 0 {
		return nil, xerr.NewErrMsg("会话不存在")
	}
	uid := ctxdata.GetUidFromCtx(l.ctx)
	n, err := l.svcCtx.Dao.DeleteConversation(l.ctx, req.ConversationId, l.scene, uid)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, xerr.NewErrMsg("会话不存在或无权删除")
	}
	return &types.ConversationDeleteResp{Ok: true}, nil
}
