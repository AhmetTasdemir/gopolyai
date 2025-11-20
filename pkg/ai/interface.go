package ai

import "context"

type AIProvider interface {
	Configure(cfg Config) error
	Generate(ctx context.Context, prompt string) (string, error)
	Name() string
}

type Config struct {
	APIKey      string
	BaseURL     string
	ModelName   string
	MaxTokens   int
	Temperature float64
}
