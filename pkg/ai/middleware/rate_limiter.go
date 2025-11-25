package middleware

import (
	"context"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"golang.org/x/time/rate"
)

type RateLimiterMiddleware struct {
	next    ai.AIProvider
	limiter *rate.Limiter
}

func NewRateLimiterMiddleware(next ai.AIProvider, rps int, burst int) *RateLimiterMiddleware {
	limit := rate.Limit(rps)
	if rps <= 0 {
		limit = rate.Inf
	}

	return &RateLimiterMiddleware{
		next:    next,
		limiter: rate.NewLimiter(limit, burst),
	}
}

func (r *RateLimiterMiddleware) Configure(cfg ai.Config) error {
	return r.next.Configure(cfg)
}

func (r *RateLimiterMiddleware) Name() string {
	return r.next.Name()
}

func (r *RateLimiterMiddleware) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return r.next.Generate(ctx, req)
}

func (r *RateLimiterMiddleware) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	if err := r.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return r.next.GenerateStream(ctx, req)
}
