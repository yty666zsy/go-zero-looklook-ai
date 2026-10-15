package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 删除会话
func ConversationDeleteHandler(svcCtx *svc.ServiceContext, scene string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConversationDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewConversationDeleteLogic(r.Context(), svcCtx, scene)
		resp, err := l.ConversationDelete(&req)
		result.HttpResult(r, w, resp, err)
	}
}
