package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

const defaultBaseURL = "https://api.anthropic.com/v1/messages"

type claudeRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

type Client struct {
	apiKey string
	model  string
	client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  "claude-3-5-sonnet-20240620",
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Configure(cfg ai.Config) error {
	c.apiKey = cfg.APIKey
	if cfg.ModelName != "" {
		c.model = cfg.ModelName
	}
	return nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := claudeRequest{
		Model:     c.model,
		MaxTokens: 1024,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", defaultBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic status: %d", resp.StatusCode)
	}

	var apiResp claudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if len(apiResp.Content) > 0 {
		return apiResp.Content[0].Text, nil
	}

	return "", fmt.Errorf("no content from claude")
}

func (c *Client) Name() string {
	return "Anthropic Claude (" + c.model + ")"
}
