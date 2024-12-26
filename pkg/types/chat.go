package types

// Message represents a chat message with role and content
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents the request body for chat completions
type ChatRequest struct {
	Messages    []Message `json:"messages"`
	Model       string    `json:"model"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Delta represents a chunk of the streaming response
type Delta struct {
	Content string `json:"content,omitempty"`
	Role    string `json:"role,omitempty"`
}

// Choice represents a completion choice in the streaming response
type Choice struct {
	Index int   `json:"index"`
	Delta Delta `json:"delta"`
}

// StreamChunk represents a single chunk in the SSE stream
type StreamChunk struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	SystemFingerprint string   `json:"system_fingerprint"`
}
