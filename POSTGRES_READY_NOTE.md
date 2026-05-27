# PostgreSQL Vector Mode

This project now supports two vector backends:

- `memory`
- `postgres`

If `RAG_VECTOR_BACKEND` is not set:

- it uses `postgres` when `DATABASE_URL` or `POSTGRES_DSN` is set
- otherwise it falls back to `memory`

## Environment variables

- `RAG_VECTOR_BACKEND`
- `DATABASE_URL`
- `POSTGRES_DSN`
- `RAG_VECTOR_TABLE`
- `RAG_PG_MAX_OPEN_CONNS`
- `RAG_PG_MAX_IDLE_CONNS`
- `LM_STUDIO_BASE_URL`
- `LM_STUDIO_EMBEDDING_MODEL`
- `LM_STUDIO_CHAT_MODEL`

## PostgreSQL requirements

Before running in PostgreSQL mode, create a database in pgAdmin and prepare it with:

- the `vector` extension enabled
- a table for chunk embeddings

The Go code expects the table name from `RAG_VECTOR_TABLE`, or `rag_chunks` by default.

The expected columns are:

- `file_name`
- `chunk_text`
- `start_word`
- `end_word`
- `model_name`
- `embedding`

## What the app does in PostgreSQL mode

- connects through `DATABASE_URL` or `POSTGRES_DSN`
- checks that the `vector` extension exists
- checks that the embeddings table exists
- deletes older rows for the same source file and model before re-inserting fresh chunks
- searches with pgvector cosine distance

## Important note

This project code is now ready to talk to PostgreSQL, but it still needs:

- a real PostgreSQL database created by you
- the pgvector extension enabled in that database
- the embeddings table created in that database
- LM Studio running for embedding generation and answer generation
