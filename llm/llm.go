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

const (
	DefaultBaseURL   = "http://127.0.0.1:1234/v1"
	DefaultModelName = "qwen3.5-0.8b"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	ModelName  string
}

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
		BaseURL:   baseURL,
		ModelName: modelName,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func (c *Client) Generate(ctx context.Context, messages []Message) (string, error) {
	payload := map[string]interface{}{
		"model":       c.ModelName,
		"messages":    messages,
		"temperature": 0.7,
		"max_tokens":  300,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call LM Studio (is it running?): %w", err)
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("read LLM response body: %w", readErr)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LM Studio returned error %d: %s", resp.StatusCode, string(body))
	}

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
