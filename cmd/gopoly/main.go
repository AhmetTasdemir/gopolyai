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
	provider := flag.String("p", "ollama", "AI Provider")
	apiKey := flag.String("k", os.Getenv("AI_API_KEY"), "API Key")
	modelName := flag.String("m", "", "Model name")
	streamMode := flag.Bool("s", false, "Turn on streaming mode")
	structMode := flag.Bool("struct", false, "Turn on structured output mode")
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
		fmt.Printf(">> Rate Limiter Active: %d req/s\n", *rateLimit)
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
	breakerClient := middleware.NewCircuitBreaker(loggedClient, 3, 30*time.Second)

	finalClient := middleware.NewTracingMiddleware(breakerClient)

	fmt.Printf("--- 🧠 %s Initializing ---\n", finalClient.Name())

	req := ai.ChatRequest{
		Model: *modelName,
		Messages: []ai.ChatMessage{
			{Role: "user", Content: []ai.Content{{Type: "text", Text: prompt}}},
		},
		Temperature: 0.7,
	}

	start := time.Now()

	if *streamMode {
		fmt.Println(">> STREAM MODE ACTIVE...")
		streamChan, err := finalClient.GenerateStream(context.Background(), req)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}

		for packet := range streamChan {
			if packet.Err != nil {
				fmt.Printf("\n Stream Error: %v\n", packet.Err)
				break
			}
			fmt.Print(packet.Chunk)
		}
		fmt.Println()
	} else if *structMode {
		fmt.Println(">> STRUCTURED MODE ACTIVE...")
		type ResponseStruct struct {
			Answer   string   `json:"answer" description:"The direct answer to the question"`
			Keywords []string `json:"keywords" description:"List of important keywords related to the answer"`
			IsCode   bool     `json:"is_code" description:"Whether the answer contains code snippets"`
		}

		var target ResponseStruct
		err := ai.GenerateStruct(context.Background(), finalClient, req, &target)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		fmt.Printf(">> RESPONSE STRUCT:\n%+v\n", target)
	} else {
		resp, err := finalClient.Generate(context.Background(), req)
		if err != nil {
			log.Fatalf("ERROR: %v", err)
		}
		fmt.Println(">> RESPONSE:", resp.Content)
	}

	fmt.Printf("\n------------------------------------------------")
	fmt.Printf("\n Client Duration: %v (Logs might be written asynchronously in the background)\n", time.Since(start))

	time.Sleep(100 * time.Millisecond)
}
