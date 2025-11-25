package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/logger"
)

type LoggingMiddleware struct {
	next   ai.AIProvider
	logger logger.Logger
	config logger.Config
}

func NewLoggingMiddleware(next ai.AIProvider, l logger.Logger, cfg logger.Config) *LoggingMiddleware {
	if l == nil {
		l = &logger.NoOpLogger{}
	}
	return &LoggingMiddleware{
		next:   next,
		logger: l,
		config: cfg,
	}
}

func (l *LoggingMiddleware) Configure(cfg ai.Config) error {
	return l.next.Configure(cfg)
}

func (l *LoggingMiddleware) Name() string {
	return l.next.Name()
}

func (l *LoggingMiddleware) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	start := time.Now()

	resp, err := l.next.Generate(ctx, req)

	duration := time.Since(start)

	go func(finalResp *ai.ChatResponse, finalErr error, dur time.Duration) {
		if l.config.LogErrorsOnly && finalErr == nil {
			return
		}

		entry := logger.LogEntry{
			Timestamp: start.Add(dur),
			Duration:  dur,
			Provider:  l.next.Name(),
			Model:     req.Model,
			Operation: "Generate",
			Error:     finalErr,
		}

		if l.config.LogPayloads {

			entry.RequestPayload = fmt.Sprintf("%v", req.Messages)
		}

		if finalResp != nil {
			entry.InputTokens = finalResp.Usage.InputTokens
			entry.OutputTokens = finalResp.Usage.OutputTokens
			entry.TotalTokens = finalResp.Usage.TotalTokens
			entry.CostUSD = finalResp.Usage.CostUSD

			if l.config.LogPayloads {
				entry.ResponsePayload = finalResp.Content
			}
		}

		l.logger.Log(context.Background(), entry)
	}(resp, err, duration)

	return resp, err
}

func (l *LoggingMiddleware) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	start := time.Now()

	originalChan, err := l.next.GenerateStream(ctx, req)
	if err != nil {

		go l.logStreamSummary(start, req.Model, nil, "", err)
		return nil, err
	}

	proxyChan := make(chan ai.StreamResponse, 10)

	go func() {
		defer close(proxyChan)

		var fullContentBuilder string
		var finalUsage *ai.TokenUsage
		var lastErr error

		for packet := range originalChan {

			if packet.Err != nil {
				lastErr = packet.Err
			}

			if l.config.LogPayloads && packet.Chunk != "" {
				fullContentBuilder += packet.Chunk
			}

			if packet.Usage != nil {
				finalUsage = packet.Usage
			}

			proxyChan <- packet
		}

		l.logStreamSummary(start, req.Model, finalUsage, fullContentBuilder, lastErr)
	}()

	return proxyChan, nil
}

func (l *LoggingMiddleware) logStreamSummary(start time.Time, model string, usage *ai.TokenUsage, content string, err error) {
	if l.config.LogErrorsOnly && err == nil {
		return
	}

	duration := time.Since(start)
	entry := logger.LogEntry{
		Timestamp:       start.Add(duration),
		Duration:        duration,
		Provider:        l.next.Name(),
		Model:           model,
		Operation:       "GenerateStream",
		Error:           err,
		ResponsePayload: content,
	}

	if usage != nil {
		entry.InputTokens = usage.InputTokens
		entry.OutputTokens = usage.OutputTokens
		entry.TotalTokens = usage.TotalTokens
		entry.CostUSD = usage.CostUSD
	}

	l.logger.Log(context.Background(), entry)
}
