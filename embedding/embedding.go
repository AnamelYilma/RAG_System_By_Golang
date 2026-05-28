package embedding

import (
	"MyRagByCivic/chunker"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// =============================================
// FILE PURPOSE
// This file converts text chunks into vectors (numbers) by calling LM Studio.
// =============================================

const (
	// Default values if .env is not set
	DefaultBaseURL   = "http://127.0.0.1:1234/v1"
	DefaultModelName = "text-embedding-nomic-embed-text-v1.5"
)

// =============================================
// DATA STRUCTURES
// =============================================

// Embedding holds one chunk + its vector
// What it does: Keeps text and its numerical version together
// Why: Needed for storage and search
type Embedding struct {
	Chunk     Chunk     // Original text chunk
	Vector    []float32 // The numbers (vector)
	ModelName string    // Which model created this vector
}

// Chunk reuses type from chunker (alias)
// Why: Avoids duplicating the same struct
type Chunk = chunker.Chunk

// =============================================
// MAIN CLIENT STRUCTURE
// =============================================

// Client holds everything needed to talk to LM Studio
// What it does: Stores connection settings
// Why: So we can reuse the same client many times
type Client struct {
	BaseURL    string        // API base URL
	HTTPClient *http.Client  // Reusable HTTP client with timeout
	ModelName  string        // Embedding model name
}

// =============================================
// CONSTRUCTOR FUNCTION
// =============================================

// NewClient creates a new embedding client
// What it does + Why: Reads settings from .env and prepares connection
// How: Checks environment variables with fallback to defaults
func NewClient(modelName string) *Client {
	// Read base URL from .env or use default
	baseURL := strings.TrimSpace(os.Getenv("LM_STUDIO_BASE_URL"))
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	// Set model name (priority: parameter > .env > default)
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
			Timeout: 60 * time.Second, // Give enough time for slow models
		},
	}
}

// =============================================
// CORE FUNCTIONS
// =============================================

// GetEmbedding gets vector for one single text
// What it does: Sends text to LM Studio → returns vector numbers
// Why: This is the basic operation for embeddings
func (c *Client) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	// 1. Prepare JSON payload
	payload := map[string]interface{}{
		"model": c.ModelName,
		"input": text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 2. Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	// 3. Send request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call LM Studio: %w", err)
	}
	defer resp.Body.Close()

	// 4. Check response status
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

	// 5. Parse JSON response
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

// GetEmbeddingsForChunks processes many chunks concurrently
// What it does: Takes multiple chunks → returns embeddings
// Why: Much faster than calling one by one
func (c *Client) GetEmbeddingsForChunks(ctx context.Context, chunks []Chunk) ([]Embedding, error) {
	embeddings := make([]Embedding, len(chunks))

	// errgroup = group of goroutines with error handling
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(5) // Limit concurrent requests (safe for weak PC)

	var mu sync.Mutex
	completed := 0

	for i, chunk := range chunks {
		i, chunk := i, chunk // Important: capture loop variables

		g.Go(func() error {
			vector, err := c.GetEmbedding(gctx, chunk.Text)
			if err != nil {
				return fmt.Errorf("failed at chunk %d: %w", i, err)
			}

			// Store result
			embeddings[i] = Embedding{
				Chunk:     chunk,
				Vector:    vector,
				ModelName: c.ModelName,
			}

			// Safe logging
			mu.Lock()
			completed++
			slog.Info("creating embedding", "current", completed, "total", len(chunks), "file", chunk.FileName)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all to finish
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return embeddings, nil
}