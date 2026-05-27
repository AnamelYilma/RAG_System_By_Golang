package chunker

import "testing"

func TestCleanText_FixesPDFWordGlue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "splits common glued suffix words",
			input: "Civicsand Ethics Understandingthe Meaningof city-staterespectively",
			want:  "Civics and Ethics Understanding the Meaning of city-state respectively",
		},
		{
			name:  "splits lowercase uppercase and punctuation joins",
			input: "1.UnderstandingCivicsandEthics,Democracy:HumanRights",
			want:  "1. Understanding Civics and Ethics, Democracy: Human Rights",
		},
		{
			name:  "fixes known OCR typo",
			input: "CHAPTER-TWO2.Understnding State & Government",
			want:  "CHAPTER-TWO2. Understanding State & Government",
		},
		{
			name:  "does not split valid words that happen to end with short syllables",
			input: "Latin association constitution civis civitas",
			want:  "Latin association constitution civis civitas",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := cleanText(tt.input)
			if got != tt.want {
				t.Fatalf("cleanText(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSliceText_CleansBeforeChunking(t *testing.T) {
	t.Parallel()

	chunks := SliceText("Civicsand Ethics Understandingthe Meaningof", 10, 0, "sample.pdf")
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}

	want := "Civics and Ethics Understanding the Meaning of"
	if chunks[0].Text != want {
		t.Fatalf("chunk text = %q, want %q", chunks[0].Text, want)
	}
}
