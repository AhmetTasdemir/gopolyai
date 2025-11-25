package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

type MockProvider struct {
	CallCount int
}

func (m *MockProvider) Configure(cfg ai.Config) error { return nil }
func (m *MockProvider) Name() string                  { return "Mock" }
func (m *MockProvider) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	m.CallCount++
	return &ai.ChatResponse{Content: "ok"}, nil
}
func (m *MockProvider) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	return nil, nil
}

func TestRateLimiter(t *testing.T) {
	mock := &MockProvider{}
	limiter := NewRateLimiterMiddleware(mock, 5, 5)

	start := time.Now()
	for i := 0; i < 10; i++ {
		_, err := limiter.Generate(context.Background(), ai.ChatRequest{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	duration := time.Since(start)

	if duration < 900*time.Millisecond {
		t.Errorf("expected duration > 900ms, got %v", duration)
	}
}
