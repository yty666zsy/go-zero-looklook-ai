package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider 对接任意 OpenAI 兼容的 /chat/completions 网关。
type OpenAIProvider struct {
	baseUrl     string
	apiKey      string
	model       string
	temperature float64
	maxTokens   int
	client      *http.Client
}

func NewOpenAI(baseUrl, apiKey, model string, temperature float64, maxTokens int, timeoutMs int64) *OpenAIProvider {
	if timeoutMs <= 0 {
		timeoutMs = 60000
	}
	return &OpenAIProvider{
		baseUrl:     strings.TrimRight(baseUrl, "/"),
		apiKey:      apiKey,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		client: &http.Client{
			// 注意：流式响应不能设 Client.Timeout（会掐断长连接），用 context 控制
			Timeout: 0,
			Transport: &http.Transport{
				MaxIdleConns:        20,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (p *OpenAIProvider) Name() string { return "openai:" + p.model }

type oaRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Tools       []ToolSpec `json:"tools,omitempty"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	StreamOptions *struct {
		IncludeUsage bool `json:"include_usage"`
	} `json:"stream_options,omitempty"`
}

type oaChunk struct {
	Model   string `json:"model"`
	Choices []struct {
		Delta struct {
			Content   string `json:"content"`
			ToolCalls []struct {
				Index    int    `json:"index"`
				Id       string `json:"id"`
				Type     string `json:"type"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *OpenAIProvider) ChatStream(ctx context.Context, req Request, onDelta func(string) error) (*Response, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("LLM ApiKey 未配置")
	}

	body := oaRequest{
		Model:       p.model,
		Messages:    req.Messages,
		Tools:       req.Tools,
		Stream:      true,
		Temperature: pickFloat(req.Temperature, p.temperature),
		MaxTokens:   pickInt(req.MaxTokens, p.maxTokens),
	}
	body.StreamOptions = &struct {
		IncludeUsage bool `json:"include_usage"`
	}{IncludeUsage: true}

	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(p.timeout()) * time.Millisecond)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseUrl+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("LLM 返回 %d: %s", resp.StatusCode, string(b))
	}

	out := &Response{Model: p.model}
	// tool_calls 会分片下发，按 index 聚合
	acc := map[int]*ToolCall{}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}

		var chunk oaChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			continue
		}
		if chunk.Error != nil {
			return nil, fmt.Errorf("LLM 错误: %s", chunk.Error.Message)
		}
		if chunk.Model != "" {
			out.Model = chunk.Model
		}
		if chunk.Usage != nil {
			out.Usage.PromptTokens = chunk.Usage.PromptTokens
			out.Usage.CompletionTokens = chunk.Usage.CompletionTokens
		}
		if len(chunk.Choices) == 0 {
			continue
		}
		d := chunk.Choices[0].Delta
		if d.Content != "" {
			out.Content += d.Content
			if onDelta != nil {
				if err := onDelta(d.Content); err != nil {
					return out, err
				}
			}
		}
		for _, tc := range d.ToolCalls {
			item, ok := acc[tc.Index]
			if !ok {
				item = &ToolCall{Type: "function"}
				acc[tc.Index] = item
			}
			if tc.Id != "" {
				item.Id = tc.Id
			}
			if tc.Type != "" {
				item.Type = tc.Type
			}
			if tc.Function.Name != "" {
				item.Function.Name = tc.Function.Name
			}
			item.Function.Arguments += tc.Function.Arguments
		}
	}
	if err := scanner.Err(); err != nil {
		return out, err
	}

	for i := 0; i < len(acc); i++ {
		if tc, ok := acc[i]; ok {
			out.ToolCalls = append(out.ToolCalls, *tc)
		}
	}
	return out, nil
}

func (p *OpenAIProvider) timeout() int64 { return 60000 }

func pickFloat(v, def float64) float64 {
	if v == 0 {
		return def
	}
	return v
}

func pickInt(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
