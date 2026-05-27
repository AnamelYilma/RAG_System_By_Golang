package rag

import (
	"context"
	"fmt"
	"strings"

	"MyRagByCivic/embedding"
	"MyRagByCivic/llm"
	"MyRagByCivic/vectorstore"
)

type RAGSystem struct {
	Embedder    *embedding.Client
	VectorStore vectorstore.Store
	LLM         *llm.Client
}

func NewRAGSystem(ctx context.Context, embedModel, llmModel string) (*RAGSystem, error) {
	store, err := vectorstore.NewStore(ctx)
	if err != nil {
		return nil, err
	}

	return &RAGSystem{
		Embedder:    embedding.NewClient(embedModel),
		VectorStore: store,
		LLM:         llm.NewClient(llmModel),
	}, nil
}

func (rag *RAGSystem) Close() error {
	if rag == nil || rag.VectorStore == nil {
		return nil
	}

	return rag.VectorStore.Close()
}

func (rag *RAGSystem) IndexDocuments(ctx context.Context, chunks []embedding.Chunk) error {
	fmt.Printf("   Indexing %d chunks...\n", len(chunks))

	embeddings, err := rag.Embedder.GetEmbeddingsForChunks(ctx, chunks)
	if err != nil {
		return err
	}

	if err := rag.VectorStore.Add(ctx, embeddings); err != nil {
		return err
	}

	fmt.Printf("   Indexed %d chunks successfully\n", len(embeddings))
	return nil
}

func (rag *RAGSystem) Ask(ctx context.Context, question string) (string, error) {
	questionVec, err := rag.Embedder.GetEmbedding(ctx, question)
	if err != nil {
		return "", fmt.Errorf("failed to embed question: %w", err)
	}

	relevantChunks, err := rag.VectorStore.Search(ctx, rag.Embedder.ModelName, questionVec, 3)
	if err != nil {
		return "", fmt.Errorf("failed to search vector store: %w", err)
	}
	if len(relevantChunks) == 0 {
		return "I couldn't find any relevant information in the documents.", nil
	}

	var contextBuilder strings.Builder
	contextBuilder.WriteString("Here are the relevant excerpts from your documents:\n\n")

	sourcesList := make([]string, 0, len(relevantChunks))
	for i, chunk := range relevantChunks {
		fileName := chunk.Chunk.FileName
		sourcesList = append(sourcesList, fileName)

		relevance := int(chunk.Score * 100)
		contextBuilder.WriteString(fmt.Sprintf(
			"SOURCE %d: [%s] (Relevance: %d%%)\n%s\n\n",
			i+1,
			fileName,
			relevance,
			chunk.Chunk.Text,
		))
	}

	uniqueSources := getUniqueSources(sourcesList)
	exampleSource := "your documents"
	if len(uniqueSources) > 0 {
		exampleSource = uniqueSources[0]
	}

	prompt := fmt.Sprintf(`You are a helpful assistant answering questions based ONLY on the provided context.

CONTEXT:
%s

QUESTION: %s

IMPORTANT INSTRUCTIONS:
1. Answer based ONLY on the context above
2. ALWAYS mention which source file the information comes from
3. If information comes from multiple sources, list all of them
4. Be specific - say "According to [filename]..."
5. If answer isn't in context, say "I don't have enough information in your documents"

EXAMPLE ANSWER FORMAT:
"According to %s, [your answer based on the text]"

ANSWER:`, contextBuilder.String(), question, exampleSource)

	messages := []llm.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant that ALWAYS cites which document file the information comes from. Never make up sources.",
		},
		{Role: "user", Content: prompt},
	}

	answer, err := rag.LLM.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	sourceFooter := buildSourceFooter(relevantChunks)
	return fmt.Sprintf("%s\n\n%s", answer, sourceFooter), nil
}

func getUniqueSources(sources []string) []string {
	seen := make(map[string]bool, len(sources))
	unique := make([]string, 0, len(sources))
	for _, source := range sources {
		if seen[source] {
			continue
		}

		seen[source] = true
		unique = append(unique, source)
	}

	return unique
}

func buildSourceFooter(chunks []vectorstore.SearchResult) string {
	var footer strings.Builder
	footer.WriteString("Sources:\n")

	seen := make(map[string]bool, len(chunks))
	for _, chunk := range chunks {
		fileName := chunk.Chunk.FileName
		if !seen[fileName] {
			seen[fileName] = true
			relevance := int(chunk.Score * 100)
			footer.WriteString(fmt.Sprintf("- %s (Relevance: %d%%)\n", fileName, relevance))
		}

		preview := chunk.Chunk.Text
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		footer.WriteString(fmt.Sprintf("  \"%s\"\n", preview))
	}

	return footer.String()
}
