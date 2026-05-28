package chunker

import (
	"regexp"
	"strings"
	"unicode"
)

// =============================================
// FILE PURPOSE: Text Cleaning + Chunking
// This file prepares raw PDF text for the RAG system.
// =============================================

// =============================================
// GLOBAL VARIABLES (Created once when package loads)
// These are "tools" that live for the whole program.
// Reason: Building regex is expensive, so we do it only once.
// =============================================

// These regex patterns fix common PDF extraction problems
var (
	// punctuationSpacingRE: Fixes "word." + "Next" → "word. Next"
	punctuationSpacingRE = regexp.MustCompile(`([.,:;!?])(\p{L})`)
	
	// lowerUpperSpacingRE: Fixes "CivicEducation" → "Civic Education"
	lowerUpperSpacingRE  = regexp.MustCompile(`(\p{Ll})(\p{Lu})`)
	
	beforeBulletRE       = regexp.MustCompile(`([\p{L}\p{N}])([❖➢✓])`)
	afterBulletRE        = regexp.MustCompile(`([❖➢✓])([\p{L}\p{N}])`)
	
	// These fix glued words like "lawsthe" → "laws the"
	gluedPluralAndRE     = regexp.MustCompile(`([\p{L}]{4,}s)(and)\b`)
	gluedTheRE           = regexp.MustCompile(`([\p{L}-]{7,})(the)\b`)
	gluedOfRE            = regexp.MustCompile(`([\p{L}-]{7,})(of)\b`)
	gluedRespectivelyRE  = regexp.MustCompile(`([\p{L}-]{4,})(respectively)\b`)

	// Fixes common OCR (scanning) mistakes
	commonPDFTypos = strings.NewReplacer(
		"Understnding", "Understanding",
	)
)

// =============================================
// DATA STRUCTURE
// =============================================

// Chunk holds one small piece of text + metadata
// Why we need this struct: So we can trace back where each chunk came from
type Chunk struct {
	Text      string // The actual cleaned text for this chunk
	FileName  string // Original PDF filename (for source citation)
	StartWord int    // Starting word index (useful for debugging)
	EndWord   int    // Ending word index
}

// =============================================
// MAIN FUNCTION
// =============================================

// SliceText is the main public function of this package
// What it does: Takes full PDF text → returns list of clean overlapping chunks
// Why: Large text cannot fit in LLM prompt. Small chunks help precise retrieval.
func SliceText(text string, size, overlap int, filename string) []Chunk {
	// Step 1: Clean messy PDF text first
	text = cleanText(text)
	
	// Step 2: Split into words
	words := splitIntoWords(text)

	var chunks []Chunk

	// Calculate step size for overlapping
	// Example: size=450, overlap=90 → step=360 (move forward 360 words each time)
	step := size - overlap
	if step < 1 {
		step = 1 // Safety: never go backwards
	}

	// Sliding window loop - creates overlapping chunks
	for i := 0; i < len(words); i += step {
		end := i + size
		if end > len(words) {
			end = len(words)
		}

		chunkWords := words[i:end]
		chunkText := strings.Join(chunkWords, " ")

		chunks = append(chunks, Chunk{
			Text:      chunkText,
			FileName:  filename,
			StartWord: i,
			EndWord:   end,
		})

		if end == len(words) {
			break
		}
	}

	return chunks
}

// =============================================
// HELPER FUNCTIONS
// =============================================

// cleanText fixes common PDF problems
// Why: PDF text from scanners is often broken
func cleanText(text string) string {
	// Fix known typos
	text = commonPDFTypos.Replace(text)

	// Fix spacing issues
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " ,", ",")
	text = strings.ReplaceAll(text, " :", ":")

	// Use regex to fix more complex spacing
	text = punctuationSpacingRE.ReplaceAllString(text, `$1 $2`)
	text = beforeBulletRE.ReplaceAllString(text, `$1 $2`)
	text = afterBulletRE.ReplaceAllString(text, `$1 $2`)
	text = lowerUpperSpacingRE.ReplaceAllString(text, `$1 $2`)

	// Fix glued words (very common issue)
	text = splitCommonSuffixWords(text)

	// Normalize all types of spaces (tab, newline, etc.) to single space
	text = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, text)

	text = strings.TrimSpace(text)

	// Remove multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}

// splitCommonSuffixWords fixes words stuck together like "lawsthe"
func splitCommonSuffixWords(text string) string {
	for {
		updated := text
		updated = gluedPluralAndRE.ReplaceAllString(updated, `$1 $2`)
		updated = gluedTheRE.ReplaceAllString(updated, `$1 $2`)
		updated = gluedOfRE.ReplaceAllString(updated, `$1 $2`)
		updated = gluedRespectivelyRE.ReplaceAllString(updated, `$1 $2`)

		// Stop when no more changes
		if updated == text {
			return updated
		}
		text = updated
	}
}

// splitIntoWords breaks text into list of words
// strings.Fields is smart: splits on any whitespace
func splitIntoWords(text string) []string {
	return strings.Fields(text)
}