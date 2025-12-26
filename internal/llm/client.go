package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cgpd/internal/config"
)

type Client interface {
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)
	GenerateDocs(ctx context.Context, diff string) (string, error)
}

func NewClient(cfg config.LLMConfig) (Client, error) {
	provider := strings.TrimSpace(strings.ToLower(cfg.Provider))
	if provider == "" {
		return nil, errors.New("llm.provider is required (set in .cgpd.yaml or env CGPD_LLM_PROVIDER)")
	}

	lang := strings.TrimSpace(strings.ToLower(cfg.Language))
	if lang == "" {
		lang = "en"
	}
	if lang != "en" && lang != "zh" {
		return nil, fmt.Errorf("unsupported llm.language %q (supported: en, zh)", lang)
	}

	switch provider {
	case "openai", "openai-compatible":
		baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
		if baseURL == "" {
			baseURL = "https://api.openai.com"
		}
		apiKey := strings.TrimSpace(cfg.APIKey)
		if apiKey == "" {
			return nil, errors.New("llm.api_key is required (set in .cgpd.yaml or env OPENAI_API_KEY)")
		}
		model := strings.TrimSpace(cfg.Model)
		if model == "" {
			return nil, errors.New("llm.model is required (set in .cgpd.yaml or env CGPD_LLM_MODEL)")
		}
		if _, err := buildEndpoint(baseURL); err != nil {
			return nil, err
		}
		return &openaiClient{
			baseURL:  baseURL,
			apiKey:   apiKey,
			model:    model,
			language: lang,
			http:     &http.Client{Timeout: 75 * time.Second},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported llm.provider %q", provider)
	}
}

type openaiClient struct {
	baseURL  string
	apiKey   string
	model    string
	language string
	http     *http.Client
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

const maxResponseBytes = 2 << 20 // 2 MiB

var commitPrompts = map[string]string{
	"en": `You are a Git commit message generator.
Rules:
- Output ONLY the commit subject line, no quotes, no markdown
- Use imperative mood (Add, Fix, Refactor, Update)
- Maximum 72 characters
- Be specific to the actual changes
- Response in English`,
	"zh": `你是一个 Git commit 信息生成器。
规则：
- 仅输出 commit 主题行，不要引号，不要 markdown
- 使用祈使语气（添加、修复、重构、更新）
- 最多 72 个字符
- 具体描述实际的变更内容
- 使用简体中文回复`,
}

var docsPrompts = map[string]string{
	"en": `You are a changelog generator.
Rules:
- Output valid Markdown
- Start with a brief summary section
- Group changes by category (API, Config, Docs, Tests) when applicable
- Include behavior changes and migration notes if needed
- Response in English`,
	"zh": `你是一个变更日志生成器。
规则：
- 输出有效的 Markdown 格式
- 以简要概述开头
- 按类别分组（API、配置、文档、测试）
- 包含行为变更和迁移说明（如需要）
- 使用简体中文回复`,
}

func (c *openaiClient) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := commitPrompts[c.language]
	return c.chat(ctx, prompt, "Staged diff:\n\n"+diff, 0.2)
}

func (c *openaiClient) GenerateDocs(ctx context.Context, diff string) (string, error) {
	prompt := docsPrompts[c.language]
	return c.chat(ctx, prompt, "Staged diff:\n\n"+diff, 0.2)
}

func buildEndpoint(baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	u, err := url.Parse(baseURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("invalid llm.base_url %q", baseURL)
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return "", fmt.Errorf("unsupported llm.base_url scheme %q", u.Scheme)
	}

	path := strings.TrimRight(u.Path, "/")
	if path == "/v1" {
		return baseURL + "/chat/completions", nil
	}
	return baseURL + "/v1/chat/completions", nil
}

func (c *openaiClient) chat(ctx context.Context, system, user string, temp float64) (string, error) {
	body := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
		Temperature: temp,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("encode request: %w", err)
	}

	endpoint, err := buildEndpoint(c.baseURL)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}
	if len(respBody) > maxResponseBytes {
		return "", fmt.Errorf("response exceeds %d bytes", maxResponseBytes)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var parsed chatResponse
		if json.Unmarshal(respBody, &parsed) == nil && parsed.Error != nil {
			return "", fmt.Errorf("LLM error (%s): %s", resp.Status, parsed.Error.Message)
		}
		return "", fmt.Errorf("LLM error (%s): %s", resp.Status, strings.TrimSpace(string(respBody)))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("LLM returned no choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", errors.New("LLM returned an empty message")
	}
	return content, nil
}
