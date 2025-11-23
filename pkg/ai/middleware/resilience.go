package middleware

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type ResilientClient struct {
	provider ai.AIProvider
	config   RetryConfig
}

func NewResilientClient(p ai.AIProvider, cfg RetryConfig) *ResilientClient {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 3
	}
	if cfg.BaseDelay <= 0 {
		cfg.BaseDelay = 2 * time.Second
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = 30 * time.Second
	}

	return &ResilientClient{
		provider: p,
		config:   cfg,
	}
}

func (r *ResilientClient) Configure(cfg ai.Config) error {
	return r.provider.Configure(cfg)
}

func (r *ResilientClient) Name() string {
	return fmt.Sprintf("%s (Resilient)", r.provider.Name())
}

func (r *ResilientClient) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	var lastErr error

	for i := 0; i <= r.config.MaxRetries; i++ {
		resp, err := r.provider.Generate(ctx, req)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		fmt.Printf("[%s] Hata (Deneme %d/%d): %v\n", r.provider.Name(), i+1, r.config.MaxRetries+1, err)

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		if i == r.config.MaxRetries {
			break
		}

		backoff := float64(r.config.BaseDelay) * math.Pow(2, float64(i))
		sleepDuration := time.Duration(backoff)

		if sleepDuration > r.config.MaxDelay {
			sleepDuration = r.config.MaxDelay
		}

		fmt.Printf(">>> %v bekleyip tekrar deniyor...\n", sleepDuration)

		select {
		case <-time.After(sleepDuration):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (r *ResilientClient) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	return r.provider.GenerateStream(ctx, req)
}
