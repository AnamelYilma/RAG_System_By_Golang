# Chunker Note

## What this file does

`chunker.go` is the text cleaner and text cutter for your RAG project.

Its job is simple:

1. Take raw text from a PDF.
2. Fix common PDF text problems.
3. Split the text into words.
4. Build small text chunks with overlap.
5. Return those chunks so the next part of the RAG system can use them.

So this file helps turn messy PDF text into clean, usable pieces.

## The full idea in one small story

Imagine you have one very long paragraph from a PDF.

That paragraph is usually messy:

- some words are joined together
- some spaces are missing
- some OCR spelling is wrong
- some punctuation looks strange

This file cleans that paragraph first.
After that, it cuts the paragraph into smaller pieces called `chunks`.
Each chunk keeps a little overlap with the next one, so meaning is not lost.

## What each part is doing

### `package chunker`

This says the file belongs to the `chunker` package.
Other files can call it like this:

```go
chunker.SliceText(...)
```

### Imports

```go
import (
	"regexp"
	"strings"
	"unicode"
)
```

These are helper tools:

- `regexp` is used to find text patterns
- `strings` is used to replace, join, and split text
- `unicode` is used to detect spaces safely

### Regex variables and `commonPDFTypos`

This block prepares text-fixing rules one time:

```go
var (
    ...
)
```

Why this is useful:

- PDF text is often broken
- some words get glued together
- some punctuation has no space after it
- some bullet symbols touch words
- some OCR mistakes happen

Examples:

- `HumanRights` becomes `Human Rights`
- `Meaningof` becomes `Meaning of`
- `1.Understanding` becomes `1. Understanding`
- `Understnding` becomes `Understanding`

This part is like the repair box of the file.

### `type Chunk struct`

```go
type Chunk struct {
	Text      string
	FileName  string
	StartWord int
	EndWord   int
}
```

This struct stores one final chunk.

What each field means:

- `Text`: the real text inside the chunk
- `FileName`: which PDF this chunk came from
- `StartWord`: where the chunk starts in the word list
- `EndWord`: where the chunk ends in the word list

So a chunk is not only text. It also keeps useful info about where that text came from.

### `SliceText(...)`

This is the main function of the file.

```go
func SliceText(text string, size, overlap int, filename string) []Chunk
```

It does the main work in this order:

1. Clean the text with `cleanText`
2. Split the clean text into words
3. Move through the words step by step
4. Build chunks with the chosen size
5. Keep overlap between nearby chunks
6. Return all chunks

### Inside `SliceText`

#### 1. Clean the text

```go
text = cleanText(text)
```

Before cutting text, the code fixes messy PDF problems.
This is smart, because bad text would create bad chunks.

#### 2. Split into words

```go
words := splitIntoWords(text)
```

Now the text becomes a word list.
Chunking by words is easier and safer than chunking by random characters.

#### 3. Create `step`

```go
step := size - overlap
if step < 1 {
	step = 1
}
```

This decides how far the loop moves each time.

Example:

- if `size = 450`
- and `overlap = 90`
- then `step = 360`

That means:

- first chunk uses words `0 -> 450`
- next chunk starts at word `360`

So the two chunks share `90` words.

This shared part helps keep context.

The `if step < 1` check protects the program from bad input.

#### 4. Loop over the words

```go
for i := 0; i < len(words); i += step {
```

This loop moves through the word list and creates one chunk at a time.

#### 5. Find the end of the chunk

```go
end := i + size
if end > len(words) {
	end = len(words)
}
```

This makes sure the chunk does not go past the last word.

#### 6. Build chunk text

```go
chunkWords := words[i:end]
chunkText := strings.Join(chunkWords, " ")
```

Here the code takes a small piece of the word list and joins it back into normal text.

#### 7. Save the chunk

```go
chunks = append(chunks, Chunk{
	Text:      chunkText,
	FileName:  filename,
	StartWord: i,
	EndWord:   end,
})
```

This stores the chunk in the final result list.

#### 8. Stop at the end

```go
if end == len(words) {
	break
}
```

When the code reaches the last word, it stops the loop.

## `cleanText(text string) string`

This function is the cleaner.
It prepares ugly PDF text before chunking.

### What it fixes

#### 1. Known OCR mistakes

```go
text = commonPDFTypos.Replace(text)
```

This corrects known wrong spellings, like:

- `Understnding` to `Understanding`

#### 2. Bad spacing around punctuation

```go
text = strings.ReplaceAll(text, " .", ".")
text = strings.ReplaceAll(text, " ,", ",")
text = strings.ReplaceAll(text, " :", ":")
```

This removes wrong spaces before punctuation.

Then these regex rules add missing spaces after punctuation or around bullet marks:

```go
text = punctuationSpacingRE.ReplaceAllString(text, `$1 $2`)
text = beforeBulletRE.ReplaceAllString(text, `$1 $2`)
text = afterBulletRE.ReplaceAllString(text, `$1 $2`)
```

#### 3. Words glued together

```go
text = lowerUpperSpacingRE.ReplaceAllString(text, `$1 $2`)
text = splitCommonSuffixWords(text)
```

This fixes cases where two words were pushed together.

Examples:

- `HumanRights` -> `Human Rights`
- `Meaningof` -> `Meaning of`
- `Civicsand` -> `Civics and`

#### 4. Extra spaces, tabs, and line breaks

```go
text = strings.Map(func(r rune) rune {
	if unicode.IsSpace(r) {
		return ' '
	}
	return r
}, text)
```

This changes all space-like characters into normal spaces.

That means:

- new lines
- tabs
- mixed spacing

all become simple spaces.

Then this part removes extra spaces:

```go
text = strings.TrimSpace(text)

for strings.Contains(text, "  ") {
	text = strings.ReplaceAll(text, "  ", " ")
}
```

So the final text becomes clean and tidy.

## `splitCommonSuffixWords(text string) string`

This helper function keeps fixing glued words until no more fixes are needed.

```go
for {
	updated := text
	...
	if updated == text {
		return updated
	}
	text = updated
}
```

Why use a loop here?

Because one text may have many glued words.
The function keeps checking again and again until the text stops changing.

This is a nice small safety idea.

## `splitIntoWords(text string) []string`

```go
return strings.Fields(text)
```

This is the simplest function in the file.

It splits clean text into words and automatically ignores extra spaces.

## How `chunker.go` finishes its work

This file finishes its job in this order:

1. Get raw text
2. Clean broken PDF text
3. Split text into words
4. Make chunk windows with overlap
5. Save each chunk with file name and word position
6. Return `[]Chunk`

So the output is ready for the next RAG steps like embedding, search, or retrieval.

## Small beginner review

This file is good because:

- it does one clear job
- it cleans text before chunking
- it keeps overlap for better context
- it saves helpful metadata with each chunk

In short:

`chunker.go` turns messy PDF text into clean, organized, overlapping text pieces.
That is why it is an important helper file in your RAG system.
