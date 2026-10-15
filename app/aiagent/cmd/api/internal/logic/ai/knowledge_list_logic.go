package ai

import (
	"context"

	"looklook/app/aiagent/cmd/api/internal/conv"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowledgeListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowledgeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeListLogic {
	return &KnowledgeListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KnowledgeListLogic) KnowledgeList(req *types.KnowledgeListReq) (resp *types.KnowledgeListResp, err error) {
	rows, err := l.svcCtx.Dao.AdminListKnowledge(l.ctx, req.Keyword, req.Scene)
	if err != nil {
		return nil, err
	}
	list := make([]types.KnowledgeItem, 0, len(rows))
	for _, r := range rows {
		list = append(list, conv.Knowledge(r))
	}
	return &types.KnowledgeListResp{List: list}, nil
}
