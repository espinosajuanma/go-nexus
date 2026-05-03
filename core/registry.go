package core

import "strings"

// BlockFactory is a function that returns a new, initialized Block.
type BlockFactory func() Block

// BlockRegistry holds the mapping of block names to their factory functions.
var BlockRegistry = make(map[string]BlockFactory)

// RegisterBlock allows any block to register itself with the core parser.
// This allows for extreme extensibility, even for custom/private blocks .
func RegisterBlock(name string, factory BlockFactory) {
	BlockRegistry[strings.ToUpper(name)] = factory
}
