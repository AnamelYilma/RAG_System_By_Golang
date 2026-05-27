package chunker

import (
	"regexp"
	"strings"
	"unicode"
)

// =============================================
// Global variables (created once when program starts)
// These are like "tools" that help clean messy PDF text
// =============================================

// These regex patterns fix common PDF problems (built once for speed)
var (
	punctuationSpacingRE = regexp.MustCompile(`([.,:;!?])(\p{L})`)
	lowerUpperSpacingRE  = regexp.MustCompile(`(\p{Ll})(\p{Lu})`)
	beforeBulletRE       = regexp.MustCompile(`([\p{L}\p{N}])([❖➢✓])`)
	afterBulletRE        = regexp.MustCompile(`([❖➢✓])([\p{L}\p{N}])`)
	gluedPluralAndRE     = regexp.MustCompile(`([\p{L}]{4,}s)(and)\b`)
	gluedTheRE           = regexp.MustCompile(`([\p{L}-]{7,})(the)\b`)
	gluedOfRE            = regexp.MustCompile(`([\p{L}-]{7,})(of)\b`)
	gluedRespectivelyRE  = regexp.MustCompile(`([\p{L}-]{4,})(respectively)\b`)
	

	// Fixes common typing mistakes from PDF extraction
	commonPDFTypos = strings.NewReplacer(
		"Understnding", "Understanding",
	)
)

// Chunk is one piece of text + useful information
// Goal: Store each chunk so we can later create embeddings and show source
type Chunk struct {
	Text      string // The actual text content
	FileName  string // Which PDF it came from
	StartWord int    // Starting word position (useful for debugging)
	EndWord   int    // Ending word position
}

// SliceText is the main function of this file
// Goal: Take full PDF text → return many small overlapping chunks
func SliceText(text string, size, overlap int, filename string) []Chunk {
	text = cleanText(text)        // First clean the messy PDF text
	words := splitIntoWords(text) // Split into list of words

	var chunks []Chunk

	// We move forward by (size - overlap) so chunks share some words
	step := size - overlap
	if step < 1 {
		step = 1 // Prevent going backwards or zero step
	}

	// Create chunks by sliding window
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

// cleanText fixes common problems from PDF extraction
// Goal: Make text clean and readable before chunking
func cleanText(text string) string {
	// Step 1: Fix known spelling mistakes
	text = commonPDFTypos.Replace(text)

	// Step 2: Fix spacing problems
	text = strings.ReplaceAll(text, " .", ".")
	text = strings.ReplaceAll(text, " ,", ",")
	text = strings.ReplaceAll(text, " :", ":")

	text = punctuationSpacingRE.ReplaceAllString(text, `$1 $2`)
	text = beforeBulletRE.ReplaceAllString(text, `$1 $2`)
	text = afterBulletRE.ReplaceAllString(text, `$1 $2`)
	text = lowerUpperSpacingRE.ReplaceAllString(text, `$1 $2`)

	// Step 3: Fix glued words (very common in PDFs)
	text = splitCommonSuffixWords(text)

	// Step 4: Normalize all spaces
	text = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, text)

	text = strings.TrimSpace(text)

	// Remove double spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}

// splitCommonSuffixWords fixes words that got stuck together
func splitCommonSuffixWords(text string) string {
	for {
		updated := text
		updated = gluedPluralAndRE.ReplaceAllString(updated, `$1 $2`)
		updated = gluedTheRE.ReplaceAllString(updated, `$1 $2`)
		updated = gluedOfRE.ReplaceAllString(updated, `$1 $2`)
		updated = gluedRespectivelyRE.ReplaceAllString(updated, `$1 $2`)

		if updated == text {
			return updated
		}
		text = updated
	}
}

// splitIntoWords breaks text into words
func splitIntoWords(text string) []string {
	return strings.Fields(text)
}