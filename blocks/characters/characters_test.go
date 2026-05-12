package characters_test

import (
	"strings"
	"testing"

	"github.com/espinosajuanma/nexus"
	"github.com/espinosajuanma/nexus/blocks/characters"
	"github.com/espinosajuanma/nexus/core"
)

// TestAddTaxonNilPointerFix ensures the CharactersBlock init and New functions
// properly set the Matrix back-pointer, preventing panics on AddTaxon.
func TestAddTaxonNilPointerFix(t *testing.T) {
	cb := characters.New(&core.Nexus{}, characters.Standard)

	// AddTaxon used to panic here because cb.Matrix.parent was nil
	taxon := cb.AddTaxon("Taxon1")

	if taxon.Name != "Taxon1" {
		t.Errorf("Expected taxon name 'Taxon1', got '%s'", taxon.Name)
	}
}

// TestParseInterleavedMatrix ensures that data split across multiple chunks
// merges correctly onto the same row based on taxon Index progression.
func TestParseInterleavedMatrix(t *testing.T) {
	nexusData := `#NEXUS
	BEGIN CHARACTERS;
		DIMENSIONS NCHAR=10;
		FORMAT DATATYPE=DNA INTERLEAVE;
		MATRIX
		  taxon1  ACGTA
		  taxon2  CGTAC

		  taxon1  TGCAA
		  taxon2  GCATG
		;
	END;`

	nex, err := nexus.Parse(strings.NewReader(nexusData))
	if err != nil {
		t.Fatalf("Failed to parse nexus string: %v", err)
	}

	charBlock, ok := core.GetBlock[*characters.CharactersBlock](nex)
	if !ok {
		t.Fatal("Failed to extract CharactersBlock")
	}

	t1 := charBlock.Matrix.GetTaxon("taxon1")
	if t1 == nil {
		t.Fatal("taxon1 not found")
	}

	// Verify the first character of chunk 1
	state1 := t1.GetState(charBlock.Matrix.GetCharacterByIndex(1))
	if state1.Values[0].Symbol != "A" {
		t.Errorf("Expected taxon1 char 1 to be 'A', got %v", state1.Values[0].Symbol)
	}

	// Verify the first character of chunk 2
	state6 := t1.GetState(charBlock.Matrix.GetCharacterByIndex(6))
	if state6.Values[0].Symbol != "T" {
		t.Errorf("Expected taxon1 char 6 to be 'T', got %v", state6.Values[0].Symbol)
	}
}

// TestParseMatchChar ensures that match characters (e.g., '.')
// correctly clone the expanded state from the first taxon.
func TestParseMatchChar(t *testing.T) {
	nexusData := `#NEXUS
	BEGIN CHARACTERS;
		DIMENSIONS NCHAR=4;
		FORMAT DATATYPE=DNA MATCHCHAR=.;
		MATRIX
		  taxon1  ACGT
		  taxon2  A..T
		;
	END;`

	nex, err := nexus.Parse(strings.NewReader(nexusData))
	if err != nil {
		t.Fatalf("Failed to parse nexus string: %v", err)
	}

	charBlock, ok := core.GetBlock[*characters.CharactersBlock](nex)
	if !ok {
		t.Fatal("Failed to extract CharactersBlock")
	}

	t2 := charBlock.Matrix.GetTaxon("taxon2")

	// Char 2 (index 2) was a matchchar '.', should copy taxon1's 'C'
	state2 := t2.GetState(charBlock.Matrix.GetCharacterByIndex(2))
	if state2.Values[0].Symbol != "C" {
		t.Errorf("Expected taxon2 char 2 to copy 'C', got %v", state2.Values[0].Symbol)
	}

	// Char 3 (index 3) was a matchchar '.', should copy taxon1's 'G'
	state3 := t2.GetState(charBlock.Matrix.GetCharacterByIndex(3))
	if state3.Values[0].Symbol != "G" {
		t.Errorf("Expected taxon2 char 3 to copy 'G', got %v", state3.Values[0].Symbol)
	}
}

// TestParseEquatesAndPolymorphic ensures that ambiguous codes (like 'R')
// and explicit polymorphic blocks '(AC)' expand accurately to multiple values.
func TestParseEquatesAndPolymorphic(t *testing.T) {
	nexusData := `#NEXUS
	BEGIN CHARACTERS;
		DIMENSIONS NCHAR=3;
		FORMAT DATATYPE=DNA;
		MATRIX
		  taxon1  R(AC)T
		  taxon2  (AG)MT
		;
	END;`

	nex, err := nexus.Parse(strings.NewReader(nexusData))
	if err != nil {
		t.Fatalf("Failed to parse nexus string: %v", err)
	}

	charBlock, ok := core.GetBlock[*characters.CharactersBlock](nex)
	if !ok {
		t.Fatal("Failed to extract CharactersBlock")
	}

	t1 := charBlock.Matrix.GetTaxon("taxon1")
	t2 := charBlock.Matrix.GetTaxon("taxon2")
	if t1 == nil || t2 == nil {
		t.Fatal("taxon1 or taxon2 not found")
	}
	// Char 1: 'R' should equate to A and G (based on DefaultEquates in symbols.go)
	state1 := t1.GetState(charBlock.Matrix.GetCharacterByIndex(1))
	if len(state1.Values) != 2 || state1.Values[0].Symbol != "A" || state1.Values[1].Symbol != "G" {
		t.Errorf("Expected Equate 'R' to expand to [A, G], got %v", state1.Values)
	}

	// Char 2: '(AC)' should be parsed as StatePolymorphic
	state2 := t1.GetState(charBlock.Matrix.GetCharacterByIndex(2))
	if state2.Type != characters.StatePolymorphic {
		t.Errorf("Expected state to be StatePolymorphic, got %v", state2.Type)
	}
	if len(state2.Values) != 2 || state2.Values[0].Symbol != "A" || state2.Values[1].Symbol != "C" {
		t.Errorf("Expected Polymorphic '(AC)' to expand to [A, C], got %v", state2.Values)
	}
}
