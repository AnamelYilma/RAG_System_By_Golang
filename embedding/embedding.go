package embedding

import (
	"MyRagByCivic/chunker"
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
	DefaultModelName = "text-embedding-nomic-embed-text-v1.5"
)

// Embedding represents one vector for one chunk
type Embedding struct {
	Chunk     Chunk     // From your chunker
	Vector    []float32 // The actual numbers (vector)
	ModelName string    // Which model we used
}

// Chunk reuses the chunker package type so callers don't need conversions.
type Chunk = chunker.Chunk

// Client holds connection to LM Studio
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	ModelName  string
}

// NewClient creates a new embedding client
// Goal: Prepare connection to local LM Studio
func NewClient(modelName string) *Client {
	baseURL := strings.TrimSpace(os.Getenv("LM_STUDIO_BASE_URL"))
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	modelName = strings.TrimSpace(modelName)
	if modelName == "" {
		modelName = strings.TrimSpace(os.Getenv("LM_STUDIO_EMBEDDING_MODEL"))
	}
	if modelName == "" {
		modelName = DefaultModelName
	}

	return &Client{
		BaseURL:   baseURL,
		ModelName: modelName,
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second, // Give time for embedding
		},
	}
}

// GetEmbedding gets vector for one text
// Goal: Send one chunk text → receive vector
func (c *Client) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	payload := map[string]interface{}{
		"model": c.ModelName,
		"input": text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call LM Studio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = http.StatusText(resp.StatusCode)
		}

		return nil, fmt.Errorf(
			"LM Studio embeddings request failed for model %q at %s: status %d: %s",
			c.ModelName,
			c.BaseURL,
			resp.StatusCode,
			message,
		)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}

	return result.Data[0].Embedding, nil
}

// GetEmbeddingsForChunks processes many chunks
// Goal: Take list of chunks → return list of embeddings
func (c *Client) GetEmbeddingsForChunks(ctx context.Context, chunks []Chunk) ([]Embedding, error) {
	var embeddings []Embedding

	for i, chunk := range chunks {
		fmt.Printf("Creating embedding %d/%d: %s...\n", i+1, len(chunks), chunk.FileName)

		vector, err := c.GetEmbedding(ctx, chunk.Text)
		if err != nil {
			return nil, fmt.Errorf("failed at chunk %d: %w", i, err)
		}

		embeddings = append(embeddings, Embedding{
			Chunk:     chunk,
			Vector:    vector,
			ModelName: c.ModelName,
		})

		// Small delay to not overload your PC
		time.Sleep(100 * time.Millisecond)
	}

	return embeddings, nil
}
