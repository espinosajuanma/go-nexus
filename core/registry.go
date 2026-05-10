package core

import (
	"strings"
)

// TaxaRegistry defines the interface for blocks that can register taxa.
type TaxaRegistry interface {
	AddTaxon(name string)
}

// BlockFactory is a function that returns a new, initialized Block.
type BlockFactory func(name string) Block

// BlockRegistry holds the mapping of block names to their factory functions.
var BlockRegistry = make(map[string]BlockFactory)

// RegisterBlock allows any block to register itself with the core parser.
// This allows for extreme extensibility, even for custom/private blocks .
func RegisterBlock(name string, factory BlockFactory) {
	BlockRegistry[strings.ToUpper(name)] = factory
}

// RegisterTaxon ensures the taxon exists. It decouples the core from the taxa block.
func (n *Nexus) RegisterTaxon(name string) {
	sanitizedName := DecodeName(name)

	// Look for an existing block that implements the TaxaRegistry interface
	for _, b := range n.Blocks {
		if registry, ok := b.(TaxaRegistry); ok {
			registry.AddTaxon(sanitizedName)
			return
		}
	}
}
