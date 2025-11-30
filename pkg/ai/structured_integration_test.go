package ai_test

import (
	"context"
	"testing"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/mock"
)

type UserProfile struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestGenerateStructIntegration(t *testing.T) {
	jsonResponse := `{"name": "Alice", "age": 25}`
	mockProvider := mock.NewClient(jsonResponse, false)

	req := ai.ChatRequest{
		Model: "test-model",
		Messages: []ai.ChatMessage{
			{Role: "user", Content: []ai.Content{{Type: "text", Text: "Generate user"}}},
		},
	}

	var target UserProfile
	err := ai.GenerateStruct(context.Background(), mockProvider, req, &target)
	if err != nil {
		t.Fatalf("GenerateStruct failed: %v", err)
	}

	if target.Name != "Alice" {
		t.Errorf("Expected Name 'Alice', got '%s'", target.Name)
	}
	if target.Age != 25 {
		t.Errorf("Expected Age 25, got %d", target.Age)
	}
}
