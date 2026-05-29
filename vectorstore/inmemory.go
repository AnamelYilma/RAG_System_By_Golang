package vectorstore

import (
	"context"
	"math"
	"sort"

	"MyRagByCivic/embedding"
)

// =============================================
// FILE PURPOSE
// This file implements the in-memory (RAM) storage backend.
// Data exists only while the program is running.
// =============================================

// InMemoryStore holds all embeddings in memory
// Why struct with slice: Simple and fast for small to medium data
type InMemoryStore struct {
	embeddings []embedding.Embedding // All stored embeddings
}

// =============================================
// CONSTRUCTOR
// =============================================

// NewInMemoryStore creates a new empty memory store
// What it does: Initializes empty slice
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		embeddings: []embedding.Embedding{}, // Start with empty list
	}
}

// =============================================
// MAIN METHODS (Implement Store Interface)
// =============================================

// Add saves new embeddings and replaces old ones for same file+model
// What it does: Replace old data for same document to avoid duplicates
func (vs *InMemoryStore) Add(_ context.Context, embeddings []embedding.Embedding) error {
	if len(embeddings) == 0 {
		return nil // Nothing to do
	}

	// Step 1: Find which sources (file+model) we need to replace
	replaceSources := make(map[string]struct{}, len(embeddings))
	for _, emb := range embeddings {
		replaceSources[sourceKey(emb.Chunk.FileName, emb.ModelName)] = struct{}{}
	}

	// Step 2: Remove old data for those sources
	filtered := make([]embedding.Embedding, 0, len(vs.embeddings))
	for _, existing := range vs.embeddings {
		key := sourceKey(existing.Chunk.FileName, existing.ModelName)
		if _, shouldReplace := replaceSources[key]; shouldReplace {
			continue // Skip old version
		}
		filtered = append(filtered, existing)
	}

	// Step 3: Add new embeddings
	vs.embeddings = append(filtered, embeddings...)
	return nil
}

// Search finds most similar chunks to the question vector
// What it does: Calculate similarity → sort → return top results
func (vs *InMemoryStore) Search(_ context.Context, modelName string, queryVector []float32, topK int) ([]SearchResult, error) {
	if len(vs.embeddings) == 0 {
		return []SearchResult{}, nil
	}

	topK = normalizeTopK(topK)

	results := make([]SearchResult, 0, len(vs.embeddings))

	// Calculate similarity score for each embedding
	for i, emb := range vs.embeddings {
		// Filter by model name if specified
		if modelName != "" && emb.ModelName != modelName {
			continue
		}

		results = append(results, SearchResult{
			Chunk:    emb.Chunk,
			Score:    cosineSimilarity(queryVector, emb.Vector),
			Position: i,
		})
	}

	// Sort by score (highest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Return only topK results
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// Close does nothing for memory store
// Why: No resources to clean up
func (vs *InMemoryStore) Close() error {
	return nil
}

// =============================================
// HELPER FUNCTION
// =============================================

// cosineSimilarity calculates how similar two vectors are
// What it does: Returns score between -1 and 1 (higher = more similar)
// Why: This is the core of vector search (semantic similarity)
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float64
	var normA float64
	var normB float64

	// Calculate dot product and magnitudes
	for i := 0; i < len(a); i++ {
		ai := float64(a[i])
		bi := float64(b[i])
		dotProduct += ai * bi
		normA += ai * ai
		normB += bi * bi
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	score := dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))

	// Clamp between -1 and 1
	if score < -1 {
		return -1
	}
	if score > 1 {
		return 1
	}

	return float32(score)
}