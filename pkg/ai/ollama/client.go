package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

const defaultBaseURL = "http://localhost:11434/api/chat"

type Client struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		model:   "llama3",
		client:  &http.Client{Timeout: 0},
	}
}

func (c *Client) Configure(cfg ai.Config) error {
	if cfg.ModelName != "" {
		c.model = cfg.ModelName
	}
	if cfg.Timeout > 0 {
		c.client.Timeout = cfg.Timeout
	}
	return nil
}

func (c *Client) Name() string {
	return "Ollama Local (" + c.model + ")"
}

type ollamaMessage struct {
	Role    string   `json:"role"`
	Content string   `json:"content"`
	Images  []string `json:"images,omitempty"`
}

func (c *Client) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	modelToUse := c.model
	if req.Model != "" {
		modelToUse = req.Model
	}

	var oMessages []ollamaMessage

	for _, msg := range req.Messages {
		fullText := ""
		var images []string

		for _, part := range msg.Content {
			if part.Type == "text" {
				fullText += part.Text
			} else if part.Type == "image_url" && part.ImageURL != nil {
				images = append(images, *part.ImageURL)
			}
		}

		oMessages = append(oMessages, ollamaMessage{
			Role:    msg.Role,
			Content: fullText,
			Images:  images,
		})
	}

	ollamaReq := map[string]interface{}{
		"model":    modelToUse,
		"messages": oMessages,
		"stream":   false,
	}

	if req.Temperature > 0 {
		ollamaReq["options"] = map[string]float64{
			"temperature": req.Temperature,
		}
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama status: %d", resp.StatusCode)
	}

	var apiResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		PromptEvalCount int `json:"prompt_eval_count"`
		EvalCount       int `json:"eval_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &ai.ChatResponse{
		Content: apiResp.Message.Content,
		Usage: ai.TokenUsage{
			InputTokens:  apiResp.PromptEvalCount,
			OutputTokens: apiResp.EvalCount,
			TotalTokens:  apiResp.PromptEvalCount + apiResp.EvalCount,
		},
	}, nil
}

func (c *Client) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	streamChan := make(chan ai.StreamResponse, 10)

	modelToUse := c.model
	if req.Model != "" {
		modelToUse = req.Model
	}

	var oMessages []ollamaMessage
	for _, msg := range req.Messages {
		fullText := ""
		for _, part := range msg.Content {
			if part.Type == "text" {
				fullText += part.Text
			}

		}
		oMessages = append(oMessages, ollamaMessage{
			Role:    msg.Role,
			Content: fullText,
		})
	}

	ollamaReq := map[string]interface{}{
		"model":    modelToUse,
		"messages": oMessages,
		"stream":   true,
	}

	jsonData, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, ai.ErrProviderDown
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama stream status: %d", resp.StatusCode)
	}

	go func() {
		defer resp.Body.Close()
		defer close(streamChan)

		decoder := json.NewDecoder(resp.Body)

		for {
			var chunk struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
				Done            bool `json:"done"`
				PromptEvalCount int  `json:"prompt_eval_count"`
				EvalCount       int  `json:"eval_count"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err.Error() == "EOF" {
					return
				}
				streamChan <- ai.StreamResponse{Err: fmt.Errorf("ollama decode error: %w", err)}
				return
			}

			if chunk.Message.Content != "" {
				streamChan <- ai.StreamResponse{Chunk: chunk.Message.Content}
			}

			if chunk.Done {
				streamChan <- ai.StreamResponse{
					Usage: &ai.TokenUsage{
						InputTokens:  chunk.PromptEvalCount,
						OutputTokens: chunk.EvalCount,
						TotalTokens:  chunk.PromptEvalCount + chunk.EvalCount,
					},
				}
				return
			}
		}
	}()

	return streamChan, nil
}
