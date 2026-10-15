package ai

import (
	"context"
	"strings"

	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/xerr"

	"github.com/zeromicro/go-zero/core/logx"
)

type KnowledgeSaveLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewKnowledgeSaveLogic(ctx context.Context, svcCtx *svc.ServiceContext) *KnowledgeSaveLogic {
	return &KnowledgeSaveLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// KnowledgeSave 新增/更新知识库条目，保存后立即重建内存索引（热更新）。
func (l *KnowledgeSaveLogic) KnowledgeSave(req *types.KnowledgeSaveReq) (resp *types.KnowledgeSaveResp, err error) {
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	if title == "" || content == "" {
		return nil, xerr.NewErrMsg("标题和内容不能为空")
	}
	scene := req.Scene
	if scene == "" {
		scene = "all"
	}
	enabled := req.Enabled
	// 新增（id=0）时未显式传 enabled 视为启用；更新时 0 表示停用
	if req.Id == 0 && enabled == 0 {
		enabled = 1
	}
	if enabled != 0 && enabled != 1 {
		enabled = 1
	}

	id, err := l.svcCtx.Dao.UpsertKnowledge(l.ctx, dao.Knowledge{
		Id:       req.Id,
		Scene:    scene,
		Category: strings.TrimSpace(req.Category),
		Title:    title,
		Content:  content,
		Keywords: strings.TrimSpace(req.Keywords),
		Enabled:  enabled,
	})
	if err != nil {
		return nil, err
	}
	l.svcCtx.ReloadKnowledge(l.ctx)
	return &types.KnowledgeSaveResp{Id: id}, nil
}
