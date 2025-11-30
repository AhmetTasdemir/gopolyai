package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/logger"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/mock"
)

type CapturingLogger struct {
	LastEntry logger.LogEntry
	CallCount int
}

func (c *CapturingLogger) Log(ctx context.Context, entry logger.LogEntry) {
	c.LastEntry = entry
	c.CallCount++
}

func TestLoggingMiddleware_Generate(t *testing.T) {
	mockProvider := mock.NewClient("Test Response", false)
	capture := &CapturingLogger{}

	priced := NewCostEstimator(mockProvider)

	mw := NewLoggingMiddleware(priced, capture, logger.Config{LogPayloads: true})

	req := ai.ChatRequest{Messages: []ai.ChatMessage{{Role: "user"}}}

	_, err := mw.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if capture.CallCount != 1 {
		t.Errorf("Log not called. Count: %d", capture.CallCount)
	}
	if capture.LastEntry.Operation != "Generate" {
		t.Errorf("Wrong operation name: %s", capture.LastEntry.Operation)
	}
	if capture.LastEntry.CostUSD <= 0 {
		t.Error("Cost not logged (CostEstimator integration error)")
	}
	if capture.LastEntry.ResponsePayload != "MOCK: Test Response" {
		t.Errorf("Payload not logged: %s", capture.LastEntry.ResponsePayload)
	}
}

func TestLoggingMiddleware_Stream(t *testing.T) {

	mockGen := &mockStreamProvider{}
	capture := &CapturingLogger{}
	mw := NewLoggingMiddleware(mockGen, capture, logger.Config{LogPayloads: true})

	stream, _ := mw.GenerateStream(context.Background(), ai.ChatRequest{Model: "test-stream"})

	for range stream {
	}

	time.Sleep(50 * time.Millisecond)

	if capture.CallCount != 1 {
		t.Errorf("Log not created at end of stream")
	}
	if capture.LastEntry.Operation != "GenerateStream" {
		t.Errorf("Stream operation name wrong")
	}
	if capture.LastEntry.ResponsePayload != "Hello World" {
		t.Errorf("Stream content not merged: %s", capture.LastEntry.ResponsePayload)
	}
}

type mockStreamProvider struct{}

func (m *mockStreamProvider) Configure(cfg ai.Config) error { return nil }
func (m *mockStreamProvider) Name() string                  { return "MockStream" }
func (m *mockStreamProvider) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	return nil, nil
}
func (m *mockStreamProvider) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 2)
	go func() {
		ch <- ai.StreamResponse{Chunk: "Hello "}
		ch <- ai.StreamResponse{Chunk: "World", Usage: &ai.TokenUsage{TotalTokens: 10, CostUSD: 0.002}}
		close(ch)
	}()
	return ch, nil
}
