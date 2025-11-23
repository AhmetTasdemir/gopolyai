package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/anthropic"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/google"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/middleware"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/ollama"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/openai"
)

func main() {

	provider := flag.String("p", "ollama", "AI Sağlayıcısı")
	apiKey := flag.String("k", os.Getenv("AI_API_KEY"), "API Key")
	modelName := flag.String("m", "", "Model ismi")
	streamMode := flag.Bool("s", false, "Streaming (Daktilo) modunu aç")

	flag.Parse()

	prompt := "Go dili neden hızlıdır?"
	if len(flag.Args()) > 0 {
		prompt = flag.Args()[0]
	}

	var baseClient ai.AIProvider
	switch *provider {
	case "openai":
		baseClient = openai.NewClient(*apiKey)
	case "google":
		baseClient = google.NewClient(*apiKey)
	case "anthropic":
		baseClient = anthropic.NewClient(*apiKey)
	case "ollama":
		baseClient = ollama.NewClient()
	default:
		log.Fatalf("❌ Bilinmeyen sağlayıcı: %s", *provider)
	}

	cfg := ai.Config{Temperature: 0.7}
	if *modelName != "" {
		cfg.ModelName = *modelName
	}
	baseClient.Configure(cfg)

	pricedClient := middleware.NewCostEstimator(baseClient)
	retryClient := middleware.NewResilientClient(pricedClient, middleware.RetryConfig{
		MaxRetries: 3, BaseDelay: 1 * time.Second, MaxDelay: 5 * time.Second,
	})
	finalClient := middleware.NewCircuitBreaker(retryClient, 3, 30*time.Second)

	fmt.Printf("--- 🧠 %s Kullanılıyor ---\n", finalClient.Name())

	req := ai.ChatRequest{
		Model: *modelName,
		Messages: []ai.ChatMessage{
			{Role: "user", Content: []ai.Content{{Type: "text", Text: prompt}}},
		},
		Temperature: 0.7,
	}

	start := time.Now()
	var finalUsage ai.TokenUsage
	if *streamMode {
		fmt.Println(">> CEVAP (Streaming):")

		streamChan, err := finalClient.GenerateStream(context.Background(), req)
		if err != nil {
			log.Fatalf("HATA: %v", err)
		}

		for packet := range streamChan {
			if packet.Err != nil {
				fmt.Printf("\n❌ Stream Hatası: %v\n", packet.Err)
				break
			}

			fmt.Print(packet.Chunk)
			if packet.Usage != nil {
				finalUsage = *packet.Usage
			}
		}
		fmt.Println()

	} else {
		resp, err := finalClient.Generate(context.Background(), req)
		if err != nil {
			log.Fatalf("HATA: %v", err)
		}
		fmt.Println(">> CEVAP:", resp.Content)
		finalUsage = resp.Usage
	}

	fmt.Println("\n------------------------------------------------")
	fmt.Printf("⏱️: %v\n", time.Since(start))

	if finalUsage.TotalTokens > 0 {
		fmt.Printf("🔢 Token: %d (Girdi: %d / Çıktı: %d)\n",
			finalUsage.TotalTokens,
			finalUsage.InputTokens,
			finalUsage.OutputTokens)

		if finalUsage.CostUSD > 0 {
			fmt.Printf("💰: $%.6f\n", finalUsage.CostUSD)
		} else {
			fmt.Printf("💰: $0.000000 (Local/Ücretsiz)\n")
		}
	}
}
