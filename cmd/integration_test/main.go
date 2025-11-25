package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/anthropic"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/google"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/logger"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/middleware"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/ollama"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/openai"
)

type SimpleJSONLogger struct{}

func (s *SimpleJSONLogger) Log(ctx context.Context, entry logger.LogEntry) {
	data, _ := json.MarshalIndent(entry, "", "  ")
	fmt.Printf("\n📝 [TELEMETRY LOG] >>\n%s\n", string(data))
}

func main() {
	provider := flag.String("p", "ollama", "AI Sağlayıcısı")
	apiKey := flag.String("k", os.Getenv("AI_API_KEY"), "API Key")
	modelName := flag.String("m", "", "Model ismi")
	streamMode := flag.Bool("s", false, "Turn on streaming mode")
	rateLimit := flag.Int("rate-limit", 0, "Rate limit (requests per second). 0 = unlimited")
	flag.Parse()

	prompt := "What is an interface in Go?"
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
		log.Fatalf("Unknown provider: %s", *provider)
	}

	cfg := ai.Config{Temperature: 0.7}
	if *modelName != "" {
		cfg.ModelName = *modelName
	}
	baseClient.Configure(cfg)

	pricedClient := middleware.NewCostEstimator(baseClient)

	var rateLimitedClient ai.AIProvider = pricedClient
	if *rateLimit > 0 {
		fmt.Printf(">> Rate Limiter Aktif: %d req/s\n", *rateLimit)
		rateLimitedClient = middleware.NewRateLimiterMiddleware(pricedClient, *rateLimit, *rateLimit)
	}

	retryClient := middleware.NewResilientClient(rateLimitedClient, middleware.RetryConfig{
		MaxRetries: 2, BaseDelay: 1 * time.Second, MaxDelay: 3 * time.Second,
	})

	myLogger := &SimpleJSONLogger{}
	logConfig := logger.Config{
		LogPayloads:   true,
		LogErrorsOnly: false,
	}
	loggedClient := middleware.NewLoggingMiddleware(retryClient, myLogger, logConfig)
	finalClient := middleware.NewCircuitBreaker(loggedClient, 3, 30*time.Second)

	fmt.Printf("--- 🧠 %s Başlatılıyor ---\n", finalClient.Name())

	req := ai.ChatRequest{
		Model: *modelName,
		Messages: []ai.ChatMessage{
			{Role: "user", Content: []ai.Content{{Type: "text", Text: prompt}}},
		},
		Temperature: 0.7,
	}

	start := time.Now()

	if *streamMode {
		fmt.Println(">> STREAM MODU AKTİF...")
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
		}
		fmt.Println()
	} else {
		resp, err := finalClient.Generate(context.Background(), req)
		if err != nil {
			log.Fatalf("HATA: %v", err)
		}
		fmt.Println(">> CEVAP:", resp.Content)
	}

	fmt.Printf("\n------------------------------------------------")
	fmt.Printf("\n⏱️ İstemci Süresi: %v (Loglar asenkron arkada yazılıyor olabilir)\n", time.Since(start))

	time.Sleep(100 * time.Millisecond)
}
