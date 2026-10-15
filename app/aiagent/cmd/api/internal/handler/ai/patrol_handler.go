// 异常巡检（主动型 Agent）接口：手写，非 goctl 生成。
package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 立即跑一次异常巡检（定时任务之外的「手动触发」，也方便验证规则）
func PatrolRunHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.PatrolRunReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewPatrolRunLogic(r.Context(), svcCtx)
		resp, err := l.PatrolRun(&req)
		result.HttpResult(r, w, resp, err)
	}
}
