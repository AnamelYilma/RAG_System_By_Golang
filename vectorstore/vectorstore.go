package vectorstore

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	"MyRagByCivic/embedding"
)

// =============================================
// FILE PURPOSE
// This file defines the storage interface and configuration logic.
// It acts as a bridge between rag.go and actual storage (memory or postgres).
// =============================================

const (
	BackendMemory   = "memory"   // Fast, temporary storage
	BackendPostgres = "postgres" // Permanent database storage

	DefaultTableName = "rag_chunks"
)

// validIdentifierRE protects against SQL injection in table names
var validIdentifierRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// =============================================
// DATA STRUCTURES
// =============================================

// SearchResult holds one found chunk + similarity score
// Why: Used when answering questions to return relevant text
type SearchResult struct {
	Chunk    embedding.Chunk // The actual text chunk
	Score    float32         // How similar it is (0.0 to 1.0)
	Position int             // Order in results
}

// Store is the main interface
// What it does: Defines what any storage must be able to do
// Why: Allows switching between memory and postgres easily
type Store interface {
	Add(ctx context.Context, embeddings []embedding.Embedding) error
	Search(ctx context.Context, modelName string, queryVector []float32, topK int) ([]SearchResult, error)
	Close() error
}

// Config holds all settings for storage
type Config struct {
	Backend         string
	DatabaseURL     string
	TableName       string
	MaxOpenConns    int32
	MaxIdleConns    int32
	VectorDimension int
}

// =============================================
// CONFIGURATION FUNCTIONS
// =============================================

// LoadConfigFromEnv reads settings from .env file
// What it does: Decides memory or postgres and builds connection string
func LoadConfigFromEnv() Config {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv("RAG_VECTOR_BACKEND")))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("POSTGRES_DSN"))
	}

	// If user gave separate DB_ variables, build full URL
	if databaseURL == "" {
		host := strings.TrimSpace(os.Getenv("DB_HOST"))
		if host != "" {
			user := strings.TrimSpace(os.Getenv("DB_USER"))
			password := strings.TrimSpace(os.Getenv("DB_PASSWORD"))
			dbName := strings.TrimSpace(os.Getenv("DB_NAME"))
			port := strings.TrimSpace(os.Getenv("DB_PORT"))
			if port == "" {
				port = "5432"
			}

			sslmode := strings.TrimSpace(os.Getenv("DB_SSLMODE"))
			if sslmode == "" {
				sslmode = "disable"
			}

			if user == "" {
				user = "postgres"
			}
			if dbName == "" {
				dbName = "postgres"
			}

			databaseURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
				url.QueryEscape(user), url.QueryEscape(password), host, port, dbName, sslmode)
		}
	}

	// Default to memory if nothing specified
	if backend == "" {
		if databaseURL != "" {
			backend = BackendPostgres
		} else {
			backend = BackendMemory
		}
	}

	tableName := strings.TrimSpace(os.Getenv("RAG_VECTOR_TABLE"))
	if tableName == "" {
		tableName = DefaultTableName
	}

	return Config{
		Backend:         backend,
		DatabaseURL:     databaseURL,
		TableName:       tableName,
		MaxOpenConns:    getEnvInt32("RAG_PG_MAX_OPEN_CONNS", 10),
		MaxIdleConns:    getEnvInt32("RAG_PG_MAX_IDLE_CONNS", 5),
		VectorDimension: int(getEnvInt32("RAG_VECTOR_DIMENSION", 0)),
	}
}

// NewStore creates the correct storage type
func NewStore(ctx context.Context) (Store, error) {
	return NewStoreWithConfig(ctx, LoadConfigFromEnv())
}

// NewStoreWithConfig chooses memory or postgres based on config
func NewStoreWithConfig(ctx context.Context, cfg Config) (Store, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Backend)) {
	case "", BackendMemory:
		return NewInMemoryStore(), nil
	case BackendPostgres:
		return NewPostgresStore(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported vector backend %q", cfg.Backend)
	}
}

// =============================================
// HELPER FUNCTIONS
// =============================================

// normalizeTopK ensures we always return at least 1 result
func normalizeTopK(topK int) int {
	if topK < 1 {
		return 1
	}
	return topK
}

// sanitizeIdentifier protects table name for SQL safety
func sanitizeIdentifier(name string) (string, error) {
	if !validIdentifierRE.MatchString(name) {
		return "", fmt.Errorf("invalid SQL identifier %q", name)
	}
	return name, nil
}

// getEnvInt32 reads integer from environment with fallback
func getEnvInt32(name string, fallback int32) int32 {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return fallback
	}

	return int32(value)
}

// sourceKey creates unique key for each file + model combination
func sourceKey(fileName, modelName string) string {
	return fileName + "\x00" + modelName // \x00 is null character separator
}