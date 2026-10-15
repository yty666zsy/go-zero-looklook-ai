// Package auth 负责从请求里取 token 并解析出身份。
//
// 标准的 JSON 接口走 go-zero 的 rest.WithJwt；只有 SSE 流式接口是自己写的
// http.Handler（需要 Flusher），所以这里额外提供一个手写解析。
package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"looklook/pkg/ctxdata"

	"github.com/golang-jwt/jwt/v4"
)

// Token 从 Authorization 头里取 Bearer token
func Token(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	if len(h) > 7 && strings.EqualFold(h[:7], "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return strings.TrimSpace(h)
}

// Parse 校验并解析 token，返回 userId/adminId
func Parse(secret, token string) (int64, error) {
	if token == "" {
		return 0, fmt.Errorf("未登录")
	}
	claims := make(jwt.MapClaims)
	if _, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	}); err != nil {
		return 0, fmt.Errorf("登录状态已失效，请重新登录")
	}
	v, ok := claims[ctxdata.CtxKeyJwtUserId]
	if !ok {
		return 0, fmt.Errorf("token 缺少用户标识")
	}
	switch t := v.(type) {
	case float64:
		return int64(t), nil
	case json.Number:
		n, err := t.Int64()
		if err != nil {
			return 0, fmt.Errorf("token 用户标识非法")
		}
		return n, nil
	case int64:
		return t, nil
	}
	return 0, fmt.Errorf("token 用户标识非法")
}
