package order

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/admin/cmd/api/internal/logic/order"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 订单状态统计
func OrderStatHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.OrderStatReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := order.NewOrderStatLogic(r.Context(), svcCtx)
		resp, err := l.OrderStat(&req)
		result.HttpResult(r, w, resp, err)
	}
}
