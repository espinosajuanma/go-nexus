package nexus

import (
	"strings"
	"testing"
)

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

	nex, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify TAXA block
	taxa, ok := GetBlock[*TaxaBlock](nex)
	if !ok {
		t.Fatal("Expected to find a TAXA block, but got none")
	}
	if taxa.Dimensions.Count != 2 {
		t.Errorf("Expected 2 taxa, got %d", taxa.Dimensions.Count)
	}
	if len(taxa.TaxLabels) != 2 || taxa.TaxLabels[0] != "fish" {
		t.Errorf("Unexpected taxa labels: %v", taxa.TaxLabels)
	}

	// Verify CHARACTERS block
	chars, ok := GetBlock[*CharactersBlock](nex)
	if !ok {
		t.Fatal("Expected to find a CHARACTERS block, but got none")
	}
	if chars.Dimensions.NChar != 10 {
		t.Errorf("Expected 10 characters, got %d", chars.Dimensions.NChar)
	}
	if chars.Format.DataType != "DNA" {
		t.Errorf("Expected DATATYPE=DNA, got %s", chars.Format.DataType)
	}
	if chars.Format.Gap != "-" {
		t.Errorf("Expected GAP=-, got %s", chars.Format.Gap)
	}
	if len(chars.Matrix) != 2 {
		t.Errorf("Expected 2 matrix rows, got %d", len(chars.Matrix))
	}
}

func TestParseInvalidHeader(t *testing.T) {
	input := `BEGIN TAXA;
		DIMENSIONS NTAX=2;
	END;`

	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("Expected error for missing #NEXUS header, but got none")
	}
	if !strings.Contains(err.Error(), "expected #NEXUS") {
		t.Errorf("Expected error about #NEXUS, got: %v", err)
	}
}
