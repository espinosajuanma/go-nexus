package xread_test

import (
	"strings"
	"testing"

	"github.com/espinosajuanma/nexus"
	"github.com/espinosajuanma/nexus/blocks/characters"
	"github.com/espinosajuanma/nexus/exports/xread"
)

// setupTestNexus programmatically creates a realistic Nexus AST for testing.
func setupTestNexus() *nexus.Nexus {
	nex := nexus.New()
	cb := nex.NewCharactersBlock(characters.Standard)
	cb.Title = "Morphology"

	// Setup Characters
	// Character 0: Has a name and state labels
	c1 := cb.Matrix.AddCharacter("Eye_Color", "Blue", "Brown")
	// Character 1: Unnamed BUT HAS state labels (This triggers the fallback logic)
	c2 := cb.Matrix.AddCharacter("", "Present", "Absent")

	// Setup Taxa
	t1 := cb.Matrix.AddTaxon("Species_A")
	t2 := cb.Matrix.AddTaxon("Species B")

	// Populate Species_A (States: 0, ?)
	t1.SetState(c1, characters.StateSingle, "Blue")
	t1.SetState(c2, characters.StateMissing)

	// Populate Species_B (States: [01], 1)
	t2.SetState(c1, characters.StatePolymorphic, "Blue", "Brown")
	t2.SetState(c2, characters.StateSingle, "Absent") // Resolves to '1'

	return nex
}

func TestNonaExport(t *testing.T) {
	nex := setupTestNexus()

	exporter := xread.New(nex, xread.NONA).
		SetProject("Test Project").
		SetAuthor("Test Author")

	out, err := exporter.Render()
	if err != nil {
		t.Fatalf("Failed to render NONA: %v", err)
	}

	// Check Metadata Header
	if !strings.Contains(out, "Project: Test Project") {
		t.Errorf("Expected project metadata in output")
	}
	if !strings.Contains(out, "Author: Test Author") {
		t.Errorf("Expected author metadata in output")
	}

	// Check basic xread structure
	if !strings.Contains(out, "xread") {
		t.Errorf("Expected 'xread' command")
	}
	if !strings.Contains(out, "'Morphology'") {
		t.Errorf("Expected title 'Morphology'")
	}
	if !strings.Contains(out, "2 2") {
		t.Errorf("Expected dimensions '2 2', got:\n%s", out)
	}

	// Check matrix encoding and polymorphisms
	if !strings.Contains(out, "Species_A 0?") {
		t.Errorf("Expected 'Species_A 0?' in matrix, got:\n%s", out)
	}
	if !strings.Contains(out, "Species_B [01]1") {
		t.Errorf("Expected 'Species_B [01]1' in matrix, got:\n%s", out)
	}

	// Check cnames generation
	if !strings.Contains(out, "cnames") {
		t.Errorf("Expected 'cnames' block")
	}
	if !strings.Contains(out, "{ 0 Eye_Color Blue Brown ;") {
		t.Errorf("Expected specific character labels in cnames, got:\n%s", out)
	}

	// Ensure the unnamed character with state labels generated a fallback name
	if !strings.Contains(out, "{ 1 Char_1 Present Absent ;") {
		t.Errorf("Expected fallback name and state labels for empty character, got:\n%s", out)
	}
}

func TestTntExport(t *testing.T) {
	nex := setupTestNexus()

	// Configure TNT specifically
	exporter := xread.New(nex, xread.TNT).
		SetTaxname(true).
		AddCommand("nstates 32;").
		AddCommand("rseed 0;")

	out, err := exporter.Render()
	if err != nil {
		t.Fatalf("Failed to render TNT: %v", err)
	}

	// Check TNT specific commands
	if !strings.Contains(out, "taxname=;") {
		t.Errorf("Expected 'taxname=;' command")
	}
	if !strings.Contains(out, "nstates 32;") {
		t.Errorf("Expected custom command 'nstates 32;'")
	}
	if !strings.Contains(out, "rseed 0;") {
		t.Errorf("Expected custom command 'rseed 0;'")
	}

	// Ensure commands appear before xread
	taxnameIndex := strings.Index(out, "taxname=;")
	xreadIndex := strings.Index(out, "xread")
	if taxnameIndex > xreadIndex {
		t.Errorf("Expected 'taxname=;' to appear before 'xread' block")
	}

	// Ensure basic matrix translates properly
	if !strings.Contains(out, "2 2") {
		t.Errorf("Expected dimensions '2 2'")
	}
	if !strings.Contains(out, "cnames") {
		t.Errorf("Expected 'cnames' block")
	}
}

func TestNoLabels(t *testing.T) {
	// Create an empty matrix with no character names or states
	nex := nexus.New()
	cb := nex.NewCharactersBlock(characters.Standard)
	cb.Matrix.AddCharacter("") // Empty name
	t1 := cb.Matrix.AddTaxon("TaxonA")
	t1.SetState(cb.Matrix.Characters[0], characters.StateMissing)

	exporter := xread.New(nex, xread.NONA)
	out, err := exporter.Render()
	if err != nil {
		t.Fatalf("Failed to render: %v", err)
	}

	// Ensure the `cnames` block completely hides itself if no labels exist
	if strings.Contains(out, "cnames") {
		t.Errorf("Did not expect 'cnames' block when no labels are provided")
	}
}
