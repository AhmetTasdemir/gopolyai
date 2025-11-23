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

// GÜNCELLENDİ: String yerine ChatRequest alıyor, ChatResponse dönüyor
func (f *FallbackClient) Generate(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	fmt.Printf("[Fallback] %s deneniyor...\n", f.Primary.Name())

	resp, err := f.Primary.Generate(ctx, req)
	if err == nil {
		return resp, nil
	}

	fmt.Printf("[UYARI] Primary hata verdi: %v. Secondary (%s) devreye giriyor...\n", err, f.Secondary.Name())

	respSec, errSec := f.Secondary.Generate(ctx, req)
	if errSec != nil {
		return nil, errors.New("SYSTEM FAILURE: Both providers failed")
	}

	return respSec, nil
}

// GÜNCELLENDİ: Stream imzası da değişti
func (f *FallbackClient) GenerateStream(ctx context.Context, req ChatRequest) (<-chan StreamResponse, error) {
	// Şimdilik basit geçiş, ileride detaylandıracağız
	return f.Primary.GenerateStream(ctx, req)
}
