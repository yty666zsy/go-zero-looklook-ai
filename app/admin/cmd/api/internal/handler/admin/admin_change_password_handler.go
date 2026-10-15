package admin

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/admin/cmd/api/internal/logic/admin"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 修改密码
func AdminChangePasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.AdminChangePasswordReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := admin.NewAdminChangePasswordLogic(r.Context(), svcCtx)
		resp, err := l.AdminChangePassword(&req)
		result.HttpResult(r, w, resp, err)
	}
}
