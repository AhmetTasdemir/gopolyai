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

	// 1. CHECK STATE (Inside Lock)
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			// State transition: Open -> HalfOpen
			cb.state = StateHalfOpen
			fmt.Println("⚡ [Circuit Breaker] Timeout exceeded, testing system (Half-Open)...")
		} else {
			// Fast failure
			cb.mu.Unlock()
			return nil, fmt.Errorf("circuit breaker is OPEN: requests are blocked for safety")
		}
	}
	cb.mu.Unlock()

	// 2. EXECUTE PROVIDER (Outside Lock to allow concurrency)
	resp, err := cb.provider.Generate(ctx, req)

	// 3. UPDATE STATE (Inside Lock)
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailureTime = time.Now()

		fmt.Printf("⚡ [Circuit Breaker] Error detected (%d/%d)\n", cb.failures, cb.failureThreshold)

		if cb.failures >= cb.failureThreshold {
			if cb.state != StateOpen {
				fmt.Println("🔥 [Circuit Breaker] THRESHOLD EXCEEDED! CIRCUIT OPENING (System Protected).")
			}
			cb.state = StateOpen
		}
		return nil, err
	}

	// Success case
	if cb.state == StateHalfOpen {
		fmt.Println("✅ [Circuit Breaker] Test successful! Circuit closing (System Returned to Normal).")
		cb.state = StateClosed
		cb.failures = 0
	} else if cb.state == StateClosed {
		// Optional: Reset failure count on success to effectively use "consecutive failures" logic
		cb.failures = 0
	}

	return resp, nil
}

func (cb *CircuitBreaker) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	// For v1.4.0, Stream passes through. In v2.0, Full Stream Protection will be added.
	return cb.provider.GenerateStream(ctx, req)
}
