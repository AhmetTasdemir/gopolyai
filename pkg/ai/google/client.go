package google

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

const baseURL = "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s"

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   "gemini-1.5-flash",
		baseURL: baseURL,
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
	return "Google Gemini (" + c.model + ")"
}

type geminiRequest struct {
	Contents         []geminiContent `json:"contents"`
	GenerationConfig genConfig       `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text,omitempty"`
}

type genConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

func (c *Client) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {

	var gContents []geminiContent

	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}

		var textContent string
		for _, content := range msg.Content {
			if content.Type == "text" {
				textContent += content.Text
			}
		}

		gContents = append(gContents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: textContent}},
		})
	}

	geminiReq := geminiRequest{
		Contents: gContents,
		GenerationConfig: genConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		},
	}

	jsonData, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(c.baseURL, c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google api status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
			TotalTokenCount      int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	if len(apiResp.Candidates) == 0 || len(apiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from google")
	}

	return &ai.ChatResponse{
		Content: apiResp.Candidates[0].Content.Parts[0].Text,
		Usage: ai.TokenUsage{
			InputTokens:  apiResp.UsageMetadata.PromptTokenCount,
			OutputTokens: apiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  apiResp.UsageMetadata.TotalTokenCount,
		},
	}, nil
}

func (c *Client) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	streamChan := make(chan ai.StreamResponse, 10)

	var gContents []geminiContent
	for _, msg := range req.Messages {
		role := "user"
		if msg.Role == "assistant" {
			role = "model"
		}
		var textContent string
		for _, content := range msg.Content {
			if content.Type == "text" {
				textContent += content.Text
			}
		}
		gContents = append(gContents, geminiContent{Role: role, Parts: []geminiPart{{Text: textContent}}})
	}

	geminiReq := geminiRequest{
		Contents: gContents,
		GenerationConfig: genConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
		},
	}

	jsonData, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(c.baseURL, c.model, c.apiKey)
	streamURL := strings.Replace(url, "generateContent", "streamGenerateContent", 1)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", streamURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("google stream status: %d", resp.StatusCode)
	}

	go func() {
		defer resp.Body.Close()
		defer close(streamChan)

		decoder := json.NewDecoder(resp.Body)
		if token, err := decoder.Token(); err == nil {
			if delim, ok := token.(json.Delim); ok && delim == '[' {

			}
		}

		for decoder.More() {
			var chunk struct {
				Candidates []struct {
					Content struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					} `json:"content"`
				} `json:"candidates"`
				UsageMetadata *struct {
					PromptTokenCount     int `json:"promptTokenCount"`
					CandidatesTokenCount int `json:"candidatesTokenCount"`
					TotalTokenCount      int `json:"totalTokenCount"`
				} `json:"usageMetadata"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err.Error() == "EOF" {
					return
				}
				streamChan <- ai.StreamResponse{Err: fmt.Errorf("google decode error: %w", err)}
				return
			}

			if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
				text := chunk.Candidates[0].Content.Parts[0].Text
				streamChan <- ai.StreamResponse{Chunk: text}
			}

			if chunk.UsageMetadata != nil {
				streamChan <- ai.StreamResponse{
					Usage: &ai.TokenUsage{
						InputTokens:  chunk.UsageMetadata.PromptTokenCount,
						OutputTokens: chunk.UsageMetadata.CandidatesTokenCount,
						TotalTokens:  chunk.UsageMetadata.TotalTokenCount,
					},
				}
			}
		}
	}()

	return streamChan, nil
}
