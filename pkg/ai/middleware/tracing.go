package middleware

import (
	"context"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/google/uuid"
)

type contextKey string

const (
	TraceIDKey contextKey = "trace_id"
)

type TracingMiddleware struct {
	next ai.AIProvider
}

func NewTracingMiddleware(next ai.AIProvider) *TracingMiddleware {
	return &TracingMiddleware{
		next: next,
	}
}

func (t *TracingMiddleware) Configure(cfg ai.Config) error {
	return t.next.Configure(cfg)
}

func (t *TracingMiddleware) Name() string {
	return t.next.Name()
}

func (t *TracingMiddleware) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	ctx = t.ensureTraceID(ctx)
	return t.next.Generate(ctx, req)
}

func (t *TracingMiddleware) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	ctx = t.ensureTraceID(ctx)
	return t.next.GenerateStream(ctx, req)
}

func (t *TracingMiddleware) ensureTraceID(ctx context.Context) context.Context {
	if ctx.Value(TraceIDKey) != nil {
		return ctx
	}
	id := uuid.New().String()
	return context.WithValue(ctx, TraceIDKey, id)
}

func GetTraceID(ctx context.Context) string {
	if val, ok := ctx.Value(TraceIDKey).(string); ok {
		return val
	}
	return ""
}
