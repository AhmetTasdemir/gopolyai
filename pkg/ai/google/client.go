package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

type geminiRequest struct {
	Contents []content `json:"contents"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
}

type Client struct {
	apiKey string
	model  string
	client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  "gemini-1.5-flash",
		client: &http.Client{Timeout: 30 * time.Second},
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

	reqBody := geminiRequest{
		Contents: []content{
			{Parts: []part{{Text: prompt}}},
		},
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(baseURL, c.model, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("google api error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google returned status: %d", resp.StatusCode)
	}

	var apiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if len(apiResp.Candidates) > 0 && len(apiResp.Candidates[0].Content.Parts) > 0 {
		return apiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no response from gemini")
}

func (c *Client) Name() string {
	return "Google Gemini (" + c.model + ")"
}
