// Package biz 封装对现有业务接口的调用。
//
// 设计取舍：Agent 不重复实现业务逻辑，而是把「工具」映射到已有的 HTTP 接口，
// 并把调用方的 JWT 一路透传下去。这样：
//   - 业务逻辑只有一份，C 端接口改了 Agent 自动跟着变；
//   - 权限模型完全复用（C 端用户只能查到自己的订单，管理端 token 决定管理权限）。
package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	travelBase string
	orderBase  string
	adminBase  string
	client     *http.Client
}

func New(travelBase, orderBase, adminBase string, timeoutMs int64) *Client {
	if timeoutMs <= 0 {
		timeoutMs = 8000
	}
	return &Client{
		travelBase: trimRight(travelBase),
		orderBase:  trimRight(orderBase),
		adminBase:  trimRight(adminBase),
		client: &http.Client{
			Timeout: time.Duration(timeoutMs) * time.Millisecond,
			Transport: &http.Transport{
				MaxIdleConns:        50,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// 统一响应包装
type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (c *Client) post(ctx context.Context, base, path, token string, req any, out any) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, base+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("调用业务接口失败: %w", err)
	}
	defer resp.Body.Close()

	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		return fmt.Errorf("解析业务接口响应失败: %w", err)
	}
	if env.Code != 200 {
		msg := env.Msg
		if msg == "" {
			msg = fmt.Sprintf("业务接口返回 code=%d", env.Code)
		}
		return fmt.Errorf("%s", msg)
	}
	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return fmt.Errorf("解析业务数据结构失败: %w", err)
		}
	}
	return nil
}

func trimRight(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
