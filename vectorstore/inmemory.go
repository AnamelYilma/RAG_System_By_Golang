package vectorstore

import (
	"context"
	"math"
	"sort"

	"MyRagByCivic/embedding"
)

type InMemoryStore struct {
	embeddings []embedding.Embedding
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		embeddings: []embedding.Embedding{},
	}
}

func (vs *InMemoryStore) Add(_ context.Context, embeddings []embedding.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	replaceSources := make(map[string]struct{}, len(embeddings))
	for _, emb := range embeddings {
		replaceSources[sourceKey(emb.Chunk.FileName, emb.ModelName)] = struct{}{}
	}

	filtered := make([]embedding.Embedding, 0, len(vs.embeddings))
	for _, existing := range vs.embeddings {
		if _, shouldReplace := replaceSources[sourceKey(existing.Chunk.FileName, existing.ModelName)]; shouldReplace {
			continue
		}

		filtered = append(filtered, existing)
	}

	vs.embeddings = append(filtered, embeddings...)
	return nil
}

func (vs *InMemoryStore) Search(_ context.Context, modelName string, queryVector []float32, topK int) ([]SearchResult, error) {
	if len(vs.embeddings) == 0 {
		return []SearchResult{}, nil
	}

	topK = normalizeTopK(topK)

	results := make([]SearchResult, 0, len(vs.embeddings))
	for i, emb := range vs.embeddings {
		if modelName != "" && emb.ModelName != modelName {
			continue
		}

		results = append(results, SearchResult{
			Chunk:    emb.Chunk,
			Score:    cosineSimilarity(queryVector, emb.Vector),
			Position: i,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

func (vs *InMemoryStore) Close() error {
	return nil
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float32
	var normA float32
	var normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	score := dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
	if score < -1 {
		return -1
	}
	if score > 1 {
		return 1
	}

	return score
}
