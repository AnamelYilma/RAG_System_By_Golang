package vectorstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"MyRagByCivic/embedding"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	pgvectorpgx "github.com/pgvector/pgvector-go/pgx"
)

const defaultVectorDimension = 768

// =============================================
// FILE PURPOSE
// This file implements permanent vector storage using PostgreSQL + pgvector.
// It saves embeddings to database so they survive program restarts.
// =============================================

// PostgresStore holds database connection and prepared SQL
type PostgresStore struct {
	pool         *pgxpool.Pool   // Connection pool to PostgreSQL
	tableName    string          // Table name (rag_chunks by default)
	vectorDim    int             // Dimension of vectors (e.g. 768)
	insertSQL    string          // Prepared INSERT statement
	deleteSQL    string          // Prepared DELETE statement
	searchSQL    string          // Prepared SEARCH statement
	createSQL    string          // Not used directly (see ensureTable)
	tableCreated bool            // Track if table was created
}

// =============================================
// CONSTRUCTOR
// =============================================

// NewPostgresStore creates and initializes PostgreSQL storage
// What it does: Connects to DB, prepares SQL, validates setup
func NewPostgresStore(ctx context.Context, cfg Config) (*PostgresStore, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, errors.New("postgres vector backend selected but DATABASE_URL or POSTGRES_DSN is empty")
	}

	// Sanitize table name for safety
	tableName, err := sanitizeIdentifier(cfg.TableName)
	if err != nil {
		return nil, err
	}

	// Parse connection string
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres connection config: %w", err)
	}

	// Connection pool settings
	poolConfig.MaxConns = cfg.MaxOpenConns
	if cfg.MaxIdleConns > cfg.MaxOpenConns {
		cfg.MaxIdleConns = cfg.MaxOpenConns
	}
	poolConfig.MinConns = cfg.MaxIdleConns

	// Register pgvector types when connecting
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgvectorpgx.RegisterTypes(ctx, conn)
	}

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect postgres vector store: %w", err)
	}

	// Set vector dimension
	vectorDim := cfg.VectorDimension
	if vectorDim < 1 {
		vectorDim = defaultVectorDimension
	}

	// Prepare SQL statements (for performance)
	store := &PostgresStore{
		pool:      pool,
		tableName: tableName,
		vectorDim: vectorDim,
		insertSQL: fmt.Sprintf(
			"INSERT INTO %s (file_name, chunk_text, start_word, end_word, model_name, embedding) VALUES ($1, $2, $3, $4, $5, $6)",
			tableName,
		),
		deleteSQL: fmt.Sprintf(
			"DELETE FROM %s WHERE file_name = $1 AND model_name = $2",
			tableName,
		),
		searchSQL: fmt.Sprintf(
			"SELECT file_name, chunk_text, start_word, end_word, 1 - (embedding <=> $1) AS score FROM %s WHERE model_name = $2 ORDER BY embedding <=> $1 LIMIT $3",
			tableName,
		),
	}

	// Validate connection and extension
	if err := store.validateSetup(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return store, nil
}

// =============================================
// TABLE MANAGEMENT
// =============================================

// ensureTable creates the table if it doesn't exist
// What it does: Creates table with correct vector dimension
func (ps *PostgresStore) ensureTable(ctx context.Context, dim int) error {
	if ps.tableCreated && dim == ps.vectorDim {
		return nil
	}

	createSQL := fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
			id BIGSERIAL PRIMARY KEY,
			file_name TEXT NOT NULL,
			chunk_text TEXT NOT NULL,
			start_word INTEGER NOT NULL DEFAULT 0,
			end_word INTEGER NOT NULL DEFAULT 0,
			model_name TEXT NOT NULL,
			embedding VECTOR(%d) NOT NULL
		)`,
		ps.tableName, dim,
	)

	if _, err := ps.pool.Exec(ctx, createSQL); err != nil {
		return fmt.Errorf("create vector table %q with dimension %d: %w", ps.tableName, dim, err)
	}

	ps.vectorDim = dim
	ps.tableCreated = true
	return nil
}

// =============================================
// CORE METHODS (Implement Store Interface)
// =============================================

// Add inserts embeddings into database
func (ps *PostgresStore) Add(ctx context.Context, embeddings []embedding.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	dim := len(embeddings[0].Vector)
	if dim == 0 {
		return errors.New("embedding vector is empty")
	}

	// Ensure table exists with correct dimension
	if err := ps.ensureTable(ctx, dim); err != nil {
		return err
	}

	// Use transaction for safety
	tx, err := ps.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin postgres vector transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Delete old entries for same file+model
	replacedSources := make(map[string]struct{}, len(embeddings))
	for _, emb := range embeddings {
		key := sourceKey(emb.Chunk.FileName, emb.ModelName)
		if _, seen := replacedSources[key]; seen {
			continue
		}
		replacedSources[key] = struct{}{}

		if _, err := tx.Exec(ctx, ps.deleteSQL, emb.Chunk.FileName, emb.ModelName); err != nil {
			return fmt.Errorf("delete existing chunks for %s: %w", emb.Chunk.FileName, err)
		}
	}

	// Batch insert new embeddings
	batch := &pgx.Batch{}
	for _, emb := range embeddings {
		batch.Queue(
			ps.insertSQL,
			emb.Chunk.FileName,
			emb.Chunk.Text,
			emb.Chunk.StartWord,
			emb.Chunk.EndWord,
			emb.ModelName,
			pgvector.NewVector(emb.Vector),
		)
	}

	results := tx.SendBatch(ctx, batch)
	for i := 0; i < len(embeddings); i++ {
		if _, err := results.Exec(); err != nil {
			_ = results.Close()
			return fmt.Errorf("insert embedding row %d: %w", i, err)
		}
	}

	if err := results.Close(); err != nil {
		return fmt.Errorf("close postgres batch insert: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit postgres vector transaction: %w", err)
	}

	return nil
}

// Search finds similar vectors using pgvector
func (ps *PostgresStore) Search(ctx context.Context, modelName string, queryVector []float32, topK int) ([]SearchResult, error) {
	topK = normalizeTopK(topK)

	rows, err := ps.pool.Query(ctx, ps.searchSQL, pgvector.NewVector(queryVector), modelName, topK)
	if err != nil {
		return nil, fmt.Errorf("search postgres vectors: %w", err)
	}
	defer rows.Close()

	results := make([]SearchResult, 0, topK)
	position := 0
	for rows.Next() {
		var fileName string
		var chunkText string
		var startWord int
		var endWord int
		var score float64

		if err := rows.Scan(&fileName, &chunkText, &startWord, &endWord, &score); err != nil {
			return nil, fmt.Errorf("scan postgres search row: %w", err)
		}

		results = append(results, SearchResult{
			Chunk: embedding.Chunk{
				Text:      chunkText,
				FileName:  fileName,
				StartWord: startWord,
				EndWord:   endWord,
			},
			Score:    clampScore(float32(score)),
			Position: position,
		})
		position++
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate postgres search rows: %w", err)
	}

	return results, nil
}

// Close closes database connection pool
func (ps *PostgresStore) Close() error {
	ps.pool.Close()
	return nil
}

// validateSetup checks connection and pgvector extension
func (ps *PostgresStore) validateSetup(ctx context.Context) error {
	if err := ps.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres vector store: %w", err)
	}

	if _, err := ps.pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		return fmt.Errorf("postgres is reachable, but the pgvector extension is not available. Install pgvector first: %w", err)
	}

	return nil
}

// clampScore keeps score between 0 and 1
func clampScore(score float32) float32 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}