package vectorstore

import (
	"context"
	"testing"

	"MyRagByCivic/embedding"
)

// =============================================
// FILE PURPOSE
// This file tests the InMemoryStore to ensure it behaves correctly.
// Tests help catch bugs early and serve as documentation of expected behavior.
// =============================================

// =============================================
// TEST 1: Replacement Logic
// =============================================

// TestInMemoryStoreAddReplacesExistingDocument checks that old chunks
// for the same file+model are replaced when new chunks arrive.
// Why this test is important: Prevents duplicate/old data accumulation.
func TestInMemoryStoreAddReplacesExistingDocument(t *testing.T) {
	t.Parallel() // Run this test in parallel with others for faster execution

	store := NewInMemoryStore()
	ctx := context.Background()

	// First add old data
	initial := []embedding.Embedding{
		{
			Chunk: embedding.Chunk{
				FileName:  "unit1.pdf",
				Text:      "old text",
				StartWord: 0,
				EndWord:   5,
			},
			Vector:    []float32{1, 0},
			ModelName: "model-a",
		},
	}

	// Then add replacement data for same file+model
	replacement := []embedding.Embedding{
		{
			Chunk: embedding.Chunk{
				FileName:  "unit1.pdf",
				Text:      "new text",
				StartWord: 0,
				EndWord:   4,
			},
			Vector:    []float32{0, 1},
			ModelName: "model-a",
		},
	}

	// Execute the operations
	if err := store.Add(ctx, initial); err != nil {
		t.Fatalf("add initial embeddings: %v", err)
	}
	if err := store.Add(ctx, replacement); err != nil {
		t.Fatalf("add replacement embeddings: %v", err)
	}

	// Verify that old text was replaced
	results, err := store.Search(ctx, "model-a", []float32{0, 1}, 5)
	if err != nil {
		t.Fatalf("search embeddings: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(results))
	}
	if results[0].Chunk.Text != "new text" {
		t.Fatalf("expected replacement chunk text, got %q", results[0].Chunk.Text)
	}
}

// =============================================
// TEST 2: Model Filtering
// =============================================

// TestInMemoryStoreSearchFiltersByModel checks that search respects model name
// Why: Different embedding models should not mix in search results
func TestInMemoryStoreSearchFiltersByModel(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	ctx := context.Background()

	embeddings := []embedding.Embedding{
		{
			Chunk:     embedding.Chunk{FileName: "unit1.pdf", Text: "match"},
			Vector:    []float32{1, 0},
			ModelName: "model-a",
		},
		{
			Chunk:     embedding.Chunk{FileName: "unit2.pdf", Text: "other model"},
			Vector:    []float32{1, 0},
			ModelName: "model-b",
		},
	}

	// Add mixed data
	if err := store.Add(ctx, embeddings); err != nil {
		t.Fatalf("add embeddings: %v", err)
	}

	// Search only for model-a
	results, err := store.Search(ctx, "model-a", []float32{1, 0}, 10)
	if err != nil {
		t.Fatalf("search embeddings: %v", err)
	}

	// Should return only model-a result
	if len(results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(results))
	}
	if results[0].Chunk.FileName != "unit1.pdf" {
		t.Fatalf("expected model-a result from unit1.pdf, got %q", results[0].Chunk.FileName)
	}
}