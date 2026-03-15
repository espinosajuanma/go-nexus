package nexus

import "strings"

// BlockFactory is a function that returns a new, initialized Block.
type BlockFactory func() Block

// blockRegistry holds the mapping of block names to their factory functions.
var blockRegistry = make(map[string]BlockFactory)

// RegisterBlock allows any block to register itself with the core parser.
// This allows for extreme extensibility, even for custom/private blocks .
func RegisterBlock(name string, factory BlockFactory) {
	blockRegistry[strings.ToUpper(name)] = factory
}
