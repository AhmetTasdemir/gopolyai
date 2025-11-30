package ai_test

import (
	"context"
	"testing"

	// FIX: Import paths must be full module name (Watch for lowercase)
	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/mock"
)

// TestPolymorphism tests if the interface works with the new Request/Response structure.
func TestPolymorphism(t *testing.T) {
	// Scenario: Successful Response
	fakeAI := mock.NewClient("Everything is fine", false)

	// Create Request Struct
	req := ai.ChatRequest{
		Model: "test-model",
		Messages: []ai.ChatMessage{
			{
				Role: "user",
				Content: []ai.Content{
					{Type: "text", Text: "Hello"},
				},
			},
		},
	}

	// Call Generate
	resp, err := fakeAI.Generate(context.Background(), req)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Check Struct response
	expected := "MOCK: Everything is fine"
	if resp.Content != expected {
		t.Errorf("Wrong answer.\nExpected: %s\nGot: %s", expected, resp.Content)
	}

	// Metadata test
	if resp.Usage.TotalTokens != 35 {
		t.Errorf("Token count wrong. Expected: 35, Got: %d", resp.Usage.TotalTokens)
	}
}

// TestMultipleProviders tests iterating through polymorphic array.
func TestMultipleProviders(t *testing.T) {
	providers := []ai.AIProvider{
		mock.NewClient("Bot 1", false),
		mock.NewClient("Bot 2", false),
	}

	for i, p := range providers {
		if p.Name() != "Mock AI" {
			t.Errorf("Provider %d name wrong: %s", i, p.Name())
		}
	}
}
