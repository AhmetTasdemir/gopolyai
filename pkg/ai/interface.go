package ai

import "context"

type AIProvider interface {
	Configure(cfg Config) error

	Generate(ctx context.Context, req ChatRequest) (*ChatResponse, error)
	GenerateStream(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error)

	Name() string
}

type StreamResponse struct {
	Chunk string
	Err   error
	Usage *TokenUsage
}
