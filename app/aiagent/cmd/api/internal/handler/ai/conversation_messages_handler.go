package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 会话消息（前端用来回放历史，含工具步骤与卡片）
func ConversationMessagesHandler(svcCtx *svc.ServiceContext, scene string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConversationMessagesReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewConversationMessagesLogic(r.Context(), svcCtx, scene)
		resp, err := l.ConversationMessages(&req)
		result.HttpResult(r, w, resp, err)
	}
}
