package taxa_test

import (
	"strings"
	"testing"

	"github.com/espinosajuanma/go-nexus/blocks/taxa"
)

func TestTaxaBlock_Render(t *testing.T) {
	tests := []struct {
		name     string
		block    *taxa.TaxaBlock
		expected string
	}{
		{
			name: "Minimal Block",
			block: &taxa.TaxaBlock{
				Dimensions: 3,
				TaxLabels:  []string{"frog", "toad", "tree frog"},
			},
			expected: `BEGIN TAXA;
	DIMENSIONS NTAX=3;
	TAXLABELS
		frog
		toad
		'tree frog'
	;
END;`,
		},
		{
			name: "Block with Title",
			block: &taxa.TaxaBlock{
				Title:      "My Taxa",
				Dimensions: 2,
				TaxLabels:  []string{"A", "B"},
			},
			expected: `BEGIN TAXA;
	TITLE 'My Taxa';
	DIMENSIONS NTAX=2;
	TAXLABELS
		A
		B
	;
END;`,
		},
		{
			name: "Block with TaxSets (Standard)",
			block: &taxa.TaxaBlock{
				Dimensions: 4,
				TaxLabels:  []string{"t1", "t2", "t3", "t4"},
				TaxSets: map[string]taxa.TaxSet{
					"set1": {Format: taxa.StandardFormat, TaxaList: []string{"t1", "t2"}},
				},
			},
			expected: `BEGIN TAXA;
	DIMENSIONS NTAX=4;
	TAXLABELS
		t1
		t2
		t3
		t4
	;
	TAXSET set1 = t1 t2;
END;`,
		},
		{
			name: "Block with TaxSets (Vector)",
			block: &taxa.TaxaBlock{
				Dimensions: 4,
				TaxLabels:  []string{"t1", "t2", "t3", "t4"},
				TaxSets: map[string]taxa.TaxSet{
					"vset": {Format: taxa.VectorFormat, TaxaList: []string{"1", "1", "0", "0"}},
				},
			},
			expected: `BEGIN TAXA;
	DIMENSIONS NTAX=4;
	TAXLABELS
		t1
		t2
		t3
		t4
	;
	TAXSET vset (VECTOR) = 1 1 0 0;
END;`,
		},
		{
			name: "Block with TaxPartition (Standard)",
			block: &taxa.TaxaBlock{
				Dimensions: 4,
				TaxLabels:  []string{"t1", "t2", "t3", "t4"},
				TaxPartitions: map[string]taxa.TaxPartition{
					"part1": {
						Format: taxa.StandardFormat,
						Subsets: map[string][]string{
							"subA": {"t1", "t2"},
						},
					},
				},
			},
			expected: `BEGIN TAXA;
	DIMENSIONS NTAX=4;
	TAXLABELS
		t1
		t2
		t3
		t4
	;
	TAXPARTITION part1 =subA: t1 t2;
END;`,
		},
		{
			name: "Block with TaxPartition (Vector)",
			block: &taxa.TaxaBlock{
				Dimensions: 4,
				TaxLabels:  []string{"t1", "t2", "t3", "t4"},
				TaxPartitions: map[string]taxa.TaxPartition{
					"vpart": {
						Format: taxa.VectorFormat,
						Subsets: map[string][]string{
							"VECTOR_DATA": {"A", "A", "B", "B"},
						},
					},
				},
			},
			expected: `BEGIN TAXA;
	DIMENSIONS NTAX=4;
	TAXLABELS
		t1
		t2
		t3
		t4
	;
	TAXPARTITION vpart (VECTOR) = A A B B;
END;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := tt.block.Render()
			if err != nil {
				t.Fatalf("Unexpected error during Render(): %v", err)
			}

			// Normalize line endings to ensure tests pass on both Windows (\r\n) and Unix (\n)
			outNormalized := strings.TrimSpace(strings.ReplaceAll(out, "\r\n", "\n"))
			expectedNormalized := strings.TrimSpace(strings.ReplaceAll(tt.expected, "\r\n", "\n"))

			// Exact match comparison
			if outNormalized != expectedNormalized {
				t.Errorf("Render() output mismatch.\n\nExpected:\n%s\n\nGot:\n%s", expectedNormalized, outNormalized)
			}
		})
	}
}
