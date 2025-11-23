package ai_test

import (
	"context"
	"testing"

	// DÜZELTME: Import yolları tam modül adı olmalı (Küçük harfe dikkat)
	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/mock"
)

// TestPolymorphism, interface'in yeni Request/Response yapısıyla çalışmasını test eder.
func TestPolymorphism(t *testing.T) {
	// Senaryo: Başarılı Cevap
	fakeAI := mock.NewClient("Her şey yolunda", false)

	// Request Struct Oluştur
	req := ai.ChatRequest{
		Model: "test-model",
		Messages: []ai.ChatMessage{
			{
				Role: "user",
				Content: []ai.Content{
					{Type: "text", Text: "Merhaba"},
				},
			},
		},
	}

	// Generate çağır
	resp, err := fakeAI.Generate(context.Background(), req)

	if err != nil {
		t.Errorf("Hata beklemiyorduk ama hata aldık: %v", err)
	}

	// Struct cevabını kontrol et
	expected := "MOCK: Her şey yolunda"
	if resp.Content != expected {
		t.Errorf("Yanlış cevap.\nBeklenen: %s\nGelen: %s", expected, resp.Content)
	}

	// Metadata testi
	if resp.Usage.TotalTokens != 35 {
		t.Errorf("Token sayımı yanlış. Beklenen: 35, Gelen: %d", resp.Usage.TotalTokens)
	}
}

// TestMultipleProviders, polimorfik dizide gezmeyi test eder.
func TestMultipleProviders(t *testing.T) {
	providers := []ai.AIProvider{
		mock.NewClient("Bot 1", false),
		mock.NewClient("Bot 2", false),
	}

	for i, p := range providers {
		if p.Name() != "Mock AI" {
			t.Errorf("Provider %d ismi yanlış: %s", i, p.Name())
		}
	}
}
