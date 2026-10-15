package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/auth"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 确认/取消一个待确认的写操作
func ActionConfirmHandler(svcCtx *svc.ServiceContext, scene string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ActionConfirmReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewActionConfirmLogic(r.Context(), svcCtx, scene, auth.Token(r))
		resp, err := l.ActionConfirm(&req)
		result.HttpResult(r, w, resp, err)
	}
}
