package scanner

import (
	"io"
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Basic words and punctuation",
			input:    "BEGIN TAXA;",
			expected: []string{"BEGIN", "TAXA", ";"},
		},
		{
			name:     "Comments are ignored",
			input:    "TAXLABELS [ignore me] fish frog;",
			expected: []string{"TAXLABELS", "fish", "frog", ";"},
		},
		{
			name:     "Nested comments are ignored",
			input:    "DIMENSIONS NTAX=4 [nested [comments] work];",
			expected: []string{"DIMENSIONS", "NTAX", "=", "4", ";"},
		},
		{
			name:     "Quoted words",
			input:    `'John''s sparrow' frog`,
			expected: []string{"John's sparrow", "frog"},
		},
		{
			name:     "Matrix with mixed spacing",
			input:    "fish\tACATA\n\tfrog\tACTTA;",
			expected: []string{"fish", "ACATA", "frog", "ACTTA", ";"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(strings.NewReader(tt.input))
			var tokens []string

			for {
				tok, err := s.NextToken()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				tokens = append(tokens, tok)
			}

			if len(tokens) != len(tt.expected) {
				t.Fatalf("Expected %d tokens, got %d (%v)", len(tt.expected), len(tokens), tokens)
			}

			for i, expected := range tt.expected {
				if tokens[i] != expected {
					t.Errorf("Token %d mismatch: expected '%s', got '%s'", i, expected, tokens[i])
				}
			}
		})
	}
}
