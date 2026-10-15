package ai

import (
	"context"

	"looklook/app/aiagent/cmd/api/internal/conv"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

type ConversationListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	scene  string
}

func NewConversationListLogic(ctx context.Context, svcCtx *svc.ServiceContext, scene string) *ConversationListLogic {
	return &ConversationListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		scene:  scene,
	}
}

func (l *ConversationListLogic) ConversationList(req *types.ConversationListReq) (resp *types.ConversationListResp, err error) {
	uid := ctxdata.GetUidFromCtx(l.ctx)
	list, err := l.svcCtx.Dao.ListConversations(l.ctx, l.scene, uid, req.Limit)
	if err != nil {
		return nil, err
	}
	return &types.ConversationListResp{List: conv.Conversations(list)}, nil
}
