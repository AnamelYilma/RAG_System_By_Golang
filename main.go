package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"MyRagByCivic/chunker"
	"MyRagByCivic/rag"

	"github.com/joho/godotenv"
	documentloaders "github.com/tmc/langchaingo/documentloaders"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	inputFolder := "./PDF"
	chunkSize := 450
	overlap := 90

	fmt.Println("Initializing RAG system...")
	ragSystem, err := rag.NewRAGSystem(ctx, "", "")
	if err != nil {
		fmt.Println("Error creating RAG system:", err)
		return
	}
	defer ragSystem.Close()

	fmt.Println("Loading and indexing PDFs...")
	files, err := os.ReadDir(inputFolder)
	if err != nil {
		fmt.Println("Error reading PDF folder:", err)
		return
	}

	totalChunks := 0
	indexingFailures := 0
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".pdf" {
			continue
		}

		filePath := filepath.Join(inputFolder, file.Name())
		fmt.Printf("\nProcessing: %s\n", file.Name())

		f, err := os.Open(filePath)
		if err != nil {
			fmt.Println("   Cannot open file:", err)
			continue
		}

		fileInfo, err := f.Stat()
		if err != nil {
			fmt.Println("   Cannot read file info:", err)
			f.Close()
			continue
		}

		loader := documentloaders.NewPDF(f, fileInfo.Size())
		docs, err := loader.Load(ctx)
		f.Close()
		if err != nil {
			fmt.Println("   PDF load error:", err)
			continue
		}

		var fullText strings.Builder
		for _, doc := range docs {
			fullText.WriteString(doc.PageContent)
			fullText.WriteString("\n")
		}

		chunks := chunker.SliceText(fullText.String(), chunkSize, overlap, file.Name())
		fmt.Printf("   Created %d chunks\n", len(chunks))

		if err := ragSystem.IndexDocuments(ctx, chunks); err != nil {
			fmt.Printf("   Indexing error: %v\n", err)
			indexingFailures++
			continue
		}

		totalChunks += len(chunks)
	}

	fmt.Printf("\nIndexing complete. %d total chunks indexed.\n", totalChunks)
	if indexingFailures > 0 {
		fmt.Printf("Indexing had %d file-level failure(s).\n", indexingFailures)
	}
	if totalChunks == 0 {
		fmt.Println("No chunks were indexed. Start LM Studio and load your embedding model before asking questions.")
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("RAG system ready. Ask questions about your documents.")
	fmt.Println("Type 'exit' to quit.")
	fmt.Println(strings.Repeat("=", 50))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		question := strings.TrimSpace(scanner.Text())
		if question == "exit" {
			break
		}
		if question == "" {
			continue
		}

		fmt.Print("\nAssistant: ")
		answer, err := ragSystem.Ask(ctx, question)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}

		fmt.Println(answer)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("\nInput error:", err)
	}

	fmt.Println("\nGoodbye.")
}
