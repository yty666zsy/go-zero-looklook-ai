package homestay

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/admin/cmd/api/internal/logic/homestay"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 调整价格
func HomestayUpdatePriceHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.HomestayUpdatePriceReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := homestay.NewHomestayUpdatePriceLogic(r.Context(), svcCtx)
		resp, err := l.HomestayUpdatePrice(&req)
		result.HttpResult(r, w, resp, err)
	}
}
