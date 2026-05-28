# Simple Guide for This RAG Project

This guide is written in simple English.

The goal is to help you understand this project step by step, even if you are still a beginner in Go, RAG, or project structure.

You do not need to understand everything at one time.
Read it slowly, section by section.

## 1. What this project does

This project is a small RAG system written in Go.

RAG means:

- first find useful text from your documents
- then give that text to an AI model
- then the AI model answers based on that text

In this project, the documents are PDF files.

So the idea is:

1. put PDF files into the `PDF/` folder
2. run the Go program
3. the program reads the PDFs
4. the program cuts PDF text into small pieces
5. the program turns each piece into numbers called embeddings
6. the program stores those embeddings
7. when you ask a question, the program finds the most related pieces
8. the program sends those pieces to the chat model
9. the chat model gives the final answer

This project does not train a new model.
It uses existing models through LM Studio style HTTP APIs.

## 2. The big idea in one small story

Think of the system like this:

- `PDF/` folder = your study materials
- `chunker` = the cleaner and cutter
- `embedding` = the part that changes text into numbers
- `vectorstore` = the memory box for those numbers
- `rag` = the manager that connects everything
- `llm` = the part that writes the final answer
- `main.go` = the starting point that runs the whole project

If you remember only one thing, remember this:

`PDF -> text -> chunks -> embeddings -> search -> answer`

## 3. Easy words used in this project

Here are the main words you will see:

- `PDF loader`: code that reads PDF files and gets text from them
- `chunk`: one small piece of text taken from a PDF
- `embedding`: a list of numbers that represents the meaning of text
- `vector`: another name for the list of numbers
- `vector store`: a place to save vectors and search similar vectors
- `retrieval`: finding the most useful chunks for a question
- `prompt`: the full message sent to the chat model
- `client`: Go code that sends HTTP requests to another service
- `LM Studio`: the local server that gives embeddings and chat answers

## 4. Main things inside this project

There are 3 groups to know:

### A. Your own Go packages in this project

- `main`
- `config`
- `chunker`
- `embedding`
- `llm`
- `rag`
- `vectorstore`

### B. Outside tools and services

- Go 1.25.1
- PDF files inside `PDF/`
- LM Studio or another OpenAI-compatible local API server
- PostgreSQL with `pgvector` if you want database vector storage
- a `.env` file for settings

### C. Go libraries used by the project

These are the direct libraries written in `go.mod`:

- `github.com/jackc/pgx/v5`
- `github.com/pgvector/pgvector-go`
- `github.com/pgvector/pgvector-go/pgx`
- `github.com/tmc/langchaingo`

These are also used in the code and are important to know:

- `github.com/joho/godotenv`
- `gorm.io/gorm`
- `gorm.io/driver/postgres`

Important real note:

- `gorm.io/gorm` and `gorm.io/driver/postgres` are imported in the code right now, but they are not listed in `go.mod` yet.
- Because of that, `go test ./...` currently fails until those packages are added.

These are the indirect libraries already listed in `go.mod`:

- `github.com/AssemblyAI/assemblyai-go-sdk`
- `github.com/PuerkitoBio/goquery`
- `github.com/andybalholm/cascadia`
- `github.com/aymerick/douceur`
- `github.com/cenkalti/backoff`
- `github.com/dlclark/regexp2`
- `github.com/google/go-querystring`
- `github.com/google/uuid`
- `github.com/gorilla/css`
- `github.com/jackc/pgpassfile`
- `github.com/jackc/pgservicefile`
- `github.com/jackc/puddle/v2`
- `github.com/joho/godotenv`
- `github.com/klauspost/compress`
- `github.com/ledongthuc/pdf`
- `github.com/microcosm-cc/bluemonday`
- `github.com/pkoukk/tiktoken-go`
- `github.com/x448/float16`
- `gitlab.com/golang-commonmark/html`
- `gitlab.com/golang-commonmark/linkify`
- `gitlab.com/golang-commonmark/markdown`
- `gitlab.com/golang-commonmark/mdurl`
- `gitlab.com/golang-commonmark/puny`
- `golang.org/x/net`
- `golang.org/x/sync`
- `golang.org/x/text`
- `nhooyr.io/websocket`

