package parser_test

import (
	"strings"
	"testing"

	"github.com/espinosajuanma/nexus/blocks/characters"
	_ "github.com/espinosajuanma/nexus/blocks/generic"
	"github.com/espinosajuanma/nexus/blocks/taxa"
	"github.com/espinosajuanma/nexus/blocks/trees"
	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/parser"
)

// Verifies TAXA and CHARACTERS blocks parse correctly.
func TestParseValidFile(t *testing.T) {
	input := `#NEXUS
BEGIN TAXA;
	DIMENSIONS NTAX=2;
	TAXLABELS fish frog;
END;
BEGIN CHARACTERS;
	DIMENSIONS NCHAR=10;
	FORMAT DATATYPE=DNA GAP=-;
	MATRIX
	fish ACATA-GAGG
	frog ACTTA-GAGG;
END;`

	nex, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify TAXA block
	taxa, ok := core.GetBlock[*taxa.TaxaBlock](nex)
	if !ok {
		t.Fatal("Expected to find a TAXA block, but got none")
	}
	if taxa.Dimensions != 2 {
		t.Errorf("Expected 2 taxa, got %d", taxa.Dimensions)
	}
	if len(taxa.TaxLabels) != 2 || taxa.TaxLabels[0] != "fish" {
		t.Errorf("Unexpected taxa labels: %v", taxa.TaxLabels)
	}

	// Verify CHARACTERS block
	chars, ok := core.GetBlock[*characters.CharactersBlock](nex)
	if !ok {
		t.Fatal("Expected to find a CHARACTERS block, but got none")
	}
	if chars.Dimensions != 10 {
		t.Errorf("Expected 10 characters, got %d", chars.Dimensions)
	}
	if chars.Format.DataType != "DNA" {
		t.Errorf("Expected DATATYPE=DNA, got %s", chars.Format.DataType)
	}
	if chars.Format.Gap != "-" {
		t.Errorf("Expected GAP=-, got %s", chars.Format.Gap)
	}
	if len(chars.Matrix.Taxa) != 2 {
		t.Errorf("Expected 2 matrix rows (taxa), got %d", len(chars.Matrix.Taxa))
	}
}

// Verifies missing header throws an error.
func TestParseInvalidHeader(t *testing.T) {
	input := `BEGIN TAXA;
	DIMENSIONS NTAX=2;
END;`

	_, err := parser.Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("Expected error for missing #NEXUS header, but got none")
	}
	if !strings.Contains(err.Error(), "expected #NEXUS") {
		t.Errorf("Expected error about #NEXUS, got: %v", err)
	}
}

// Verifies TREES block parsing.
func TestParseTreesBlock(t *testing.T) {
	input := `#NEXUS
BEGIN TREES;
	TRANSLATE
		1 fish,
		2 frog;
	TREE * tree1 = [&R] (1,2);
END;`

	nex, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	treesBlock, ok := core.GetBlock[*trees.TreesBlock](nex)
	if !ok {
		t.Fatal("Expected to find a TREES block, but got none")
	}

	if len(treesBlock.Translate) != 2 {
		t.Errorf("Expected 2 translated taxa, got %d", len(treesBlock.Translate))
	}
	if len(treesBlock.Trees) != 1 {
		t.Errorf("Expected 1 tree, got %d", len(treesBlock.Trees))
	}
	if !treesBlock.Trees[0].IsDefault {
		t.Errorf("Expected tree to be marked as default (with *)")
	}
}

// Verifies that unknown blocks are safely skipped via SkipBlock.
func TestSkipUnknownBlocks(t *testing.T) {
	input := `#NEXUS
BEGIN UNKNOWN_BLOCK;
	SOME RANDOM COMMANDS;
	AND MORE STUFF;
END;
BEGIN TAXA;
	DIMENSIONS NTAX=1;
	TAXLABELS fish;
END;`

	nex, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Ensure the parser didn't crash and still found the valid TAXA block
	taxa, ok := core.GetBlock[*taxa.TaxaBlock](nex)
	if !ok {
		t.Fatal("Expected to find a TAXA block after skipping unknown block")
	}
	if taxa.Dimensions != 1 {
		t.Errorf("Expected 1 taxon, got %d", taxa.Dimensions)
	}
}

// Verifies strict semicolon enforcement.
func TestMissingSemicolonErrors(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		errContains string
	}{
		{
			name:        "Missing semicolon after BEGIN",
			input:       "#NEXUS\nBEGIN TAXA\nDIMENSIONS NTAX=2;\nEND;",
			errContains: "expected ';'",
		},
		{
			name:        "Missing semicolon after END",
			input:       "#NEXUS\nBEGIN TAXA;\nDIMENSIONS NTAX=2;\nEND BEGIN",
			errContains: "expected ';'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(strings.NewReader(tt.input))
			if err == nil {
				t.Fatal("Expected an error for missing semicolon, but got none")
			}
			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Expected error to contain '%s', got: %v", tt.errContains, err)
			}
		})
	}
}
