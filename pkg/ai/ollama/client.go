package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

const defaultBaseURL = "http://localhost:11434/api/generate"

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type Client struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		model:   "llama3", // Default model
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Configure(cfg ai.Config) error {
	if cfg.ModelName != "" {
		c.model = cfg.ModelName
	}
	if cfg.BaseURL != "" {
		c.baseURL = cfg.BaseURL
	}
	return nil
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := ollamaRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned non-200 status: %d", resp.StatusCode)
	}

	var apiResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}
	return apiResp.Response, nil
}

func (c *Client) Name() string {
	return "Ollama Local (" + c.model + ")"
}
