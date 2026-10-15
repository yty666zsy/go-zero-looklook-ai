package ai

import (
	"context"

	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AiHealthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// AI 服务健康信息：当前模型提供方、知识库规模、可用工具
func NewAiHealthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AiHealthLogic {
	return &AiHealthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AiHealthLogic) AiHealth() (resp *types.AiHealthResp, err error) {
	name := l.svcCtx.LLM.Name()
	return &types.AiHealthResp{
		Provider:       name,
		Model:          l.svcCtx.Config.LLM.Model,
		Mock:           name == "mock:rule-planner",
		KnowledgeCount: l.svcCtx.KB.Size(),
		ToolsCustomer:  l.svcCtx.Tools.Names("c"),
		ToolsAdmin:     l.svcCtx.Tools.Names("admin"),
	}, nil
}
