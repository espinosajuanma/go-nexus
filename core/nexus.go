package core

import (
	"fmt"
	"io"

	"github.com/espinosajuanma/nexus/scanner"
)

// Nexus represents the root structure of a NEXUS file
type Nexus struct {
	Blocks []Block
}

// NexusAware allows a block to hold a reference to its parent container,
// enabling cross-block synchronization (e.g., auto-updating the TAXA block).
type NexusAware interface {
	SetNexus(n *Nexus)
}

// New creates a new, empty Nexus file container ready for building.
func New() *Nexus {
	return &Nexus{
		Blocks: make([]Block, 0),
	}
}

// Block defines the interface for any NEXUS block
type Block interface {
	Parse(s *scanner.Scanner) error
	Render() string
}

// Export serializes the Nexus structure to the provided io.Writer.
func (n *Nexus) Export(w io.Writer) error {
	fmt.Fprintf(w, "#NEXUS\n\n")
	for _, block := range n.Blocks {
		fmt.Fprintf(w, "%s\n", block.Render())
	}
	return nil
}

// GetBlock is a generic helper that searches the parsed blocks
// and returns the first block matching the requested type.
// It returns the typed block and a boolean indicating if it was found.
func GetBlock[T Block](n *Nexus) (T, bool) {
	for _, b := range n.Blocks {
		if typedBlock, ok := b.(T); ok {
			return typedBlock, true
		}
	}
	var zero T
	return zero, false
}

type TaxaRegistry interface {
	AddTaxon(name string)
}

// RegisterTaxon ensures the taxon exists. It decouples the core from the taxa block.
func (n *Nexus) RegisterTaxon(name string) {
	sanitizedName := SanitizeName(name)

	// Look for an existing block that implements the TaxaRegistry interface
	for _, b := range n.Blocks {
		if registry, ok := b.(TaxaRegistry); ok {
			registry.AddTaxon(sanitizedName)
			return
		}
	}

	// If no block implements TaxaRegistry, we can choose to create a new TAXA block
	factory, exists := BlockRegistry["TAXA"]
	if !exists {
		return
	}

	// Create the block, link it to Nexus, and append it
	newBlock := factory()
	if aware, ok := newBlock.(NexusAware); ok {
		aware.SetNexus(n)
	}
	n.Blocks = append(n.Blocks, newBlock)

	// Cast the newly created block to our interface and add the taxon
	if registry, ok := newBlock.(TaxaRegistry); ok {
		registry.AddTaxon(sanitizedName)
	}
}
