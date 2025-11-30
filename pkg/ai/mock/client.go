package mock

import (
	"context"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

type MockClient struct {
	FixedResponse string
	ShouldFail    bool
	FailCount     int
}

func NewClient(response string, shouldFail bool) *MockClient {
	return &MockClient{
		FixedResponse: response,
		ShouldFail:    shouldFail,
	}
}

func (m *MockClient) Configure(cfg ai.Config) error { return nil }
func (m *MockClient) Name() string                  { return "Mock AI" }

// Stream is empty for now
func (m *MockClient) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	return nil, nil
}

// UPDATED: New struct structures
func (m *MockClient) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	if m.ShouldFail {
		m.FailCount++
		// Fail on first 2 attempts (To test Retry mechanism)
		if m.FailCount <= 2 {
			return nil, ai.ErrProviderDown
		}
	}

	time.Sleep(100 * time.Millisecond)

	return &ai.ChatResponse{
		Content: "MOCK: " + m.FixedResponse,
		Usage: ai.TokenUsage{
			InputTokens:  15,
			OutputTokens: 20,
			TotalTokens:  35,
			CostUSD:      0.0001,
		},
	}, nil
}
