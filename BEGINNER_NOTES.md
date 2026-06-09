# Beginner Notes For This Project

This project now runs as an HTTP API, so a frontend can call it with JSON.

## 1. What changed

- The old terminal question loop is gone.
- The app now starts a web server.
- A frontend can call `POST /api/v1/chat`.
- You can re-index PDFs with `POST /api/v1/index`.

## 2. How to read `something.method()`

When you see code like this:

```go
docs, err := loader.Load(ctx)
```

`loader` is a variable.
`Load` is a method called on that variable.

There are 3 common cases:

- `pkg.Func()` means a function from a package.
- `value.Method()` means a method on a value or struct.
- `TypeName{}` means a struct literal, not a function.

## 3. Is it built-in or custom?

Use this quick check:

- If it starts with a package name like `json.Marshal`, it is from a package.
- If it looks like `loader.Load`, it is usually a method on a value.
- If the type comes from your code, the method is yours.
- If the type comes from a library, the method is from that library.

Example from this repo:

- `loader.Load(ctx)` is from `github.com/tmc/langchaingo/documentloaders`
- `chunker.SliceText(...)` is our own function
- `service.Ask(...)` is our own method
- `ragSystem.Ask(...)` is our own method

## 4. How to know where a method comes from

In Go, follow the type first:

1. Find the variable before the dot.
2. Check where that variable was created.
3. Look at the type of that variable.
4. Open that type definition.

Example:

```go
loader := documentloaders.NewPDF(file, size)
docs, err := loader.Load(ctx)
```

- `documentloaders.NewPDF(...)` creates the loader.
- `loader` is a value returned by the library.
- `Load` belongs to that loader type.

## 5. Our custom methods in this project

- `(*Service).Ask(...)` asks the RAG system a question.
- `(*Service).IndexDocumentsFromPDF(...)` indexes all PDFs.
- `(*RAGSystem).Ask(...)` builds the prompt and gets an answer.
- `(*RAGSystem).IndexDocuments(...)` stores chunk embeddings.

## 6. Default values used in code

These are the safe defaults:

- `APP_PORT=8080`
- `PDF_DIR=./PDF`
- `CHUNK_SIZE=450`
- `CHUNK_OVERLAP=90`
- embedding base URL: `http://127.0.0.1:1234/v1`
- embedding model: `text-embedding-nomic-embed-text-v1.5`
- chat model: `qwen3.5-0.8b`

## 7. Frontend-ready API routes

- `GET /` returns a simple status message
- `GET /api/v1/health` returns readiness and indexing status
- `POST /api/v1/chat` accepts `{ "question": "..." }`
- `POST /api/v1/index` re-indexes the PDF folder

## 8. Short mental model

Think of the flow like this:

`PDF -> text -> chunks -> embeddings -> search -> answer`

That is the main idea behind the whole app.
