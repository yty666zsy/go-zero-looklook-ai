package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 删除知识条目（管理端）
func KnowledgeDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.KnowledgeDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewKnowledgeDeleteLogic(r.Context(), svcCtx)
		resp, err := l.KnowledgeDelete(&req)
		result.HttpResult(r, w, resp, err)
	}
}
