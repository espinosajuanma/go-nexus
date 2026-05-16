package core

import (
	"fmt"
	"io"

	"github.com/espinosajuanma/go-nexus/scanner"
)

// New creates a new, empty Nexus file container ready for building.
func New() *Core {
	return &Core{
		Blocks: make([]Block, 0),
	}
}

// Core represents the root structure of a NEXUS file
type Core struct {
	Blocks []Block
}

// NexusAware allows a block to hold a reference to its parent container,
// enabling cross-block synchronization (e.g., auto-updating the TAXA block).
type NexusAware interface {
	SetNexus(n *Core)
}

// Block defines the interface for any NEXUS block
type Block interface {
	Parse(s *scanner.Scanner) error
	Render() (string, error)
	GetName() string
}

// Export serializes the Nexus structure to the provided io.Writer.
func (n *Core) Export(w io.Writer) error {
	fmt.Fprintf(w, "#NEXUS\n")
	for _, block := range n.Blocks {
		rendered, err := block.Render()
		if err != nil {
			return err
		}
		fmt.Fprintf(w, "\n%s\n", rendered)
	}
	return nil
}
