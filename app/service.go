package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"MyRagByCivic/chunker"
	"MyRagByCivic/config"
	"MyRagByCivic/rag"

	documentloaders "github.com/tmc/langchaingo/documentloaders"
)

// IndexReport describes the result of a PDF indexing run.
type IndexReport struct {
	PDFFiles        int       `json:"pdf_files"`
	IndexedFiles    int       `json:"indexed_files"`
	ChunksIndexed   int       `json:"chunks_indexed"`
	FailedFiles     int       `json:"failed_files"`
	LastIndexedAt   time.Time `json:"last_indexed_at"`
	Ready           bool      `json:"ready"`
	SourceDirectory string    `json:"source_directory"`
}

// Service owns the RAG system and the PDF indexing workflow.
type Service struct {
	cfg    config.Config
	rag    *rag.RAGSystem
	mu     sync.RWMutex
	ready  bool
	report IndexReport
}

// NewService creates the shared backend service used by HTTP handlers.
func NewService(ctx context.Context, cfg config.Config) (*Service, error) {
	ragSystem, err := rag.NewRAGSystem(ctx, "", "")
	if err != nil {
		return nil, err
	}

	return &Service{
		cfg: cfg,
		rag: ragSystem,
		report: IndexReport{
			SourceDirectory: cfg.PDFDir,
		},
	}, nil
}

// Close releases any resources held by the RAG system.
func (s *Service) Close() error {
	if s == nil || s.rag == nil {
		return nil
	}
	return s.rag.Close()
}

// Status returns the latest indexing snapshot.
func (s *Service) Status() IndexReport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.report
}

// IsReady reports whether the app has at least one indexed chunk.
func (s *Service) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready
}

// Ask sends a question to the RAG system.
func (s *Service) Ask(ctx context.Context, question string) (string, error) {
	if !s.IsReady() {
		return "", errors.New("no indexed documents are available yet")
	}
	return s.rag.Ask(ctx, question)
}

// IndexDocumentsFromPDF scans the PDF folder and indexes every PDF file.
func (s *Service) IndexDocumentsFromPDF(ctx context.Context) (IndexReport, error) {
	report := IndexReport{
		SourceDirectory: s.cfg.PDFDir,
	}

	files, err := os.ReadDir(s.cfg.PDFDir)
	if err != nil {
		return report, fmt.Errorf("read PDF directory %q: %w", s.cfg.PDFDir, err)
	}

	var totalChunks int
	var indexedFiles int
	var failedFiles int
	for _, file := range files {
		if filepath.Ext(strings.ToLower(file.Name())) != ".pdf" {
			continue
		}

		report.PDFFiles++
		filePath := filepath.Join(s.cfg.PDFDir, file.Name())
		chunks, err := s.indexOnePDF(ctx, filePath, file.Name())
		if err != nil {
			failedFiles++
			continue
		}

		if len(chunks) == 0 {
			continue
		}

		if err := s.rag.IndexDocuments(ctx, chunks); err != nil {
			failedFiles++
			continue
		}

		totalChunks += len(chunks)
		indexedFiles++
	}

	report.IndexedFiles = indexedFiles
	report.ChunksIndexed = totalChunks
	report.FailedFiles = failedFiles
	report.LastIndexedAt = time.Now().UTC()
	report.Ready = totalChunks > 0

	s.mu.Lock()
	s.ready = report.Ready
	s.report = report
	s.mu.Unlock()

	if totalChunks == 0 {
		return report, errors.New("no PDF chunks were indexed")
	}

	return report, nil
}

func (s *Service) indexOnePDF(ctx context.Context, filePath, displayName string) ([]chunker.Chunk, error) {
	// This block reads a PDF file and turns page text into chunks.
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open PDF %q: %w", displayName, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat PDF %q: %w", displayName, err)
	}

	loader := documentloaders.NewPDF(file, stat.Size())
	// loader.Load is a method from the langchaingo library, not a Go builtin.
	docs, err := loader.Load(ctx)
	if err != nil {
		return nil, fmt.Errorf("load PDF %q: %w", displayName, err)
	}

	var fullText strings.Builder
	for _, doc := range docs {
		fullText.WriteString(doc.PageContent)
		fullText.WriteString("\n")
	}

	// Our own function: SliceText cleans and splits the document into chunks.
	return chunker.SliceText(fullText.String(), s.cfg.ChunkSize, s.cfg.Overlap, displayName), nil
}
