package openai

type chatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []message `json:"messages"`
	// MaxTokens   int     `json:"max_tokens,omitempty"`
	// Temperature float64 `json:"temperature,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
}
