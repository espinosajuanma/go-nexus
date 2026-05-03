package core

import (
	"fmt"
	"io"

	"github.com/espinosajuanma/nexus/scanner"
)

// New creates a new, empty Nexus file container ready for building.
func New() *Nexus {
	return &Nexus{
		Blocks: make([]Block, 0),
	}
}

// Nexus represents the root structure of a NEXUS file
type Nexus struct {
	Blocks []Block
}

// NexusAware allows a block to hold a reference to its parent container,
// enabling cross-block synchronization (e.g., auto-updating the TAXA block).
type NexusAware interface {
	SetNexus(n *Nexus)
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
