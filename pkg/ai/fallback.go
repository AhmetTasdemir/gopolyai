package ai

import (
	"context"
	"errors"
	"fmt"
)

type FallbackClient struct {
	Primary   AIProvider
	Secondary AIProvider
}

func NewFallbackClient(primary, secondary AIProvider) *FallbackClient {
	return &FallbackClient{
		Primary:   primary,
		Secondary: secondary,
	}
}

func (f *FallbackClient) Configure(cfg Config) error {
	if err := f.Primary.Configure(cfg); err != nil {
		return err
	}
	return f.Secondary.Configure(cfg)
}

func (f *FallbackClient) Name() string {
	return fmt.Sprintf("SmartFallback (Pri: %s -> Sec: %s)", f.Primary.Name(), f.Secondary.Name())
}

// UPDATED: Takes ChatRequest instead of String, returns ChatResponse
func (f *FallbackClient) Generate(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	fmt.Printf("[Fallback] Trying %s...\n", f.Primary.Name())

	resp, err := f.Primary.Generate(ctx, req)
	if err == nil {
		return resp, nil
	}

	fmt.Printf("[WARNING] Primary failed: %v. Secondary (%s) activating...\n", err, f.Secondary.Name())

	respSec, errSec := f.Secondary.Generate(ctx, req)
	if errSec != nil {
		return nil, errors.New("SYSTEM FAILURE: Both providers failed")
	}

	return respSec, nil
}

// UPDATED: Stream signature also changed
func (f *FallbackClient) GenerateStream(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error) {
	// Simple switch for now, will detail later
	return f.Primary.GenerateStream(ctx, req)
}
