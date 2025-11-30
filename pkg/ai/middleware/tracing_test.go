package middleware

import (
	"context"
	"testing"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

type TracingMockProvider struct {
	CapturedContext context.Context
}

func (m *TracingMockProvider) Configure(cfg ai.Config) error { return nil }
func (m *TracingMockProvider) Name() string                  { return "mock" }
func (m *TracingMockProvider) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	m.CapturedContext = ctx
	return &ai.ChatResponse{}, nil
}
func (m *TracingMockProvider) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	m.CapturedContext = ctx
	ch := make(chan ai.StreamResponse)
	close(ch)
	return ch, nil
}

func TestTracingMiddleware_Generate(t *testing.T) {
	mock := &TracingMockProvider{}
	middleware := NewTracingMiddleware(mock)

	ctx := context.Background()
	middleware.Generate(ctx, ai.ChatRequest{})

	traceID := GetTraceID(mock.CapturedContext)
	if traceID == "" {
		t.Error("Expected TraceID to be generated, got empty string")
	}

	existingID := "existing-trace-id"
	ctx = context.WithValue(context.Background(), TraceIDKey, existingID)
	middleware.Generate(ctx, ai.ChatRequest{})

	traceID = GetTraceID(mock.CapturedContext)
	if traceID != existingID {
		t.Errorf("Expected TraceID to be preserved as %s, got %s", existingID, traceID)
	}
}

func TestTracingMiddleware_GenerateStream(t *testing.T) {
	mock := &TracingMockProvider{}
	middleware := NewTracingMiddleware(mock)
	ctx := context.Background()
	middleware.GenerateStream(ctx, ai.ChatRequest{})

	traceID := GetTraceID(mock.CapturedContext)
	if traceID == "" {
		t.Error("Expected TraceID to be generated in stream, got empty string")
	}
}
