package logger

import (
	"context"
	"time"
)

type LogEntry struct {
	ID        string
	Timestamp time.Time
	Duration  time.Duration
	Provider  string
	Model     string
	Operation string
	Error     error

	InputTokens  int
	OutputTokens int
	TotalTokens  int
	CostUSD      float64

	RequestPayload  string
	ResponsePayload string
	TraceID         string
}

type Logger interface {
	Log(ctx context.Context, entry LogEntry)
}

type NoOpLogger struct{}

func (n *NoOpLogger) Log(ctx context.Context, entry LogEntry) {}

type Config struct {
	LogPayloads   bool
	LogErrorsOnly bool
}
