package taxa_test

import (
	"strings"
	"testing"

	. "github.com/espinosajuanma/go-nexus/blocks/taxa"
	"github.com/espinosajuanma/go-nexus/scanner"
)

func TestTaxaBlock_Parse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectErr   bool
		errContains string
		check       func(*testing.T, *TaxaBlock)
	}{
		{
			name: "Valid Basic Block",
			input: `TITLE 'My Taxa';
					DIMENSIONS NTAX=3;
					TAXLABELS frog toad 'tree frog';
					END;`,
			expectErr: false,
			check: func(t *testing.T, tb *TaxaBlock) {
				if tb.Title != "My Taxa" {
					t.Errorf("Expected title 'My Taxa', got '%s'", tb.Title)
				}
				if tb.Dimensions != 3 {
					t.Errorf("Expected NTAX=3, got %d", tb.Dimensions)
				}
				if len(tb.TaxLabels) != 3 {
					t.Fatalf("Expected 3 labels, got %d", len(tb.TaxLabels))
				}
				// Verifying core.DecodeName logic happened correctly
				if tb.TaxLabels[2] != "tree frog" {
					t.Errorf("Expected 'tree frog', got '%s'", tb.TaxLabels[2])
				}
			},
		},
		{
			name: "Valid Block with Sets and Partitions",
			input: `DIMENSIONS NTAX=4;
					TAXLABELS t1 t2 t3 t4;
					TAXSET set1 = t1 t2;
					TAXPARTITION part1 = p1:t1 t2, p2:t3 t4;
					ENDBLOCK;`,
			expectErr: false,
			check: func(t *testing.T, tb *TaxaBlock) {
				if len(tb.TaxSets) != 1 {
					t.Errorf("Expected 1 TaxSet, got %d", len(tb.TaxSets))
				}
				if len(tb.TaxPartitions) != 1 {
					t.Errorf("Expected 1 TaxPartition, got %d", len(tb.TaxPartitions))
				}
			},
		},
		{
			name: "Error: TAXLABELS before DIMENSIONS",
			input: `TAXLABELS t1 t2; 
					DIMENSIONS NTAX=2; 
					END;`,
			expectErr:   true,
			errContains: "DIMENSIONS must be defined before TAXLABELS",
		},
		{
			name: "Error: NTAX mismatch (too many labels)",
			input: `DIMENSIONS NTAX=2; 
					TAXLABELS t1 t2 t3; 
					END;`,
			expectErr:   true,
			errContains: "dimension mismatch",
		},
		{
			name: "Error: NTAX mismatch (too few labels)",
			input: `DIMENSIONS NTAX=4; 
					TAXLABELS t1 t2 t3; 
					END;`,
			expectErr:   true,
			errContains: "dimension mismatch",
		},
		{
			name: "Error: Invalid NTAX value",
			input: `DIMENSIONS NTAX=-5; 
					TAXLABELS t1; 
					END;`,
			expectErr:   true,
			errContains: "positive integer",
		},
		{
			name: "Error: Duplicate dimensions command",
			input: `DIMENSIONS NTAX=2; 
					DIMENSIONS NTAX=2; 
					TAXLABELS t1 t2; 
					END;`,
			expectErr:   true,
			errContains: "DIMENSIONS command can only appear once",
		},
		{
			name: "Error: All-digit taxon name",
			input: `DIMENSIONS NTAX=2; 
					TAXLABELS t1 123; 
					END;`,
			expectErr:   true,
			errContains: "cannot consist entirely of digits",
		},
		{
			name: "Error: Duplicate taxon names (exact match)",
			input: `DIMENSIONS NTAX=3; 
					TAXLABELS frog toad frog; 
					END;`,
			expectErr:   true,
			errContains: "duplicate taxon name",
		},
		{
			name: "Error: Duplicate taxon names (case insensitive homonym)",
			input: `DIMENSIONS NTAX=2; 
					TAXLABELS frog FROG; 
					END;`,
			expectErr:   true,
			errContains: "duplicate taxon name",
		},
		{
			name: "Error: Duplicate taxon names (underscore equivalence)",
			input: `DIMENSIONS NTAX=2; 
					TAXLABELS tree_frog 'tree frog'; 
					END;`,
			expectErr:   true,
			errContains: "duplicate taxon name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup a new scanner with the test input
			s := scanner.NewScanner(strings.NewReader(tt.input))

			// Initialize an empty TaxaBlock
			tb := &TaxaBlock{}

			// Execute the parse function
			err := tb.Parse(s)

			// Assertions
			if tt.expectErr {
				if err == nil {
					t.Fatalf("Expected error containing '%s', but got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				// Run custom assertions if provided
				if tt.check != nil {
					tt.check(t, tb)
				}
			}
		})
	}
}
