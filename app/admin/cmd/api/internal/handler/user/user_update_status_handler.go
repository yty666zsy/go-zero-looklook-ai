package user

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/admin/cmd/api/internal/logic/user"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 启用/禁用用户
func UserUpdateStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UserUpdateStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := user.NewUserUpdateStatusLogic(r.Context(), svcCtx)
		resp, err := l.UserUpdateStatus(&req)
		result.HttpResult(r, w, resp, err)
	}
}
