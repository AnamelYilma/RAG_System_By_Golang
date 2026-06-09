package main

import (
	"context"
	"fmt"
	"log"

	"MyRagByCivic/app"
	"MyRagByCivic/config"
	"MyRagByCivic/handler"
	"MyRagByCivic/router"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load local environment variables first, then fall back to the shell.
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()
	ctx := context.Background()

	// This service holds the RAG system and the PDF indexing workflow.
	service, err := app.NewService(ctx, cfg)
	if err != nil {
		log.Fatalf("create app service: %v", err)
	}
	defer service.Close()

	// Index PDFs once at startup so the frontend has data to query.
	report, err := service.IndexDocumentsFromPDF(ctx)
	if err != nil {
		log.Printf("startup indexing finished with warning: %v", err)
	}
	
	log.Printf("indexed %d file(s) and %d chunk(s)", report.IndexedFiles, report.ChunksIndexed)

	engine := gin.Default()
	router.Register(engine, handler.New(service))

	log.Printf("server listening on %s", cfg.Addr())
	if err := engine.Run(cfg.Addr()); err != nil {
		log.Fatalf("run http server: %v", err)
	}
}
