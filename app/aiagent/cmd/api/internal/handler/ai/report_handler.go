// 经营日报（主动型 Agent）接口：手写，非 goctl 生成。
//
// 这组接口和对话式 Agent 共用同一份 AI 服务与同一套工具，
// 区别只是「谁发起」：对话由用户提问触发，日报由服务按节奏自己跑。
package ai

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/aiagent/cmd/api/internal/logic/ai"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 日报列表（管理端）
func ReportListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReportListReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewReportListLogic(r.Context(), svcCtx)
		resp, err := l.ReportList(&req)
		result.HttpResult(r, w, resp, err)
	}
}

// 日报详情（含正文、指标卡片、执行步骤）
func ReportDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReportDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewReportDetailLogic(r.Context(), svcCtx)
		resp, err := l.ReportDetail(&req)
		result.HttpResult(r, w, resp, err)
	}
}

// 手动生成日报（force=false 时当天已有日报会直接返回）
func ReportGenerateHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ReportGenerateReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := ai.NewReportGenerateLogic(r.Context(), svcCtx)
		resp, err := l.ReportGenerate(&req)
		result.HttpResult(r, w, resp, err)
	}
}
