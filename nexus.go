package nexus

import (
	"fmt"
	"io"
	"strings"
)

// Nexus represents the root structure of a NEXUS file
type Nexus struct {
	Blocks []Block
}

// New creates a new, empty Nexus file container ready for building.
func New() *Nexus {
	return &Nexus{
		Blocks: make([]Block, 0),
	}
}

// Block defines the interface for any NEXUS block
type Block interface {
	Parse(s *Scanner) error
	Render() string
}

// NexusAware allows a block to hold a reference to its parent container,
// enabling cross-block synchronization (e.g., auto-updating the TAXA block).
type NexusAware interface {
	SetNexus(n *Nexus)
}

// Parse reads a NEXUS format file from an io.Reader and populates the Nexus struct.
func Parse(r io.Reader) (*Nexus, error) {
	scanner := NewScanner(r)
	nex := &Nexus{
		Blocks: make([]Block, 0),
	}

	token, err := scanner.NextToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if strings.ToUpper(token) != "#NEXUS" {
		return nil, fmt.Errorf("invalid file format: expected #NEXUS, got %s", token)
	}

	for {
		token, err = scanner.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.ToUpper(token) == "BEGIN" {
			blockName, err := scanner.NextToken()
			if err != nil {
				return nil, fmt.Errorf("expected block name after BEGIN: %w", err)
			}

			if err := expectSemicolon(scanner); err != nil {
				return nil, err
			}

			factory, exists := blockRegistry[strings.ToUpper(blockName)]
			if exists {
				// Create a new instance of the block
				block := factory()

				// Tell the block to parse its own contents
				if err := block.Parse(scanner); err != nil {
					return nil, fmt.Errorf("error parsing %s block: %w", blockName, err)
				}
				nex.Blocks = append(nex.Blocks, block)
			} else {
				// If not registered, modularity allows us to safely ignore it.
				if err := skipBlock(scanner); err != nil {
					return nil, fmt.Errorf("failed to skip unknown block %s: %w", blockName, err)
				}
			}
		}
	}

	return nex, nil
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

// RegisterTaxon ensures the taxon exists in the file's TAXA block[cite: 336].
// If no TAXA block exists, it creates one.
func (n *Nexus) RegisterTaxon(name string) {
	taxaBlock, ok := GetBlock[*TaxaBlock](n)
	if !ok {
		taxaBlock = n.NewTaxaBlock()
	}
	taxaBlock.AddTaxon(name)
}
