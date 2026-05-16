package core_test

import (
	"bytes"
	"reflect"
	"testing"

	_ "github.com/espinosajuanma/nexus/blocks/characters"
	_ "github.com/espinosajuanma/nexus/blocks/generic"
	_ "github.com/espinosajuanma/nexus/blocks/taxa"
	_ "github.com/espinosajuanma/nexus/blocks/trees"
	. "github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/scanner"
	"github.com/espinosajuanma/nexus/utils"
)

// --- MOCKS ---

// MockBlock implements the Block and TaxaRegistry interfaces for testing.
type MockBlock struct {
	RenderOutput string
	RenderErr    error
	AddedTaxa    []string
}

func (m *MockBlock) Parse(s *scanner.Scanner) error { return nil }
func (m *MockBlock) Render() (string, error)        { return m.RenderOutput, m.RenderErr }
func (m *MockBlock) AddTaxon(name string)           { m.AddedTaxa = append(m.AddedTaxa, name) }
func (m *MockBlock) GetName() string                { return "" }

// Another mock block to test the generic GetBlock function
type AnotherMockBlock struct{}

func (a *AnotherMockBlock) Parse(s *scanner.Scanner) error { return nil }
func (a *AnotherMockBlock) Render() (string, error)        { return "", nil }
func (a *AnotherMockBlock) GetName() string                { return "" }

// A third mock to test missing block lookups
type MissingMockBlock struct{}

func (m *MissingMockBlock) Parse(s *scanner.Scanner) error { return nil }
func (m *MissingMockBlock) Render() (string, error)        { return "", nil }
func (m *MissingMockBlock) GetName() string                { return "" }

// --- TESTS ---

func TestUtils_DecodeName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Short string", "A", "A"},
		{"Standard unquoted", "Homo_sapiens", "Homo sapiens"},
		{"Single quoted", "'Homo sapiens'", "Homo sapiens"},
		{"Single quoted with escaped quote", "'John''s tree'", "John's tree"},
		{"Double quoted", "\"Homo sapiens\"", "Homo sapiens"},
		{"Trims space", "  spaced_name  ", "spaced name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := utils.DecodeName(tt.input); got != tt.expected {
				t.Errorf("DecodeName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUtils_EncodeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Homo sapiens", "Homo_sapiens"},
		{"  trim me  ", "trim_me"},
		{"Already_Encoded", "Already_Encoded"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := utils.EncodeName(tt.input); got != tt.expected {
				t.Errorf("EncodeName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUtils_QuoteName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"SafeName", "SafeName"},
		{"Has Space", "'Has Space'"},
		{"Has-Dash", "'Has-Dash'"},
		{"Has,Comma", "'Has,Comma'"},
		{"John's", "'John''s'"},
		{"  Trimmed  ", "Trimmed"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := utils.QuoteName(tt.input); got != tt.expected {
				t.Errorf("QuoteName(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNexus_New(t *testing.T) {
	n := New()
	if n == nil {
		t.Fatal("New() returned nil")
	}
	if n.Blocks == nil {
		t.Error("New() should initialize the Blocks slice")
	}
}

func TestNexus_Export(t *testing.T) {
	n := New()
	n.Blocks = append(n.Blocks, &MockBlock{RenderOutput: "BEGIN MOCK;\nEND;"})

	var buf bytes.Buffer
	err := n.Export(&buf)

	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	expected := "#NEXUS\n\nBEGIN MOCK;\nEND;\n"
	if buf.String() != expected {
		t.Errorf("Export() output mismatch.\nGot:\n%q\nWant:\n%q", buf.String(), expected)
	}
}

func TestNexus_GetBlock(t *testing.T) {
	n := New()
	mock1 := &MockBlock{}
	mock2 := &AnotherMockBlock{}
	n.Blocks = append(n.Blocks, mock1, mock2)

	// Test getting the first block type
	gotMock1, ok := GetBlock[*MockBlock](n)
	if !ok || gotMock1 != mock1 {
		t.Error("GetBlock failed to retrieve *MockBlock")
	}

	// Test getting the second block type
	gotMock2, ok := GetBlock[*AnotherMockBlock](n)
	if !ok || gotMock2 != mock2 {
		t.Error("GetBlock failed to retrieve *AnotherMockBlock")
	}

	// Test missing block using a concrete type that isn't in the slice
	_, ok = GetBlock[*MissingMockBlock](n)
	if ok {
		t.Error("GetBlock should return false for a block type that doesn't exist")
	}
}

func TestRegistry_RegisterBlock(t *testing.T) {
	// Register a dummy factory
	factory := func(name string) Block { return &MockBlock{} }
	RegisterBlock("TESTBLOCK", factory)

	storedFactory, exists := BlockRegistry["TESTBLOCK"]
	if !exists {
		t.Fatal("RegisterBlock failed to add to BlockRegistry")
	}

	// Verify the factory produces the expected type
	block := storedFactory("TESTBLOCK")
	if _, ok := block.(*MockBlock); !ok {
		t.Errorf("Block factory produced wrong type: %v", reflect.TypeOf(block))
	}
}

func TestRegistry_RegisterTaxon(t *testing.T) {
	n := New()
	mock := &MockBlock{}
	n.Blocks = append(n.Blocks, mock)

	// RegisterTaxon should find the MockBlock (which implements TaxaRegistry) and call AddTaxon
	n.RegisterTaxon("Homo_sapiens")
	n.RegisterTaxon("'Felis catus'")

	if len(mock.AddedTaxa) != 2 {
		t.Fatalf("Expected 2 taxa to be added, got %d", len(mock.AddedTaxa))
	}

	// RegisterTaxon also runs DecodeName, so underscores and quotes should be resolved
	if mock.AddedTaxa[0] != "Homo sapiens" {
		t.Errorf("Taxon 1 not decoded correctly: %s", mock.AddedTaxa[0])
	}
	if mock.AddedTaxa[1] != "Felis catus" {
		t.Errorf("Taxon 2 not decoded correctly: %s", mock.AddedTaxa[1])
	}
}
