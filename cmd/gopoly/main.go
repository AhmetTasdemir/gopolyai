package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/anthropic"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/google"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/ollama"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/openai"
)

func main() {

	provider := flag.String("p", "ollama", "AI Sağlayıcısı: openai, google, anthropic, ollama")
	apiKey := flag.String("k", os.Getenv("AI_API_KEY"), "API Key (Env: AI_API_KEY)")

	modelName := flag.String("m", "", "Model ismi (örn: gpt-4o, gemini-1.5-flash, llama3)")

	flag.Parse()

	prompt := "Go dili neden hızlıdır? Kısa bir cümleyle açıkla."
	if len(flag.Args()) > 0 {
		prompt = flag.Args()[0]
	}

	var client ai.AIProvider

	if *apiKey == "" && *provider != "ollama" {
		fmt.Println("UYARI: API Key (-k) veya AI_API_KEY ortam değişkeni belirtilmedi.")
	}

	switch *provider {
	case "openai":
		client = openai.NewClient(*apiKey)
	case "google":
		client = google.NewClient(*apiKey)
	case "anthropic":
		client = anthropic.NewClient(*apiKey)
	case "ollama":
		client = ollama.NewClient()
	default:
		log.Fatalf("Bilinmeyen sağlayıcı: %s", *provider)
	}

	cfg := ai.Config{
		Temperature: 0.7,
	}
	if *modelName != "" {
		cfg.ModelName = *modelName
	}

	client.Configure(cfg)

	fmt.Printf("--- %s Kullanılıyor ---\n", client.Name())
	resp, err := client.Generate(context.Background(), prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(">>", resp)
}
