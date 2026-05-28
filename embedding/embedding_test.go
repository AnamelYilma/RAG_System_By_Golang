package embedding

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
)

// =============================================
// TEST HELPER
// =============================================

// testModelName returns model name from .env or default
// Why: So tests use the same model as your real app
func testModelName() string {
	if modelName := os.Getenv("LM_STUDIO_EMBEDDING_MODEL"); modelName != "" {
		return modelName
	}
	return DefaultModelName
}

// =============================================
// TESTS
// =============================================

// TestNewClient checks client creation
// TestNewClient checks if we can create the embedding client properly
func TestNewClient(t *testing.T) {
	// Goal: Make sure NewClient function works without error
	client := NewClient(testModelName())

	// Check important parts of the client
	if client == nil {
		t.Error("Client should not be nil")
		return
	}

	// Fixed condition: More flexible check
	if client.BaseURL == "" {
		t.Error("BaseURL should not be empty")
	} else if !strings.Contains(client.BaseURL, "1234") && client.BaseURL != DefaultBaseURL {
		t.Errorf("Wrong BaseURL, got: %s", client.BaseURL)
	}

	if client.ModelName == "" {
		t.Error("ModelName should not be empty")
	}

	fmt.Println("✅ TestNewClient passed")
}
// TestGetEmbedding tests single embedding request
// Note: Skips if LM Studio is not running
func TestGetEmbedding(t *testing.T) {
	client := NewClient(testModelName())

	text := "Ethiopia is a beautiful country with rich history and culture."

	vector, err := client.GetEmbedding(context.Background(), text)

	if err != nil {
		t.Skipf("LM Studio not running or model not loaded: %v", err)
		return
	}

	if len(vector) == 0 {
		t.Error("Vector should not be empty")
		return
	}

	fmt.Printf("✅ TestGetEmbedding passed - Vector size: %d numbers\n", len(vector))
}

// TestGetEmbeddingsForChunks tests batch processing
func TestGetEmbeddingsForChunks(t *testing.T) {
	client := NewClient(testModelName())

	sampleChunks := []Chunk{
		{Text: "This is the first chunk about Ethiopian history.", FileName: "unit1.pdf"},
		{Text: "This is the second chunk about Ethiopian culture.", FileName: "unit1.pdf"},
	}

	embeddings, err := client.GetEmbeddingsForChunks(context.Background(), sampleChunks)

	if err != nil {
		t.Skipf("LM Studio not available: %v", err)
		return
	}

	if len(embeddings) != len(sampleChunks) {
		t.Errorf("Expected %d embeddings, got %d", len(sampleChunks), len(embeddings))
		return
	}

	fmt.Printf("✅ TestGetEmbeddingsForChunks passed - Created %d embeddings\n", len(embeddings))
}