// 异常告警接口：手写，非 goctl 生成。
//
// 一键处置返回的是一次性令牌，执行仍走 /ai/v1/admin/action/confirm，
// 也就是和对话式 Agent 完全同一条「人工确认后才落地」的通道。
package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 告警列表（status=-1 表示全部）
func AlertListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AlertListReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewAlertListLogic(r.Context(), svcCtx)
		resp, err := l.AlertList(&req)
		result.HttpResult(r, w, resp, err)
	}
}

// 生成「一键处置」的待确认动作
func AlertActionHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AlertActionReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewAlertActionLogic(r.Context(), svcCtx)
		resp, err := l.AlertAction(&req)
		result.HttpResult(r, w, resp, err)
	}
}

// 标记告警：已处置 / 已忽略 / 重新打开
func AlertStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AlertStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewAlertStatusLogic(r.Context(), svcCtx)
		resp, err := l.AlertStatus(&req)
		result.HttpResult(r, w, resp, err)
	}
}