As a beginner, you do not need to memorize the indirect libraries.
Go downloads them because the main libraries depend on them.

## 5. Full folder structure

This is the current project shape:

```text
RAG_System_By_Golang/
|-- .env
|-- .gitattributes
|-- .gitignore
|-- Diagram.drawio
|-- go.mod
|-- go.sum
|-- main.go
|-- README.md
|-- chunker/
|   |-- chunker.go
|   |-- chunker_test.go
|-- config/
|   |-- dotenv.go
|-- embedding/
|   |-- embedding.go
|   |-- embedding_test.go
|-- llm/
|   |-- llm.go
|-- PDF/
|   |-- unit 1 girma pdf.pdf
|   |-- unit 2 girm pdf.pdf
|   |-- unit 3 girm pdf.pdf
|   |-- unit 4 girm pdf.pdf
|   |-- unit 5 girm pdf.pdf
|   |-- unit 6 Missed unit for girm pdf.pdf
|-- rag/
|   |-- rag.go
|-- vectorstore/
|   |-- inmemory.go
|   |-- inmemory_test.go
|   |-- postgres.go
|   |-- vectorstore.go
|-- .gocache/   generated cache folder, not core project logic
```

## 6. What each folder is for

### `PDF/`

This folder keeps the source documents.

The whole RAG system starts here.
If there are no PDF files here, the system has nothing useful to learn from.

### `config/`

This folder helps the project read settings from `.env`.

It helps the app know things like:

- which LM Studio URL to use
- which model names to use
- whether to use memory or PostgreSQL

### `chunker/`

This folder cleans broken PDF text and cuts it into smaller chunks.

Why this folder matters:

- PDF text is often messy
- some words stick together
- some spaces are missing
- long text must be broken into smaller pieces before embeddings

This folder helps the rest of the system by giving clean chunks.

### `embedding/`

This folder turns text into embeddings.

It does not save data.
It only talks to the embedding API and gets vectors back.

This folder helps `rag` and `vectorstore`.

### `llm/`

This folder talks to the chat model.

It sends messages and gets the final answer text.

This folder helps `rag` finish the answer.

### `vectorstore/`

This folder stores embeddings and searches them later.

It supports two backends:

- in-memory store
- PostgreSQL store

This folder is the searchable memory of the project.

### `rag/`

This folder is the main coordinator.

It connects:

- `embedding`
- `vectorstore`
- `llm`

This is the package that gives you the 2 main actions:

- index documents
- ask questions

## 7. How folders help each other

This is the connection map:

```text
main
 |- config
 |- PDF folder
 |- chunker
 |- rag

rag
 |- embedding
 |- vectorstore
 |- llm

vectorstore
 |- memory backend
 |- postgres backend
```

And here is the same idea in simple words:

- `main` starts everything
- `config` gives settings to `main` and other packages through environment variables
- `main` reads PDF files and sends raw text to `chunker`
- `chunker` returns clean chunks
- `main` sends chunks to `rag`
- `rag` uses `embedding` to create vectors
- `rag` uses `vectorstore` to save and search vectors
- `rag` uses `llm` to write the final answer

## 8. Important data types in this project

These structs are very important because data moves through them:

### `chunker.Chunk`

Fields:

- `Text`
- `FileName`
- `StartWord`
- `EndWord`

Meaning:

- stores one text chunk and where it came from

### `embedding.Embedding`

Fields:

- `Chunk`
- `Vector`
- `ModelName`

Meaning:

- stores one chunk together with its embedding vector

### `vectorstore.SearchResult`

Fields:

- `Chunk`
- `Score`
- `Position`

Meaning:

- stores one search match and how similar it is

### `rag.RAGSystem`

Fields:

- `Embedder`
- `VectorStore`
- `LLM`

Meaning:

