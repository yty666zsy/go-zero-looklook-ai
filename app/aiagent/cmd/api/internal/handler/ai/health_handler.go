package ai

import (
	"net/http"

	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/pkg/result"
)

// AI 服务健康信息（无需鉴权，前端用来判断是否降级到规则模式）
func AiHealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := ai.NewAiHealthLogic(r.Context(), svcCtx)
		resp, err := l.AiHealth()
		result.HttpResult(r, w, resp, err)
	}
}
