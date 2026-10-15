package ai

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"looklook/app/aiagent/cmd/api/internal/agent"
	"looklook/app/aiagent/cmd/api/internal/auth"
	"looklook/app/aiagent/cmd/api/internal/dao"
	"looklook/app/aiagent/cmd/api/internal/svc"
	"looklook/app/aiagent/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// ChatStreamHandler 是手写的 SSE 接口（不走 .api 生成，因为需要 Flusher）。
//
// 协议：每个事件一行 `data: {json}`，JSON 里的 type 字段决定前端如何渲染：
//
//	meta / step / delta / card / action / done / error
//
// scene 决定用哪套 JWT 识别身份：C 端用 JwtAuth，管理端用 AdminJwtAuth，
// 同时也决定了可用的工具集合。
func ChatStreamHandler(svcCtx *svc.ServiceContext, scene string) http.HandlerFunc {
	secret := svcCtx.Config.JwtAuth.AccessSecret
	if scene == dao.SceneAdmin {
		secret = svcCtx.Config.AdminJwtAuth.AccessSecret
	}

	return func(w http.ResponseWriter, r *http.Request) {
		token := auth.Token(r)
		uid, err := auth.Parse(secret, token)
		if err != nil {
			httpx.WriteJson(w, http.StatusUnauthorized, map[string]any{"code": 401, "msg": err.Error()})
			return
		}

		var req types.ChatStreamReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			httpx.WriteJson(w, http.StatusBadRequest, map[string]any{"code": 400, "msg": "请求体解析失败"})
			return
		}
		if strings.TrimSpace(req.Content) == "" {
			httpx.WriteJson(w, http.StatusBadRequest, map[string]any{"code": 400, "msg": "content 不能为空"})
			return
		}

		flusher, _ := w.(http.Flusher)
		header := w.Header()
		header.Set("Content-Type", "text/event-stream; charset=utf-8")
		header.Set("Cache-Control", "no-cache, no-transform")
		header.Set("Connection", "keep-alive")
		// 让 nginx 不要缓冲，否则流式会被攒成一坨
		header.Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		emit := func(ev agent.Event) error {
			b, err := json.Marshal(ev)
			if err != nil {
				return nil
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", b); err != nil {
				return err
			}
			if flusher != nil {
				flusher.Flush()
			}
			return nil
		}

		// 先发一个 ping，确保前端立刻拿到响应头（也避免中间层缓冲）
		if err := emit(agent.Event{Type: "ping"}); err != nil {
			return
		}

		_ = svcCtx.Runner.Run(r.Context(), agent.RunReq{
			Scene:          scene,
			OwnerID:        uid,
			Token:          token,
			ConversationID: req.ConversationId,
			Input:          req.Content,
		}, emit)
	}
}
