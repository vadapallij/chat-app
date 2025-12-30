package services

import (
	"bytes"
	"chat-app/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type VLLMService struct {
	baseURL string
	client  *http.Client
}

type VLLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type VLLMRequest struct {
	Model    string        `json:"model"`
	Messages []VLLMMessage `json:"messages"`
	MaxTokens int          `json:"max_tokens"`
	Temperature float64    `json:"temperature"`
}

type VLLMResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func NewVLLMService(baseURL string) *VLLMService {
	return &VLLMService{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (s *VLLMService) Chat(messages []models.Message, userMessage string) (string, error) {
	// Convert message history to vLLM format
	vllmMessages := make([]VLLMMessage, 0, len(messages)+1)

	// Add conversation history
	for _, msg := range messages {
		vllmMessages = append(vllmMessages, VLLMMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// Add new user message
	vllmMessages = append(vllmMessages, VLLMMessage{
		Role:    "user",
		Content: userMessage,
	})

	// Prepare request
	reqBody := VLLMRequest{
		Model:       "meta-llama/Meta-Llama-3.1-8B-Instruct",
		Messages:    vllmMessages,
		MaxTokens:   4096,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request to vLLM
	url := fmt.Sprintf("%s/v1/chat/completions", s.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request to vLLM: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vLLM API error (status %d): %s", resp.StatusCode, string(body))
	}

	var vllmResp VLLMResponse
	if err := json.Unmarshal(body, &vllmResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(vllmResp.Choices) == 0 {
		return "", fmt.Errorf("no response from vLLM")
	}

	return vllmResp.Choices[0].Message.Content, nil
}
