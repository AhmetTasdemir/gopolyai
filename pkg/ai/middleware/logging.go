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

	var usage ai.TokenUsage
	var responseContent string

	if resp != nil {
		usage = resp.Usage
		if l.config.LogPayloads {
			responseContent = resp.Content
		}
	}

	var requestContent string
	if l.config.LogPayloads {
		raw := fmt.Sprintf("%v", req.Messages)
		if len(raw) > 2000 {
			requestContent = raw[:2000] + " ... (truncated)"
		} else {
			requestContent = raw
		}
	}

	go func(u ai.TokenUsage, resP string, reqP string, finalErr error, dur time.Duration) {
		if l.config.LogErrorsOnly && finalErr == nil {
			return
		}

		entry := logger.LogEntry{
			Timestamp:       start.Add(dur),
			Duration:        dur,
			Provider:        l.next.Name(),
			Model:           req.Model,
			Operation:       "Generate",
			Error:           finalErr,
			TraceID:         GetTraceID(ctx),
			InputTokens:     u.InputTokens,
			OutputTokens:    u.OutputTokens,
			TotalTokens:     u.TotalTokens,
			CostUSD:         u.CostUSD,
			RequestPayload:  reqP,
			ResponsePayload: resP,
		}

		l.logger.Log(context.Background(), entry)
	}(usage, responseContent, requestContent, err, duration)

	return resp, err
}

func (l *LoggingMiddleware) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	start := time.Now()

	originalChan, err := l.next.GenerateStream(ctx, req)
	if err != nil {
		traceID := GetTraceID(ctx)
		go l.logStreamSummary(start, req.Model, nil, "", err, traceID)
		return nil, err
	}

	proxyChan := make(chan ai.StreamResponse, 10)
	traceID := GetTraceID(ctx)

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

		l.logStreamSummary(start, req.Model, finalUsage, fullContentBuilder, lastErr, traceID)
	}()

	return proxyChan, nil
}

func (l *LoggingMiddleware) logStreamSummary(start time.Time, model string, usage *ai.TokenUsage, content string, err error, traceID string) {
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
		TraceID:         traceID,
	}

	if usage != nil {
		entry.InputTokens = usage.InputTokens
		entry.OutputTokens = usage.OutputTokens
		entry.TotalTokens = usage.TotalTokens
		entry.CostUSD = usage.CostUSD
	}

	l.logger.Log(context.Background(), entry)
}
