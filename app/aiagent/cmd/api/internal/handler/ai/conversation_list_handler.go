package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 会话列表
func ConversationListHandler(svcCtx *svc.ServiceContext, scene string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConversationListReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewConversationListLogic(r.Context(), svcCtx, scene)
		resp, err := l.ConversationList(&req)
		result.HttpResult(r, w, resp, err)
	}
}
