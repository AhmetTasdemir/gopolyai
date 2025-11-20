package ai_test

import (
	"context"
	"gopolyai/pkg/ai"
	"gopolyai/pkg/ai/mock"
	"testing"
)

func TestPolymorphism(t *testing.T) {
	fakeAI := mock.NewClient("Her şey yolunda", false)

	response, err := fakeAI.Generate(context.Background(), "Merhaba")

	if err != nil {
		t.Errorf("Hata beklemiyorduk ama hata aldık: %v", err)
	}

	if response != "MOCK CEVAP: Her şey yolunda" {
		t.Errorf("Yanlış cevap: %s", response)
	}
}

func TestMultipleProviders(t *testing.T) {
	providers := []ai.AIProvider{
		mock.NewClient("Bot 1", false),
		mock.NewClient("Bot 2", false),
	}

	for i, p := range providers {
		if p.Name() != "Mock AI (Testing)" {
			t.Errorf("Provider %d ismi yanlış", i)
		}
	}
}
