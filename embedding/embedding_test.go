package embedding

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// =============================================
// embedding_test.go
// Goal of this file: Test our embedding code to make sure it works correctly
// This helps us catch problems early and is good for portfolio
// =============================================

func testModelName() string {
	if modelName := os.Getenv("LM_STUDIO_EMBEDDING_MODEL"); modelName != "" {
		return modelName
	}

	return DefaultModelName
}

// TestNewClient checks if we can create the embedding client properly
func TestNewClient(t *testing.T) {
	// Goal: Make sure NewClient function works without error
	client := NewClient(testModelName())

	// Check important parts of the client
	if client == nil {
		t.Error("Client should not be nil")
		return
	}
	if client.BaseURL != DefaultBaseURL {
		t.Errorf("Wrong BaseURL, got: %s", client.BaseURL)
	}
	if client.ModelName != testModelName() {
		t.Errorf("Wrong ModelName, got: %s", client.ModelName)
	}

	fmt.Println("✅ TestNewClient passed")
}

// TestGetEmbedding tests getting one vector from LM Studio
// Note: This test needs LM Studio running with your model loaded
func TestGetEmbedding(t *testing.T) {
	// Goal: Test converting one text into numbers (vector)
	client := NewClient(testModelName())

	text := "Ethiopia is a beautiful country with rich history and culture."

	vector, err := client.GetEmbedding(context.Background(), text)

	if err != nil {
		t.Skipf("LM Studio embedding request failed for model %q at %s: %v", client.ModelName, client.BaseURL, err)
		return
	}

	if len(vector) == 0 {
		t.Error("Vector should not be empty")
		return
	}

	fmt.Printf("✅ TestGetEmbedding passed - Vector size: %d numbers\n", len(vector))
}

// TestGetEmbeddingsForChunks tests processing multiple chunks at once
// Note: This test needs LM Studio running
func TestGetEmbeddingsForChunks(t *testing.T) {
	// Goal: Test creating embeddings for many chunks together
	client := NewClient(testModelName())

	// Create sample chunks for testing (no need for real PDF)
	sampleChunks := []Chunk{
		{
			Text:     "This is the first chunk about Ethiopian history.",
			FileName: "unit1.pdf",
		},
		{
			Text:     "This is the second chunk about Ethiopian culture.",
			FileName: "unit1.pdf",
		},
	}

	embeddings, err := client.GetEmbeddingsForChunks(context.Background(), sampleChunks)

	if err != nil {
		t.Skipf("LM Studio embedding request failed for model %q at %s: %v", client.ModelName, client.BaseURL, err)
		return
	}

	if len(embeddings) != len(sampleChunks) {
		t.Errorf("Expected %d embeddings, got %d", len(sampleChunks), len(embeddings))
		return
	}

	fmt.Printf("✅ TestGetEmbeddingsForChunks passed - Created %d embeddings\n", len(embeddings))
}
