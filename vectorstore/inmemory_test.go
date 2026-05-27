package vectorstore

import (
	"context"
	"testing"

	"MyRagByCivic/embedding"
)

func TestInMemoryStoreAddReplacesExistingDocument(t *testing.T) {
	t.Parallel()

	store := NewInMemoryStore()
	ctx := context.Background()

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

	if err := store.Add(ctx, initial); err != nil {
		t.Fatalf("add initial embeddings: %v", err)
	}
	if err := store.Add(ctx, replacement); err != nil {
		t.Fatalf("add replacement embeddings: %v", err)
	}

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

	if err := store.Add(ctx, embeddings); err != nil {
		t.Fatalf("add embeddings: %v", err)
	}

	results, err := store.Search(ctx, "model-a", []float32{1, 0}, 10)
	if err != nil {
		t.Fatalf("search embeddings: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(results))
	}
	if results[0].Chunk.FileName != "unit1.pdf" {
		t.Fatalf("expected model-a result from unit1.pdf, got %q", results[0].Chunk.FileName)
	}
}
