package anthropic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"
)

func StartMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			http.Error(w, "Missing x-api-key header", http.StatusUnauthorized)
			return
		}

		time.Sleep(10 * time.Millisecond)

		response := struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
			Usage struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			} `json:"usage"`
		}{
			Content: []struct {
				Text string `json:"text"`
			}{
				{Text: "This is a mock response from Anthropic Claude."},
			},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{
				InputTokens:  10,
				OutputTokens: 20,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}
