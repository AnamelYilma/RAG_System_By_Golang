package vectorstore

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"MyRagByCivic/embedding"
)

const (
	BackendMemory   = "memory"
	BackendPostgres = "postgres"

	DefaultTableName = "rag_chunks"
)

var validIdentifierRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

type SearchResult struct {
	Chunk    embedding.Chunk
	Score    float32
	Position int
}

type Store interface {
	Add(ctx context.Context, embeddings []embedding.Embedding) error
	Search(ctx context.Context, modelName string, queryVector []float32, topK int) ([]SearchResult, error)
	Close() error
}

type Config struct {
	Backend      string
	DatabaseURL  string
	TableName    string
	MaxOpenConns int32
	MaxIdleConns int32
}

func LoadConfigFromEnv() Config {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv("RAG_VECTOR_BACKEND")))
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		databaseURL = strings.TrimSpace(os.Getenv("POSTGRES_DSN"))
	}
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
		Backend:      backend,
		DatabaseURL:  databaseURL,
		TableName:    tableName,
		MaxOpenConns: getEnvInt32("RAG_PG_MAX_OPEN_CONNS", 10),
		MaxIdleConns: getEnvInt32("RAG_PG_MAX_IDLE_CONNS", 5),
	}
}

func NewStore(ctx context.Context) (Store, error) {
	return NewStoreWithConfig(ctx, LoadConfigFromEnv())
}

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

func normalizeTopK(topK int) int {
	if topK < 1 {
		return 1
	}

	return topK
}

func sanitizeIdentifier(name string) (string, error) {
	if !validIdentifierRE.MatchString(name) {
		return "", fmt.Errorf("invalid SQL identifier %q", name)
	}

	return name, nil
}

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

func sourceKey(fileName, modelName string) string {
	return fileName + "\x00" + modelName
}
