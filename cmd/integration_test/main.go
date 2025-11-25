package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/logger"
	"github.com/ahmettasdemir/gopolyai/pkg/ai/middleware"
)

type MockProvider struct{}

func (m *MockProvider) Configure(cfg ai.Config) error { return nil }
func (m *MockProvider) Name() string                  { return "Mock-GPT-4" }

func (m *MockProvider) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	// Yapay bir gecikme ekleyelim
	time.Sleep(100 * time.Millisecond)
	return &ai.ChatResponse{
		Content: "Bu, Unary (Tekli) istek için test cevabıdır.",
		Usage: ai.TokenUsage{
			InputTokens:  50,
			OutputTokens: 20,
			TotalTokens:  70,
		},
	}, nil
}

func (m *MockProvider) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	ch := make(chan ai.StreamResponse, 5)
	go func() {
		defer close(ch)
		chunks := []string{"Bu ", "bir ", "streaming ", "test ", "cevabıdır."}

		for _, chunk := range chunks {
			time.Sleep(50 * time.Millisecond) // Her kelime arası bekleme
			ch <- ai.StreamResponse{Chunk: chunk}
		}

		// Son pakette kullanım bilgisini gönder
		ch <- ai.StreamResponse{
			Usage: &ai.TokenUsage{
				InputTokens:  10,
				OutputTokens: 5,
				TotalTokens:  15,
			},
		}
	}()
	return ch, nil
}

// === 2. SPY LOGGER (Casus Logger) ===
// Logları konsola basan basit bir implementasyon.
type ConsoleLogger struct{}

func (c *ConsoleLogger) Log(ctx context.Context, entry logger.LogEntry) {
	// Logların asenkron geldiğini görmek için renklendirme ve formatlama yapıyoruz
	fmt.Printf("\n\n📝 [LOGGER YAKALADI] -------------------------------------\n")
	fmt.Printf("   🆔 Operasyon : %s\n", entry.Operation)
	fmt.Printf("   ⏱️  Süre      : %v\n", entry.Duration)
	fmt.Printf("   🤖 Model     : %s (%s)\n", entry.Model, entry.Provider)
	fmt.Printf("   💰 Maliyet   : $%.6f\n", entry.CostUSD)
	fmt.Printf("   🔢 Token     : %d (In) / %d (Out)\n", entry.InputTokens, entry.OutputTokens)

	if entry.RequestPayload != "" {
		fmt.Printf("   📩 Request   : %s\n", entry.RequestPayload)
	}
	if entry.ResponsePayload != "" {
		fmt.Printf("   📨 Response  : %s\n", entry.ResponsePayload)
	}
	fmt.Printf("------------------------------------------------------------\n")
}

// === 3. TEST SENARYOSU ===
func main() {
	fmt.Println("🚀 GoPolyAI Logger Entegrasyon Testi Başlıyor...")

	// A. Zinciri Kuruyoruz
	mockAI := &MockProvider{}

	// 1. Önce Maliyet Hesaplayıcı (Token -> Dolar)
	costMW := middleware.NewCostEstimator(mockAI)

	// 2. Sonra Logger (Maliyeti de loglasın diye dışta)
	myLogger := &ConsoleLogger{}
	logConfig := logger.Config{LogPayloads: true, LogErrorsOnly: false}

	// Test edilen asıl eleman:
	finalClient := middleware.NewLoggingMiddleware(costMW, myLogger, logConfig)

	ctx := context.Background()

	// --- SENARYO 1: Normal İstek (Generate) ---
	fmt.Println("\n--- [TEST 1] Unary Request Başlatılıyor... ---")
	start := time.Now()
	resp, err := finalClient.Generate(ctx, ai.ChatRequest{
		Model:    "gpt-4o", // Pahalı model seçelim ki maliyet hesaplansın
		Messages: []ai.ChatMessage{{Role: "user", Content: []ai.Content{{Type: "text", Text: "Merhaba"}}}},
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✅ İstemci Cevabı Aldı: %s (Süre: %v)\n", resp.Content, time.Since(start))

	// Logların asenkron yazılması için biraz bekleyelim
	time.Sleep(200 * time.Millisecond)

	// --- SENARYO 2: Akış İsteği (Streaming) ---
	fmt.Println("\n--- [TEST 2] Streaming Request Başlatılıyor... ---")
	streamChan, err := finalClient.GenerateStream(ctx, ai.ChatRequest{
		Model:    "gpt-3.5-turbo",
		Messages: []ai.ChatMessage{{Role: "user", Content: []ai.Content{{Type: "text", Text: "Bana hikaye anlat"}}}},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print(">> Stream Akıyor: ")
	for packet := range streamChan {
		fmt.Print(packet.Chunk)
	}
	fmt.Println("\n✅ Stream Bitti.")

	// Logların asenkron yazılması için son bir bekleme
	fmt.Println("⏳ Logların yazılması bekleniyor...")
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n🏁 Test Tamamlandı.")
}
