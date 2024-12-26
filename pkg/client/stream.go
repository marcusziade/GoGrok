package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"GoGrok/pkg/types"
)

// StreamHandler defines the interface for handling stream events
type StreamHandler interface {
	OnContent(content string)
	OnError(err error)
	OnComplete()
}

// Client represents the Grok API client
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new Grok API client
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		baseURL:    "https://api.x.ai/v1",
		httpClient: &http.Client{},
	}
}

// StreamChat initiates a streaming chat completion request
func (c *Client) StreamChat(req types.ChatRequest, handler StreamHandler) error {
	if !req.Stream {
		req.Stream = true // Ensure streaming is enabled
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			handler.OnError(fmt.Errorf("failed to read stream: %w", err))
			return err
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			handler.OnComplete()
			break
		}

		var chunk types.StreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			handler.OnError(fmt.Errorf("failed to unmarshal chunk: %w", err))
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			handler.OnContent(chunk.Choices[0].Delta.Content)
		}
	}

	return nil
}
