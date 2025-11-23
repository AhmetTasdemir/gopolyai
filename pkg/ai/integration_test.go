package ai_test

import (
	"context"
	"testing"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/anthropic"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/google"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/openai"
)

func TestOpenAIIntegration(t *testing.T) {
	// Start mock server
	server := openai.StartMockServer()
	defer server.Close()

	// Configure client
	client := openai.NewClient("test-key")
	err := client.Configure(ai.Config{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to configure client: %v", err)
	}

	// Make request
	req := ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "user", Content: []ai.Content{{Type: "text", Text: "Hello"}}},
		},
	}

	resp, err := client.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "This is a mock response from OpenAI." {
		t.Errorf("Unexpected response: %s", resp.Content)
	}
}

func TestGoogleIntegration(t *testing.T) {
	// Start mock server
	server := google.StartMockServer()
	defer server.Close()

	// Configure client
	client := google.NewClient("test-key")
	// Google client expects a format string for BaseURL
	err := client.Configure(ai.Config{
		BaseURL: server.URL + "/models/%s:generateContent?key=%s",
	})
	if err != nil {
		t.Fatalf("Failed to configure client: %v", err)
	}

	// Make request
	req := ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "user", Content: []ai.Content{{Type: "text", Text: "Hello"}}},
		},
	}

	resp, err := client.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "This is a mock response from Google Gemini." {
		t.Errorf("Unexpected response: %s", resp.Content)
	}
}

func TestAnthropicIntegration(t *testing.T) {
	// Start mock server
	server := anthropic.StartMockServer()
	defer server.Close()

	// Configure client
	client := anthropic.NewClient("test-key")
	err := client.Configure(ai.Config{
		BaseURL: server.URL,
	})
	if err != nil {
		t.Fatalf("Failed to configure client: %v", err)
	}

	// Make request
	req := ai.ChatRequest{
		Messages: []ai.ChatMessage{
			{Role: "user", Content: []ai.Content{{Type: "text", Text: "Hello"}}},
		},
	}

	resp, err := client.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if resp.Content != "This is a mock response from Anthropic Claude." {
		t.Errorf("Unexpected response: %s", resp.Content)
	}
}
