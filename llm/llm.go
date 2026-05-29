package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// =============================================
// FILE PURPOSE
// This file communicates with the chat model (LLM) to generate answers.
// =============================================

const (
	DefaultBaseURL     = "http://127.0.0.1:1234/v1"
	DefaultModelName   = "qwen3.5-0.8b"
	DefaultTemperature = 0.7
	DefaultMaxTokens   = 300
)

// =============================================
// CLIENT STRUCTURE
// =============================================

// Client holds connection to chat model
type Client struct {
	BaseURL     string
	HTTPClient  *http.Client
	ModelName   string
	Temperature float64
	MaxTokens   int
}

// NewClient creates LLM client
// What it does: Reads config and prepares client
func NewClient(modelName string) *Client {
	baseURL := strings.TrimSpace(os.Getenv("LM_STUDIO_BASE_URL"))
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		modelName = strings.TrimSpace(os.Getenv("LM_STUDIO_CHAT_MODEL"))
	}
	if modelName == "" {
		modelName = DefaultModelName
	}

	return &Client{
		BaseURL:     baseURL,
		ModelName:   modelName,
		Temperature: DefaultTemperature,
		MaxTokens:   DefaultMaxTokens,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // Longer timeout for chat
		},
	}
}

// =============================================
// DATA STRUCTURE
// =============================================

// Message represents one turn in the conversation
type Message struct {
	Role    string `json:"role"`    // "system" or "user"
	Content string `json:"content"` // The message text
}

// =============================================
// MAIN FUNCTION
// =============================================

// Generate sends messages to LLM and returns the answer text
// What it does: Calls LM Studio chat API
// Why: This is where the actual AI answer is generated
func (c *Client) Generate(ctx context.Context, messages []Message) (string, error) {
	// 1. Prepare payload
	payload := map[string]interface{}{
		"model":       c.ModelName,
		"messages":    messages,
		"temperature": c.Temperature,
		"max_tokens":  c.MaxTokens,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// 2. Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// 3. Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call LM Studio (is it running?): %w", err)
	}
	defer resp.Body.Close()

	// 4. Read response body
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("read LLM response body: %w", readErr)
	}

	// 5. Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LM Studio returned error %d: %s", resp.StatusCode, string(body))
	}

	// 6. Parse JSON response
	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response generated from model")
	}

	return result.Choices[0].Message.Content, nil
}