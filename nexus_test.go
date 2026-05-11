package nexus_test

import (
	"strings"
	"testing"

	"github.com/espinosajuanma/nexus"
)

// TestNew verifies that a new, empty Nexus container is created correctly.
func TestNew(t *testing.T) {
	nex := nexus.New()

	if nex == nil {
		t.Fatal("Expected New() to return a non-nil Nexus instance")
	}

	if len(nex.Blocks) != 0 {
		t.Errorf("Expected a newly created Nexus instance to have 0 blocks, got %d", len(nex.Blocks))
	}
}

// TestParse_Valid ensures that the Parse function correctly processes a valid
// io.Reader and that the blank imports in nexus.go successfully registered the blocks.
func TestParse_Valid(t *testing.T) {
	// A minimal valid NEXUS string using a TAXA block
	validNexus := `#NEXUS
	BEGIN TAXA;
		DIMENSIONS NTAX=3;
		TAXLABELS TaxonA TaxonB TaxonC;
	END;`

	reader := strings.NewReader(validNexus)
	nex, err := nexus.Parse(reader)

	if err != nil {
		t.Fatalf("Parse() failed with unexpected error: %v", err)
	}

	if nex == nil {
		t.Fatal("Expected Parse() to return a non-nil Nexus instance")
	}

	// If len > 0, it means the core parser successfully matched "TAXA" to the
	// registered block factory from the blank imports in nexus.go
	if len(nex.Blocks) == 0 {
		t.Error("Expected Parse() to extract blocks, but found 0")
	}
}

// TestParse_Invalid ensures the parser properly returns errors when fed bad data.
func TestParse_Invalid(t *testing.T) {
	// Missing the required #NEXUS header
	invalidNexus := `BEGIN TAXA; DIMENSIONS NTAX=3; END;`

	reader := strings.NewReader(invalidNexus)
	_, err := nexus.Parse(reader)

	if err == nil {
		t.Error("Expected Parse() to return an error for a missing #NEXUS header, but got nil")
	}
}
