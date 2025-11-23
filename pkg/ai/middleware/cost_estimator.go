package middleware

import (
	"context"
	"strings"

	"github.com/ahmettasdemir/gopolyai/pkg/ai"
)

type ModelPrice struct {
	InputPrice  float64
	OutputPrice float64
}

var DefaultPricing = map[string]ModelPrice{
	// OpenAI
	"gpt-4o":        {InputPrice: 5.00, OutputPrice: 15.00},
	"gpt-4-turbo":   {InputPrice: 10.00, OutputPrice: 30.00},
	"gpt-3.5-turbo": {InputPrice: 0.50, OutputPrice: 1.50},

	// Anthropic
	"claude-3-5-sonnet": {InputPrice: 3.00, OutputPrice: 15.00},
	"claude-3-opus":     {InputPrice: 15.00, OutputPrice: 75.00},

	// Google
	"gemini-1.5-pro":   {InputPrice: 3.50, OutputPrice: 10.50},
	"gemini-1.5-flash": {InputPrice: 0.35, OutputPrice: 1.05},

	// Local / Ollama
	"tinyllama": {InputPrice: 0.0, OutputPrice: 0.0},
	"llama3":    {InputPrice: 0.0, OutputPrice: 0.0},
}

type CostEstimator struct {
	provider ai.AIProvider
	pricing  map[string]ModelPrice
}

func NewCostEstimator(p ai.AIProvider) *CostEstimator {
	return &CostEstimator{
		provider: p,
		pricing:  DefaultPricing,
	}
}

func (ce *CostEstimator) SetPricing(p map[string]ModelPrice) {
	ce.pricing = p
}

func (ce *CostEstimator) Configure(cfg ai.Config) error {
	return ce.provider.Configure(cfg)
}

func (ce *CostEstimator) Name() string {
	return ce.provider.Name()
}
func (ce *CostEstimator) Generate(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	resp, err := ce.provider.Generate(ctx, req)
	if err != nil {
		return nil, err
	}

	modelName := req.Model
	if modelName == "" {
	}

	price, found := ce.findPrice(modelName)

	if found && (resp.Usage.InputTokens > 0 || resp.Usage.OutputTokens > 0) {
		inputCost := (float64(resp.Usage.InputTokens) / 1_000_000) * price.InputPrice
		outputCost := (float64(resp.Usage.OutputTokens) / 1_000_000) * price.OutputPrice

		totalCost := inputCost + outputCost

		resp.Usage.CostUSD = totalCost
	}

	return resp, nil
}

func (ce *CostEstimator) GenerateStream(ctx context.Context, req ai.ChatRequest) (<-chan ai.StreamResponse, error) {
	originalChan, err := ce.provider.GenerateStream(ctx, req)
	if err != nil {
		return nil, err
	}

	proxyChan := make(chan ai.StreamResponse, 10)
	modelName := req.Model
	price, found := ce.findPrice(modelName)
	go func() {
		defer close(proxyChan)

		for packet := range originalChan {
			if packet.Usage != nil && found {
				inputCost := (float64(packet.Usage.InputTokens) / 1_000_000) * price.InputPrice
				outputCost := (float64(packet.Usage.OutputTokens) / 1_000_000) * price.OutputPrice

				packet.Usage.CostUSD = inputCost + outputCost
			}

			proxyChan <- packet
		}
	}()

	return proxyChan, nil
}

func (ce *CostEstimator) findPrice(model string) (ModelPrice, bool) {

	if p, ok := ce.pricing[model]; ok {
		return p, true
	}
	for key, price := range ce.pricing {
		if strings.Contains(model, key) {
			return price, true
		}
	}

	return ModelPrice{}, false
}