- this is the main object that holds the 3 core workers together

## 9. Root files explained

### `main.go`

This is the entry point of the whole application.

It does these jobs:

- load environment variables
- create the RAG system
- read PDF files from `PDF/`
- extract PDF text
- call the chunker
- index the chunks
- start a question loop in the terminal

This file directly talks to:

- `config`
- `chunker`
- `rag`
- `langchaingo` PDF loader

### `go.mod`

This file says:

- project module name is `MyRagByCivic`
- Go version is `1.25.1`
- which libraries are required

### `go.sum`

This file stores dependency checksums.

You usually do not edit it by hand.

### `.env`

This file stores settings for your app.

Example:

- base URL for LM Studio
- model names
- database settings

### `.gitignore`

This file tells Git which files should not be tracked.

### `.gitattributes`

This file stores Git-related file rules.

### `Diagram.drawio`

This is likely a visual diagram for the project.
It is not required for running the code.

### `.gocache/`

This is generated by Go tooling.
It is not part of the real business logic.
You can mostly ignore it when learning the project.

## 10. File by file explanation

### `config/dotenv.go`

Purpose:

- read a local `.env` file
- skip empty lines and comments
- load `KEY=VALUE` into environment variables

How it helps other files:

- `main.go` calls it at startup
- other packages later read those environment variables

### `chunker/chunker.go`

Purpose:

- clean messy PDF text
- fix common joined words
- split text into overlapping chunks

How it helps other files:

- `main.go` calls `chunker.SliceText(...)`
- `embedding` and `vectorstore` later use the resulting chunks

### `chunker/chunker_test.go`

Purpose:

- check that text cleaning works
- check that chunking happens after cleaning

How it helps:

- protects the chunker from breaking

### `embedding/embedding.go`

Purpose:

- create an embedding client
- call `/embeddings`
- return vectors

How it helps other files:

- `rag.IndexDocuments(...)` uses it for document chunks
- `rag.Ask(...)` uses it for the user question

### `embedding/embedding_test.go`

Purpose:

- test client creation
- test live embedding calls if LM Studio is running

Important note:

- some tests skip if LM Studio is not available

### `llm/llm.go`

Purpose:

- create a chat client
- call `/chat/completions`
- return the model answer text

How it helps other files:

- `rag.Ask(...)` uses it after retrieval is done

### `rag/rag.go`

Purpose:

- create the whole RAG system
- index chunks
- answer questions

Why this file matters a lot:

- this is the center of the project logic
- this file joins storage, embeddings, and final answer generation

### `vectorstore/vectorstore.go`

Purpose:

- define the vector store interface
- load vector store config from environment variables
- choose the backend

How it helps other files:

- `rag.NewRAGSystem(...)` calls `vectorstore.NewStore(...)`

### `vectorstore/inmemory.go`

Purpose:

- save embeddings in memory
- search them using cosine similarity

Best use:

- local testing
- simple runs without PostgreSQL

### `vectorstore/inmemory_test.go`

Purpose:

- check replace behavior
- check model filtering during search

### `vectorstore/postgres.go`

Purpose:

- connect to PostgreSQL
- register `pgvector`
- auto-create table schema with GORM
- insert embeddings
- search embeddings in SQL

Important note:

- PostgreSQL mode needs the `vector` extension enabled first
- this file imports GORM packages that are not yet added to `go.mod`

## 11. Function by function map

This section is short on purpose.
The goal is to tell you what each function does in one simple line.

### In `main.go`

- `main()`: starts the app, indexes PDFs, then waits for user questions

### In `config/dotenv.go`

- `LoadDotEnv(path)`: reads `.env` file lines and puts them into environment variables

### In `chunker/chunker.go`

- `SliceText(text, size, overlap, filename)`: cleans full text and splits it into chunks
- `cleanText(text)`: fixes bad PDF text before chunking
- `splitCommonSuffixWords(text)`: fixes common glued words like `Meaningof`
- `splitIntoWords(text)`: splits text into words

