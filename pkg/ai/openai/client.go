package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"gopolyai/pkg/ai"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.openai.com/v1/chat/completions"

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  "gpt-3.5-turbo",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Configure(cfg ai.Config) error {
	if cfg.APIKey == "" {
		return errors.New("API key is required for OpenAI")
	}
	c.apiKey = cfg.APIKey

	if cfg.ModelName != "" {
		c.model = cfg.ModelName
	}
	return nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := chatCompletionRequest{
		Model: c.model,
		Messages: []message{
			{Role: "user", Content: prompt},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("JSON marshal error: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", defaultBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned non-200 status: %d", resp.StatusCode)
	}

	var apiResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", fmt.Errorf("JSON decode error: %w", err)
	}

	if len(apiResp.Choices) > 0 {
		return apiResp.Choices[0].Message.Content, nil
	}

	return "", errors.New("no response choice from OpenAI")
}

func (c *Client) Name() string {
	return "OpenAI (" + c.model + ")"
}
