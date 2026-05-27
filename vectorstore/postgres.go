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

type PostgresStore struct {
	pool      *pgxpool.Pool
	tableName string
	insertSQL string
	deleteSQL string
	searchSQL string
}

func NewPostgresStore(ctx context.Context, cfg Config) (*PostgresStore, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, errors.New("postgres vector backend selected but DATABASE_URL or POSTGRES_DSN is empty")
	}

	tableName, err := sanitizeIdentifier(cfg.TableName)
	if err != nil {
		return nil, err
	}

	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres connection config: %w", err)
	}

	poolConfig.MaxConns = cfg.MaxOpenConns
	if cfg.MaxIdleConns > cfg.MaxOpenConns {
		cfg.MaxIdleConns = cfg.MaxOpenConns
	}
	poolConfig.MinConns = cfg.MaxIdleConns
	poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgvectorpgx.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("connect postgres vector store: %w", err)
	}

	store := &PostgresStore{
		pool:      pool,
		tableName: tableName,
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

	if err := store.validateSetup(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return store, nil
}

func (ps *PostgresStore) Add(ctx context.Context, embeddings []embedding.Embedding) error {
	if len(embeddings) == 0 {
		return nil
	}

	tx, err := ps.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin postgres vector transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

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

func (ps *PostgresStore) Close() error {
	ps.pool.Close()
	return nil
}

func (ps *PostgresStore) validateSetup(ctx context.Context) error {
	if err := ps.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres vector store: %w", err)
	}

	var hasVectorExtension bool
	if err := ps.pool.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&hasVectorExtension); err != nil {
		return fmt.Errorf("check pgvector extension: %w", err)
	}
	if !hasVectorExtension {
		return errors.New("postgres is reachable, but the pgvector extension is not enabled. Enable the 'vector' extension before running the app")
	}

	var hasTable bool
	if err := ps.pool.QueryRow(
		ctx,
		"SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1)",
		ps.tableName,
	).Scan(&hasTable); err != nil {
		return fmt.Errorf("check vector table %q: %w", ps.tableName, err)
	}
	if !hasTable {
		return fmt.Errorf(
			"postgres is reachable, but table %q does not exist. Create a table with columns file_name, chunk_text, start_word, end_word, model_name, and embedding",
			ps.tableName,
		)
	}

	return nil
}

func clampScore(score float32) float32 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}

	return score
}
