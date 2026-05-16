package nexus

import (
	"bytes"
	"strings"
	"testing"

	"github.com/espinosajuanma/nexus/blocks/characters"
)

func TestNew(t *testing.T) {
	n := New()
	if n == nil {
		t.Fatal("Expected New() to return a valid Nexus instance, got nil")
	}
	if n.core == nil {
		t.Fatal("Expected New() to initialize the core AST, but it was nil")
	}
}

func TestParse(t *testing.T) {
	// A minimal valid NEXUS file string
	input := "#NEXUS\n"
	r := strings.NewReader(input)

	n, err := Parse(r)
	if err != nil {
		// Note: If the underlying parser requires more structure, this might fail.
		// However, a robust parser should handle an empty #NEXUS declaration.
		t.Fatalf("Parse() failed with error: %v", err)
	}
	if n == nil {
		t.Fatal("Expected Parse() to return a valid Nexus instance, got nil")
	}
}

func TestCharactersBlockAPI(t *testing.T) {
	n := New()

	// Verify it doesn't exist initially
	_, ok := n.GetCharactersBlock()
	if ok {
		t.Error("GetCharactersBlock() should return false when no block exists")
	}

	// Create a new block
	var dummyDataType characters.DataType = "DNA" // Using a mocked data type
	createdBlock := n.NewCharactersBlock(dummyDataType)
	if createdBlock == nil {
		t.Fatal("NewCharactersBlock() returned nil")
	}

	// Retrieve and verify
	fetchedBlock, ok := n.GetCharactersBlock()
	if !ok {
		t.Fatal("GetCharactersBlock() failed to find the newly created block")
	}
	if fetchedBlock != createdBlock {
		t.Errorf("GetCharactersBlock() returned a different instance. Expected %p, got %p", createdBlock, fetchedBlock)
	}
}

func TestTaxaBlockAPI(t *testing.T) {
	n := New()

	// Verify it doesn't exist initially
	_, ok := n.GetTaxaBlock()
	if ok {
		t.Error("GetTaxaBlock() should return false when no block exists")
	}

	// Create a new block
	createdBlock := n.NewTaxaBlock()
	if createdBlock == nil {
		t.Fatal("NewTaxaBlock() returned nil")
	}

	// Retrieve and verify
	fetchedBlock, ok := n.GetTaxaBlock()
	if !ok {
		t.Fatal("GetTaxaBlock() failed to find the newly created block")
	}
	if fetchedBlock != createdBlock {
		t.Errorf("GetTaxaBlock() returned a different instance. Expected %p, got %p", createdBlock, fetchedBlock)
	}
}

func TestTreesBlockAPI(t *testing.T) {
	n := New()

	// Verify it doesn't exist initially
	_, ok := n.GetTreesBlock()
	if ok {
		t.Error("GetTreesBlock() should return false when no block exists")
	}

	// Create a new block
	createdBlock := n.NewTreesBlock()
	if createdBlock == nil {
		t.Fatal("NewTreesBlock() returned nil")
	}

	// Retrieve and verify
	fetchedBlock, ok := n.GetTreesBlock()
	if !ok {
		t.Fatal("GetTreesBlock() failed to find the newly created block")
	}
	if fetchedBlock != createdBlock {
		t.Errorf("GetTreesBlock() returned a different instance. Expected %p, got %p", createdBlock, fetchedBlock)
	}
}

func TestUnknownBlockAPI(t *testing.T) {
	n := New()
	blockName := "CUSTOM_BLOCK"

	// Verify it doesn't exist initially
	_, ok := n.GetBlockByName(blockName)
	if ok {
		t.Errorf("GetBlockByName(%q) should return false when block doesn't exist", blockName)
	}

	// Create a new generic block
	createdBlock := n.NewUnknownBlock(blockName)
	if createdBlock == nil {
		t.Fatal("NewUnknownBlock() returned nil")
	}

	// Retrieve and verify
	fetchedBlock, ok := n.GetBlockByName(blockName)
	if !ok {
		t.Fatalf("GetBlockByName(%q) failed to find the newly created block", blockName)
	}
	if fetchedBlock != createdBlock {
		t.Errorf("GetBlockByName(%q) returned a different instance. Expected %p, got %p", blockName, createdBlock, fetchedBlock)
	}
}

func TestExport(t *testing.T) {
	n := New()

	// Add some blocks to ensure there is something to export
	n.NewTaxaBlock()
	n.NewTreesBlock()

	var buf bytes.Buffer
	err := n.Export(&buf)
	if err != nil {
		t.Fatalf("Export() failed with error: %v", err)
	}

	// Just checking if anything was written.
	// The exact output depends on the underlying core.Export implementation.
	if buf.Len() == 0 {
		t.Error("Export() wrote 0 bytes to the buffer")
	}
}
