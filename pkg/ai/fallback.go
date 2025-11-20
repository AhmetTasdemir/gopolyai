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

func (f *FallbackClient) Generate(ctx context.Context, prompt string) (string, error) {
	fmt.Printf("[Fallback Sistemi] %s deneniyor...\n", f.Primary.Name())
	resp, err := f.Primary.Generate(ctx, prompt)

	if err == nil {
		return resp, nil
	}

	fmt.Printf("[UYARI] Primary servis hata verdi: %v. Secondary (%s) devreye giriyor...\n", err, f.Secondary.Name())
	respSecondary, errSecondary := f.Secondary.Generate(ctx, prompt)
	if errSecondary != nil {
		return "", errors.New("SYSTEM FAILURE: Both primary and secondary providers failed")
	}
	return respSecondary, nil
}

func (f *FallbackClient) Name() string {
	return fmt.Sprintf("SmartFallback (Pri: %s -> Sec: %s)", f.Primary.Name(), f.Secondary.Name())
}
