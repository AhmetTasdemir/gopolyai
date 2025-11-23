package ai

import (
	"errors"
	"time"
)

type Config struct {
	APIKey      string
	BaseURL     string
	ModelName   string
	MaxTokens   int
	Temperature float64
	Timeout     time.Duration
}

type ModelType string

const (
	ModelText  ModelType = "text"
	ModelImage ModelType = "image"
	ModelAudio ModelType = "audio"
)

var (
	ErrModelOverloaded = errors.New("model is currently overloaded")
	ErrContextExceeded = errors.New("context window exceeded")
	ErrProviderDown    = errors.New("provider is unreachable")
)

type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	JSONMode    bool          `json:"json_mode,omitempty"`
}

type ChatMessage struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type     string  `json:"type"`
	Text     string  `json:"text,omitempty"`
	ImageURL *string `json:"image_url,omitempty"`
}

type ChatResponse struct {
	Content string     `json:"content"`
	Usage   TokenUsage `json:"usage"`
	Cached  bool       `json:"cached"`
}

type TokenUsage struct {
	InputTokens  int     `json:"prompt_tokens"`
	OutputTokens int     `json:"completion_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	CostUSD      float64 `json:"cost_usd,omitempty"`
}
