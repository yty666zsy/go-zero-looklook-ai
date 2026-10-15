package ai

import (
	"context"

	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowledgeDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowledgeDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeDeleteLogic {
	return &KnowledgeDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *KnowledgeDeleteLogic) KnowledgeDelete(req *types.KnowledgeDeleteReq) (resp *types.KnowledgeDeleteResp, err error) {
	if req.Id <= 0 {
		return nil, xerr.NewErrMsg("缺少知识条目 id")
	}
	n, err := l.svcCtx.Dao.DeleteKnowledge(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, xerr.NewErrMsg("知识条目不存在")
	}
	l.svcCtx.ReloadKnowledge(l.ctx)
	return &types.KnowledgeDeleteResp{Id: req.Id}, nil
}
