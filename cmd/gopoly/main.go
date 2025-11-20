package main

import (
	"context"
	"flag"
	"fmt"
	"gopolyai/pkg/ai"
	"gopolyai/pkg/ai/mock"
	"gopolyai/pkg/ai/ollama"
	"gopolyai/pkg/ai/openai"
	"os"
	"strings"
	"time"
)

func main() {
	providerType := flag.String("p", "mock", "AI Sağlayıcısı: openai, ollama, mock, auto")
	modelName := flag.String("m", "", "Model ismi (örn: gpt-4o, llama3)")
	apiKey := flag.String("k", "", "API Key (Sadece OpenAI için gerekli, opsiyonel)")

	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Kullanım: gopoly -p [provider] \"Sorunuz buraya\"")
		os.Exit(1)
	}
	prompt := strings.Join(args, " ")

	var aiClient ai.AIProvider

	finalAPIKey := *apiKey
	if finalAPIKey == "" {
		finalAPIKey = os.Getenv("OPENAI_API_KEY")
	}

	switch *providerType {
	case "openai":
		client := openai.NewClient(finalAPIKey)

		if *modelName != "" {
			client.Configure(ai.Config{ModelName: *modelName})
		}
		aiClient = client

	case "ollama":
		client := ollama.NewClient()
		if *modelName != "" {
			client.Configure(ai.Config{ModelName: *modelName})
		}
		aiClient = client

	case "mock":
		aiClient = mock.NewClient("Bu bir test cevabıdır.", false)

	case "auto":
		p1 := openai.NewClient(finalAPIKey)
		p2 := mock.NewClient("OpenAI çöktü, Mock devrede.", false)
		aiClient = ai.NewFallbackClient(p1, p2)

	default:
		fmt.Printf("Hata: Bilinmeyen sağlayıcı '%s'\n", *providerType)
		os.Exit(1)
	}

	runAgent(aiClient, prompt)
}

func runAgent(p ai.AIProvider, prompt string) {
	fmt.Printf("--- [%s] Çalışıyor ---\n", p.Name())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := p.Generate(ctx, prompt)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("\n[HATA] İşlem başarısız: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n>> CEVAP:\n%s\n", resp)
	fmt.Printf("\n(Süre: %v)\n", duration)
}
