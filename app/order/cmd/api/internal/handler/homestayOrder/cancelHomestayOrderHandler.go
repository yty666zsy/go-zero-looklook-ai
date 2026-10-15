package homestayOrder

import (
	"net/http"

	"looklook/app/order/cmd/api/internal/logic/homestayOrder"
	"looklook/app/order/cmd/api/internal/svc"
	"looklook/app/order/cmd/api/internal/types"
	"looklook/pkg/result"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CancelHomestayOrderHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CancelHomestayOrderReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := homestayOrder.NewCancelHomestayOrderLogic(r.Context(), ctx)
		resp, err := l.CancelHomestayOrder(req)
		result.HttpResult(r, w, resp, err)
	}
}
