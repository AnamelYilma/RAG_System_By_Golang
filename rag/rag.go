package rag

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"MyRagByCivic/embedding"
	"MyRagByCivic/llm"
	"MyRagByCivic/vectorstore"
)

// =============================================
// FILE PURPOSE
// This is the central manager of the RAG system.
// It connects all parts and provides two main functions: IndexDocuments and Ask.
// =============================================

// =============================================
// MAIN STRUCTURE
// =============================================

// RAGSystem holds the three main tools needed for RAG
// Why we use struct: To keep all tools in one place and pass them together
type RAGSystem struct {
	Embedder    *embedding.Client     // Used to create vectors
	VectorStore vectorstore.Store     // Used to save and search vectors
	LLM         *llm.Client           // Used to generate final answer
}

// =============================================
// CONSTRUCTOR
// =============================================

// NewRAGSystem creates the full RAG system
// What it does: Creates store, embedding client, and LLM client
// Why: Called once at startup in main.go
func NewRAGSystem(ctx context.Context, embedModel, llmModel string) (*RAGSystem, error) {
	// Create vector store (memory or postgres)
	store, err := vectorstore.NewStore(ctx)
	if err != nil {
		return nil, err
	}

	// Return fully connected RAG system
	return &RAGSystem{
		Embedder:    embedding.NewClient(embedModel),
		VectorStore: store,
		LLM:         llm.NewClient(llmModel),
	}, nil
}

// =============================================
// CLEANUP FUNCTION
// =============================================

// Close cleans up resources when program ends
// What it does: Closes database connection if using postgres
// Why: Prevents memory leaks and connection problems
func (rag *RAGSystem) Close() error {
	if rag == nil || rag.VectorStore == nil {
		return nil // Safety check
	}

	return rag.VectorStore.Close()
}

// =============================================
// INDEXING FUNCTION
// =============================================

// IndexDocuments takes chunks and saves them as vectors
// What it does: Embed → Store
// Why: This is how the system "learns" your PDFs
func (rag *RAGSystem) IndexDocuments(ctx context.Context, chunks []embedding.Chunk) error {
	slog.Info("indexing chunks", "count", len(chunks))

	// Step 1: Convert text chunks to vectors
	embeddings, err := rag.Embedder.GetEmbeddingsForChunks(ctx, chunks)
	if err != nil {
		return err
	}

	// Step 2: Save vectors into memory or database
	if err := rag.VectorStore.Add(ctx, embeddings); err != nil {
		return err
	}

	slog.Info("indexed chunks", "count", len(embeddings))
	return nil
}

// =============================================
// ASK FUNCTION (Most Important)
// =============================================

// Ask answers user question using RAG flow
// What it does: Embed question → Search → Build prompt → Generate answer
func (rag *RAGSystem) Ask(ctx context.Context, question string) (string, error) {
	// Step 1: Turn user question into vector
	questionVec, err := rag.Embedder.GetEmbedding(ctx, question)
	if err != nil {
		return "", fmt.Errorf("failed to embed question: %w", err)
	}

	// Step 2: Find most relevant chunks
	relevantChunks, err := rag.VectorStore.Search(ctx, rag.Embedder.ModelName, questionVec, 3)
	if err != nil {
		return "", fmt.Errorf("failed to search vector store: %w", err)
	}
	if len(relevantChunks) == 0 {
		return "I couldn't find any relevant information in the documents.", nil
	}

	// Step 3: Build clean context (no SOURCE 1, no Relevance %)
	var contextBuilder strings.Builder
	for _, chunk := range relevantChunks {
		contextBuilder.WriteString(chunk.Chunk.Text)
		contextBuilder.WriteString("\n\n")
	}

	// === Clean Prompt - This controls how AI answers ===
	prompt := fmt.Sprintf(`You are a helpful and direct teacher for Civic Education.

Answer the question in a natural, simple, and clear way.
Do NOT say "According to", "Source", "The document says", or mention file names in your answer.
Just give the answer directly.

Context from documents:
%s

Question: %s

Answer:`, contextBuilder.String(), question)

	// Prepare messages for LLM
	messages := []llm.Message{
		{Role: "system", Content: "You are a helpful assistant. Answer directly and naturally. Never mention sources in the main answer."},
		{Role: "user", Content: prompt},
	}

	// Step 4: Get answer from LLM
	answer, err := rag.LLM.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	// Step 5: Add clean sources at the bottom
	finalAnswer := strings.TrimSpace(answer)
	sourcesText := buildCleanSourceFooter(relevantChunks)

	return finalAnswer + "\n\n" + sourcesText, nil
}

// =============================================
// HELPER FUNCTIONS
// =============================================

// buildCleanSourceFooter creates simple sources list
// What it does: Shows sources without too much detail
func buildCleanSourceFooter(chunks []vectorstore.SearchResult) string {
	var footer strings.Builder
	footer.WriteString("--- Sources ---\n")

	seen := make(map[string]bool)
	for _, chunk := range chunks {
		fileName := chunk.Chunk.FileName
		if seen[fileName] {
			continue
		}
		seen[fileName] = true

		footer.WriteString(fmt.Sprintf("• %s\n", fileName))
	}

	return footer.String()
}

// (Old functions kept for compatibility - can be removed later)
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