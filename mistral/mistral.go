package mistral

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type MistralClient struct {
	apiKey       string
	httpClient   *http.Client
	baseURL      string
	systemPrompt string
}

func NewMistralClient(apiKey string, systemPromptPath string) *MistralClient {
	// open and read the system prompt from the specified file path
	content, err := os.ReadFile(systemPromptPath)
	if err != nil {
		log.Fatalf("MistralClient: Error reading system prompt file: %v", err)
	}

	return &MistralClient{
		apiKey:  apiKey,
		baseURL: "https://api.mistral.ai",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		systemPrompt: string(content),
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
}

type MessageResponse struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Choice struct {
	Index        int             `json:"index"`
	Message      MessageResponse `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type MistralResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

func (mc *MistralClient) postPrompt(message string) (*MistralResponse, error) {
	requestBody, err := mc.formatRequestBody(message)
	if err != nil {
		return nil, fmt.Errorf("failed to format request body: %w", err)
	}

	req, err := http.NewRequest("POST", mc.baseURL+"/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+mc.apiKey)

	resp, err := mc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var mistralResp MistralResponse
	if err := json.Unmarshal(respBody, &mistralResp); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return &mistralResp, nil
}

func (mc *MistralClient) GetGeneratedTitles(message string) string {
	resp, err := mc.postPrompt(message)
	if err != nil {
		log.Printf("MistralClient: Error generating titles: %v", err)
		return ""
	}

	if resp == nil || len(resp.Choices) == 0 {
		return ""
	}

	// The content might contain multiple titles separated by newlines
	content := resp.Choices[0].Message.Content
	log.Printf("Mistral: Original title: %s", message)
	log.Printf("Mistral: Generated title: %s", content)
	return content
}

func (mc *MistralClient) formatRequestBody(message string) ([]byte, error) {
	request := ChatCompletionRequest{
		Model: "mistral-large-latest",
		Messages: []Message{
			{
				Role:    "system",
				Content: mc.systemPrompt,
			},
			{
				Role:    "user",
				Content: message,
			},
		},
		Temperature: 0.9,
	}

	return json.Marshal(request)
}
