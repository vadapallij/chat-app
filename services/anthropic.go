package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"chat-app/models"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

// AnthropicService handles communication with the Anthropic API
type AnthropicService struct {
	apiKey string
	client *http.Client
}

// NewAnthropicService creates a new Anthropic service
func NewAnthropicService(apiKey string) *AnthropicService {
	return &AnthropicService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

// AnthropicMessage represents a message in the Anthropic API format
type AnthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AnthropicRequest represents a request to the Anthropic API
type AnthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []AnthropicMessage `json:"messages"`
}

// AnthropicResponse represents a response from the Anthropic API
type AnthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Chat sends a message to Claude and returns the response
func (s *AnthropicService) Chat(messages []models.Message, userMessage string) (string, error) {
	// Convert messages to Anthropic format
	var anthropicMessages []AnthropicMessage
	for _, msg := range messages {
		anthropicMessages = append(anthropicMessages, AnthropicMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	// Add the new user message
	anthropicMessages = append(anthropicMessages, AnthropicMessage{
		Role:    "user",
		Content: userMessage,
	})

	reqBody := AnthropicRequest{
		Model:     "claude-sonnet-4-20250514",
		MaxTokens: 4096,
		Messages:  anthropicMessages,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", anthropicAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var anthropicResp AnthropicResponse
	if err := json.Unmarshal(body, &anthropicResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if anthropicResp.Error != nil {
		return "", fmt.Errorf("anthropic API error: %s", anthropicResp.Error.Message)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Anthropic")
	}

	return anthropicResp.Content[0].Text, nil
}
