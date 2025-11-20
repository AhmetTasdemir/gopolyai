package mock

import (
	"context"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

type MockClient struct {
	FixedResponse string
	ShouldFail    bool
}

func NewClient(response string, shouldFail bool) *MockClient {
	return &MockClient{
		FixedResponse: response,
		ShouldFail:    shouldFail,
	}
}

func (m *MockClient) Configure(cfg ai.Config) error {
	return nil
}

func (m *MockClient) Generate(ctx context.Context, prompt string) (string, error) {

	time.Sleep(50 * time.Millisecond)

	if m.ShouldFail {
		return "", context.DeadlineExceeded
	}

	return "MOCK CEVAP: " + m.FixedResponse, nil
}

func (m *MockClient) Name() string {
	return "Mock AI (Testing)"
}
