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

type ConversationMessagesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	scene  string
}

func NewConversationMessagesLogic(ctx context.Context, svcCtx *svc.ServiceContext, scene string) *ConversationMessagesLogic {
	return &ConversationMessagesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		scene:  scene,
	}
}

func (l *ConversationMessagesLogic) ConversationMessages(req *types.ConversationMessagesReq) (resp *types.ConversationMessagesResp, err error) {
	if req.ConversationId <= 0 {
		return nil, xerr.NewErrMsg("会话不存在")
	}
	uid := ctxdata.GetUidFromCtx(l.ctx)
	c, err := l.svcCtx.Dao.GetConversation(l.ctx, req.ConversationId, l.scene, uid)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, xerr.NewErrMsg("会话不存在或无权访问")
	}
	list, err := l.svcCtx.Dao.ListMessagesPage(l.ctx, req.ConversationId, l.scene, uid, req.Limit)
	if err != nil {
		return nil, err
	}
	return &types.ConversationMessagesResp{List: conv.Messages(list)}, nil
}
