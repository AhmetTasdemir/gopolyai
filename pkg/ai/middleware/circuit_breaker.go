package middleware

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	provider         ai.AIProvider
	failureThreshold int
	resetTimeout     time.Duration

	mu              sync.Mutex
	state           CircuitState
	failures        int
	lastFailureTime time.Time
}

func NewCircuitBreaker(p ai.AIProvider, threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		provider:         p,
		failureThreshold: threshold,
		resetTimeout:     timeout,
		state:            StateClosed,
	}
}

func (cb *CircuitBreaker) Configure(cfg ai.Config) error {
	return cb.provider.Configure(cfg)
}

func (cb *CircuitBreaker) Name() string {
	return fmt.Sprintf("%s (Protected)", cb.provider.Name())
}

func (cb *CircuitBreaker) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	cb.mu.Lock()

	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			fmt.Println("⚡ [Circuit Breaker] Süre doldu, sistem test ediliyor (Half-Open)...")
		} else {
			cb.mu.Unlock()
			return nil, fmt.Errorf("circuit breaker is OPEN: requests are blocked for safety")
		}
	}
	cb.mu.Unlock()
	resp, err := cb.provider.Generate(ctx, req)

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		fmt.Printf("⚡ [Circuit Breaker] Hata algılandı (%d/%d)\n", cb.failures, cb.failureThreshold)

		if cb.failures >= cb.failureThreshold {
			if cb.state != StateOpen {
				fmt.Println("🔥 [Circuit Breaker] EŞİK AŞILDI! DEVRE AÇILIYOR (Sistem Korumada).")
			}
			cb.state = StateOpen
		}
		return nil, err
	}

	if cb.state == StateHalfOpen {
		fmt.Println("✅ [Circuit Breaker] Test başarılı! Devre kapatılıyor (Sistem Normale Döndü).")
	}
	cb.failures = 0
	cb.state = StateClosed

	return resp, nil
}

func (cb *CircuitBreaker) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	return cb.provider.GenerateStream(ctx, req)
}
