package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

const defaultBaseURL = "https://api.openai.com/v1/chat/completions"

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   "gpt-3.5-turbo",
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *Client) Configure(cfg ai.Config) error {
	if cfg.APIKey != "" {
		c.apiKey = cfg.APIKey
	}
	if cfg.ModelName != "" {
		c.model = cfg.ModelName
	}
	if cfg.BaseURL != "" {
		c.baseURL = cfg.BaseURL
	}
	if cfg.Timeout > 0 {
		c.httpClient.Timeout = cfg.Timeout
	}
	return nil
}

func (c *Client) Name() string {
	return "OpenAI (" + c.model + ")"
}

func (c *Client) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {

	openaiReq := map[string]interface{}{
		"model":       c.model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
	}

	if req.Model != "" {
		openaiReq["model"] = req.Model
	}

	jsonData, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("json marshal error: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		return nil, fmt.Errorf("openai status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("json decode error: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from openai")
	}

	return &ai.ChatResponse{
		Content: apiResp.Choices[0].Message.Content,
		Usage: ai.TokenUsage{
			InputTokens:  apiResp.Usage.PromptTokens,
			OutputTokens: apiResp.Usage.CompletionTokens,
			TotalTokens:  apiResp.Usage.TotalTokens,
		},
	}, nil
}

func (c *Client) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	streamChan := make(chan ai.StreamResponse, 10)

	openaiReq := map[string]interface{}{
		"model":       c.model,
		"messages":    req.Messages,
		"temperature": req.Temperature,
		"stream":      true,

		"stream_options": map[string]bool{"include_usage": true},
	}
	if req.Model != "" {
		openaiReq["model"] = req.Model
	}

	jsonData, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("openai stream status: %d", resp.StatusCode)
	}

	go func() {
		defer resp.Body.Close()
		defer close(streamChan)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
				Usage *struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 {
				content := chunk.Choices[0].Delta.Content
				if content != "" {
					streamChan <- ai.StreamResponse{Chunk: content}
				}
			}

			if chunk.Usage != nil {
				streamChan <- ai.StreamResponse{
					Usage: &ai.TokenUsage{
						InputTokens:  chunk.Usage.PromptTokens,
						OutputTokens: chunk.Usage.CompletionTokens,
						TotalTokens:  chunk.Usage.TotalTokens,
					},
				}
			}
		}
	}()

	return streamChan, nil
}
