package generic_test

import (
	"strings"
	"testing"

	"github.com/espinosajuanma/nexus/blocks/generic"
	"github.com/espinosajuanma/nexus/scanner"
)

func TestGenericBlock_Parse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:     "Basic generic block",
			input:    "\n\tThis is a generic block.\n\tIt should get the same content\nEND;",
			expected: "\tThis is a generic block.\n\tIt should get the same content",
		},
		{
			name:     "User specific SARASA example",
			input:    "\n    This is a generic block.\n    It should get the same content\n\n    With spaces and 'normal';\n    content;\n    ;\n    ;\nEND;",
			expected: "    This is a generic block.\n    It should get the same content\n\n    With spaces and 'normal';\n    content;\n    ;\n    ;",
		},
		{
			name:     "END inside quotes is safely ignored",
			input:    "\n\tText with 'an END; inside quotes';\n\tAnd normal content;\nEND;",
			expected: "\tText with 'an END; inside quotes';\n\tAnd normal content;",
		},
		{
			name:     "END inside comments is safely ignored",
			input:    "\n\tText with [an END; inside a comment];\n\tMore text;\nEND;",
			expected: "\tText with [an END; inside a comment];\n\tMore text;",
		},
		{
			name:     "Nested comments with END are safely ignored",
			input:    "\n\tText [[ nested END; comment ]] here;\nEND;",
			expected: "\tText [[ nested END; comment ]] here;",
		},
		{
			name:     "Preserves inner whitespace and spacing exactly",
			input:    "\n\tLine 1;\n\n\n\tLine 2;\n\nEND;",
			expected: "\tLine 1;\n\n\n\tLine 2;",
		},
		{
			name:     "Handles ENDBLOCK as a terminator",
			input:    "\n\tSome data goes here;\nENDBLOCK;",
			expected: "\tSome data goes here;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize the generic block
			block := &generic.GenericBlock{
				Name: "TESTBLOCK",
			}

			// Create a scanner for the input string (simulating reading inside a block)
			s := scanner.NewScanner(strings.NewReader(tt.input))

			// Execute Parse
			err := block.Parse(s)

			// Handle expected errors
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected an error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify the parsed content matches the expected output exactly
			if block.Content != tt.expected {
				t.Errorf("Content mismatch.\nExpected:\n%q\nGot:\n%q", tt.expected, block.Content)
			}
		})
	}
}
