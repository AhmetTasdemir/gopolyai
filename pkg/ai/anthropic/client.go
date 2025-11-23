package anthropic

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

const defaultBaseURL = "https://api.anthropic.com/v1/messages"

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   "claude-3-5-sonnet-20240620",
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
	return "Anthropic Claude (" + c.model + ")"
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []claudeMessage `json:"messages"`
	Temp      float64         `json:"temperature,omitempty"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *Client) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {

	var cMessages []claudeMessage
	for _, msg := range req.Messages {
		var textContent string
		for _, content := range msg.Content {
			if content.Type == "text" {
				textContent += content.Text
			}
		}
		cMessages = append(cMessages, claudeMessage{
			Role:    msg.Role,
			Content: textContent,
		})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	claudeReq := claudeRequest{
		Model:     c.model,
		MaxTokens: maxTokens,
		Messages:  cMessages,
		Temp:      req.Temperature,
	}

	jsonData, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from claude")
	}

	return &ai.ChatResponse{
		Content: apiResp.Content[0].Text,
		Usage: ai.TokenUsage{
			InputTokens:  apiResp.Usage.InputTokens,
			OutputTokens: apiResp.Usage.OutputTokens,
			TotalTokens:  apiResp.Usage.InputTokens + apiResp.Usage.OutputTokens,
		},
	}, nil
}

func (c *Client) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	streamChan := make(chan ai.StreamResponse, 10)

	var cMessages []claudeMessage
	for _, msg := range req.Messages {
		var textContent string
		for _, content := range msg.Content {
			if content.Type == "text" {
				textContent += content.Text
			}
		}
		cMessages = append(cMessages, claudeMessage{Role: msg.Role, Content: textContent})
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	claudeReq := map[string]interface{}{
		"model":      c.model,
		"max_tokens": maxTokens,
		"messages":   cMessages,
		"stream":     true,
	}
	if req.Temperature > 0 {
		claudeReq["temperature"] = req.Temperature
	}

	jsonData, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", c.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("anthropic stream status: %d", resp.StatusCode)
	}

	go func() {
		defer resp.Body.Close()
		defer close(streamChan)

		var currentUsage ai.TokenUsage

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				return
			}

			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"delta"`
				Usage *struct {
					InputTokens  int `json:"input_tokens"`
					OutputTokens int `json:"output_tokens"`
				} `json:"usage"`
				Message *struct {
					Usage struct {
						InputTokens int `json:"input_tokens"`
					} `json:"usage"`
				} `json:"message"`
			}

			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}
			if event.Type == "content_block_delta" && event.Delta.Type == "text_delta" {
				streamChan <- ai.StreamResponse{Chunk: event.Delta.Text}
			}
			if event.Type == "message_start" && event.Message != nil {
				currentUsage.InputTokens = event.Message.Usage.InputTokens
			}

			if event.Type == "message_delta" && event.Usage != nil {
				currentUsage.OutputTokens = event.Usage.OutputTokens
				currentUsage.TotalTokens = currentUsage.InputTokens + currentUsage.OutputTokens

				streamChan <- ai.StreamResponse{
					Usage: &currentUsage,
				}
			}
		}
	}()

	return streamChan, nil
}