### In `chunker/chunker_test.go`

- `TestCleanText_FixesPDFWordGlue(...)`: checks word-fixing behavior
- `TestSliceText_CleansBeforeChunking(...)`: checks that cleaning happens before chunking

### In `embedding/embedding.go`

- `NewClient(modelName)`: creates the embedding API client
- `GetEmbedding(ctx, text)`: sends one text to the embedding API and gets one vector
- `GetEmbeddingsForChunks(ctx, chunks)`: loops over chunks and gets vectors for all of them

### In `embedding/embedding_test.go`

- `testModelName()`: picks test model name from env or default
- `TestNewClient(...)`: checks client setup values
- `TestGetEmbedding(...)`: tests one live embedding request
- `TestGetEmbeddingsForChunks(...)`: tests many live embedding requests

### In `llm/llm.go`

- `NewClient(modelName)`: creates the chat API client
- `Generate(ctx, messages)`: sends chat messages and returns the answer text

### In `rag/rag.go`

- `NewRAGSystem(ctx, embedModel, llmModel)`: creates embedder, vector store, and LLM client
- `Close()`: closes the vector store if needed
- `IndexDocuments(ctx, chunks)`: turns chunks into embeddings and stores them
- `Ask(ctx, question)`: embeds the question, searches chunks, builds prompt, gets final answer
- `getUniqueSources(sources)`: removes duplicate file names
- `buildSourceFooter(chunks)`: builds the final source list shown under the answer

### In `vectorstore/vectorstore.go`

- `LoadConfigFromEnv()`: reads vector store settings from environment variables
- `NewStore(ctx)`: creates store using env config
- `NewStoreWithConfig(ctx, cfg)`: creates the chosen backend
- `normalizeTopK(topK)`: makes sure topK is at least 1
- `sanitizeIdentifier(name)`: protects SQL table name format
- `getEnvInt32(name, fallback)`: reads integer env values safely
- `sourceKey(fileName, modelName)`: makes one unique key from file name and model name

### In `vectorstore/inmemory.go`

- `NewInMemoryStore()`: creates the memory store
- `Add(ctx, embeddings)`: saves embeddings and replaces old chunks for the same source/model
- `Search(ctx, modelName, queryVector, topK)`: finds the most similar chunks in memory
- `Close()`: does nothing for memory store
- `cosineSimilarity(a, b)`: calculates similarity score between 2 vectors

### In `vectorstore/postgres.go`

- `NewPostgresStore(ctx, cfg)`: creates PostgreSQL store and validates setup
- `Add(ctx, embeddings)`: deletes old rows for the same source/model, then inserts new rows
- `Search(ctx, modelName, queryVector, topK)`: searches nearest vectors from PostgreSQL
- `Close()`: closes PostgreSQL connection pool
- `validateSetup(ctx)`: checks database connection, extension, and table
- `clampScore(score)`: keeps score between 0 and 1

## 12. Default values you should know

These default values are important:

- PDF folder: `./PDF`
- chunk size: `450`
- chunk overlap: `90`
- search result count: `3`
- embedding base URL default: `http://127.0.0.1:1234/v1`
- embedding model default: `text-embedding-nomic-embed-text-v1.5`
- chat model default: `qwen3.5-0.8b`
- LLM temperature: `0.7`
- LLM max tokens: `300`
- default vector table name: `rag_chunks`

## 13. Startup flow from function to function

This is what happens when you run:

```bash
go run main.go
```

### Step by step

