package comment

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"looklook/app/admin/cmd/api/internal/logic/comment"
	"looklook/app/admin/cmd/api/internal/svc"
	"looklook/app/admin/cmd/api/internal/types"
	"looklook/pkg/result"
)

// 删除评价
func CommentDeleteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.CommentDeleteReq
		if err := httpx.Parse(r, &req); err != nil {
			result.ParamErrorResult(r, w, err)
			return
		}

		l := comment.NewCommentDeleteLogic(r.Context(), svcCtx)
		resp, err := l.CommentDelete(&req)
		result.HttpResult(r, w, resp, err)
	}
}