1. `main()` starts
2. `config.LoadDotEnv(".env")` tries to read local env values
3. `godotenv.Load()` also tries to load `.env`
4. `context.Background()` is created
5. `rag.NewRAGSystem(ctx, "", "")` is called
6. inside that, `vectorstore.NewStore(ctx)` is called
7. inside that, `LoadConfigFromEnv()` reads env settings
8. inside that, `NewStoreWithConfig(...)` chooses memory or PostgreSQL
9. still inside `NewRAGSystem(...)`, `embedding.NewClient(...)` is created
10. still inside `NewRAGSystem(...)`, `llm.NewClient(...)` is created
11. back in `main()`, the app reads files from `./PDF`
12. for each PDF file, it opens the file
13. `documentloaders.NewPDF(...)` is used to read PDF text
14. all page text is joined into one big string
15. `chunker.SliceText(...)` cleans and splits the text
16. `rag.IndexDocuments(ctx, chunks)` is called
17. inside that, `Embedder.GetEmbeddingsForChunks(...)` is called
18. inside that, each chunk uses `GetEmbedding(...)`
19. when all vectors are ready, `VectorStore.Add(...)` saves them
20. after all PDFs finish, the terminal question loop starts

## 14. Question flow from function to function

This is what happens when you type a question:

1. terminal reads your text in `main()`
2. `ragSystem.Ask(ctx, question)` is called
3. `Embedder.GetEmbedding(ctx, question)` turns the question into a vector
4. `VectorStore.Search(ctx, rag.Embedder.ModelName, questionVec, 3)` finds the best chunks
5. `rag.Ask(...)` builds a text context from those chunks
6. `rag.Ask(...)` creates a prompt with source instructions
7. `LLM.Generate(ctx, messages)` sends the final request to the chat model
8. chat model returns the answer
9. `buildSourceFooter(...)` adds source names and previews
10. `main()` prints the answer in the terminal

## 15. PDF to answer flow in very simple language

This is the clean story from PDF material to final answer:

### Part 1. PDF becomes plain text

- the code opens each PDF file
- the PDF loader reads pages
- each page gives text
- all page text is joined together

### Part 2. Plain text becomes small chunks

- PDF text is often ugly and broken
- `chunker.cleanText(...)` fixes spacing and glued words
- `chunker.SliceText(...)` cuts the text into smaller overlapping pieces

Why overlap is used:

- one chunk shares some words with the next chunk
- this helps keep meaning when a sentence crosses chunk borders

### Part 3. Each chunk becomes numbers

- `embedding.GetEmbeddingsForChunks(...)` sends each chunk to the embedding model
- the model returns a vector for each chunk
- now the computer can compare chunk meaning using numbers

### Part 4. Vectors are saved

- `vectorstore.Add(...)` stores the vectors
- if memory backend is used, data stays in RAM only
- if PostgreSQL backend is used, data goes into the database table

### Part 5. Your question also becomes numbers

- when you ask something, your question is also sent to the embedding model
- the question gets its own vector

### Part 6. Similar chunks are found

- the vector store compares the question vector with chunk vectors
- the most similar chunks are returned

This is the retrieval part of RAG.

### Part 7. The final answer is written

- those selected chunks are placed into a prompt
- the prompt also tells the model to cite source file names
- `llm.Generate(...)` sends the prompt to the chat model
- the chat model writes the final answer

### Part 8. You see the answer with source info

- the answer is printed in the terminal
- a source footer is added under it
- this helps you know which PDF file the answer came from

## 16. How the in-memory vector store works

This is the simple backend.

Good for:

- quick testing
- learning
- running without PostgreSQL

How it works:

- keep embeddings in a Go slice
- when new chunks for the same file/model arrive, replace old ones
- on search, compare vectors using `cosineSimilarity(...)`
- sort by score
- return top results

Weak side:

- data disappears when the program stops

## 17. How the PostgreSQL vector store works

This is the database backend.

Good for:

- saving vectors between program runs
- bigger projects
- database-based search

How it works:

1. read database config from env
2. connect using `pgxpool`
3. register `pgvector` type support
4. open GORM
5. run `AutoMigrate` for the chunk table
6. verify PostgreSQL connection and `vector` extension
7. insert embeddings as rows
8. search with pgvector distance

Important things you must know:

- PostgreSQL mode needs the `vector` extension enabled
- the table name must be safe SQL text
- the `RagChunk` schema uses `vector(1536)`
- if your embedding model dimension is different, the schema may need change

## 18. Environment variables explained

These are the environment variables used in the code:

### LM Studio settings

- `LM_STUDIO_BASE_URL`: base API URL, default is `http://127.0.0.1:1234/v1`
- `LM_STUDIO_EMBEDDING_MODEL`: embedding model name
- `LM_STUDIO_CHAT_MODEL`: chat model name

### Vector backend choice

- `RAG_VECTOR_BACKEND`: choose `memory` or `postgres`

### Full database connection options

- `DATABASE_URL`: full PostgreSQL DSN
- `POSTGRES_DSN`: another full PostgreSQL DSN option

### Database pieces if you do not use one full DSN

- `DB_HOST`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `DB_PORT`
- `DB_SSLMODE`
- `PGSSLMODE`

### PostgreSQL vector table settings

- `RAG_VECTOR_TABLE`: table name, default is `rag_chunks`
- `RAG_PG_MAX_OPEN_CONNS`: max open DB connections
- `RAG_PG_MAX_IDLE_CONNS`: max idle DB connections

## 19. Example `.env` file

Use memory mode:

```env
LM_STUDIO_BASE_URL=http://127.0.0.1:1234/v1
LM_STUDIO_EMBEDDING_MODEL=text-embedding-nomic-embed-text-v1.5
LM_STUDIO_CHAT_MODEL=qwen3.5-0.8b
RAG_VECTOR_BACKEND=memory
```

Use PostgreSQL mode with one DSN:

```env
LM_STUDIO_BASE_URL=http://127.0.0.1:1234/v1
LM_STUDIO_EMBEDDING_MODEL=text-embedding-nomic-embed-text-v1.5
LM_STUDIO_CHAT_MODEL=qwen3.5-0.8b
RAG_VECTOR_BACKEND=postgres
DATABASE_URL=postgres://postgres:password@localhost:5432/ragdb?sslmode=disable
RAG_VECTOR_TABLE=rag_chunks
RAG_PG_MAX_OPEN_CONNS=10
RAG_PG_MAX_IDLE_CONNS=5
```

Use PostgreSQL mode with DB parts:

```env
LM_STUDIO_BASE_URL=http://127.0.0.1:1234/v1
LM_STUDIO_EMBEDDING_MODEL=text-embedding-nomic-embed-text-v1.5
LM_STUDIO_CHAT_MODEL=qwen3.5-0.8b
RAG_VECTOR_BACKEND=postgres
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ragdb
DB_PORT=5432
DB_SSLMODE=disable
RAG_VECTOR_TABLE=rag_chunks
```

## 20. How to run this project

### Simple memory mode

1. install Go
2. start LM Studio
3. load one embedding model in LM Studio
4. load one chat model in LM Studio
5. put your PDFs inside `PDF/`
6. create `.env`
7. run:

```bash
go run main.go
```

8. ask questions in the terminal
9. type `exit` to stop

### PostgreSQL mode

Do the same steps above, and also:

1. create PostgreSQL database
2. enable the `vector` extension
3. set database values in `.env`
4. make sure GORM dependencies are installed in the module

## 21. Real setup note found while reading the code

I checked the code and test status, and this is important:

- `go test ./...` currently fails because `gorm.io/gorm` and `gorm.io/driver/postgres` are imported but not yet added to `go.mod`
- `chunker` tests pass
- `embedding` tests pass or skip depending on LM Studio availability
- memory vector store tests pass

So the guide above explains the project correctly, but the repository still needs those two GORM packages added before a full build/test works cleanly.

## 22. Mental model to remember

If you forget the details, remember this:

### Indexing time

`PDF -> loader -> clean text -> chunks -> embeddings -> vector store`

### Question time

`question -> question embedding -> vector search -> prompt -> LLM answer -> source list`

That is the heart of this project.

## 23. Final short summary

This project is a Go RAG app that reads PDFs, cleans the text, breaks it into chunks, turns those chunks into embeddings, stores them, and later uses the most relevant chunks to answer your question with source names.

If you understand that one sentence, you already understand the core of the system.
