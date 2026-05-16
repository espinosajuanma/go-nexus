package nexus

import (
	"io"

	"github.com/espinosajuanma/go-nexus/blocks/characters"
	"github.com/espinosajuanma/go-nexus/blocks/generic"
	"github.com/espinosajuanma/go-nexus/blocks/taxa"
	"github.com/espinosajuanma/go-nexus/blocks/trees"
	"github.com/espinosajuanma/go-nexus/core"
	"github.com/espinosajuanma/go-nexus/parser"
)

// Nexus is the smart wrapper around the core AST.
type Nexus struct {
	core *core.Core
}

// New creates a new, empty Nexus file with the smart API.
func New() *Nexus {
	return &Nexus{
		core: core.New(),
	}
}

// Parse reads a standard io.Reader and returns the smart Nexus wrapper.
func Parse(r io.Reader) (*Nexus, error) {
	ast, err := parser.Parse(r)
	if err != nil {
		return nil, err
	}
	return &Nexus{core: ast}, nil
}

// NewCharactersBlock appends and returns a new CHARACTERS block.
func (n *Nexus) NewCharactersBlock(dt characters.DataType) *characters.CharactersBlock {
	return characters.New(n.core, dt)
}

// GetCharactersBlock fetches the CHARACTERS block if it exists.
func (n *Nexus) GetCharactersBlock() (*characters.CharactersBlock, bool) {
	return core.GetBlock[*characters.CharactersBlock](n.core)
}

// NewTaxaBlock appends and returns a new TAXA block.
func (n *Nexus) NewTaxaBlock() *taxa.TaxaBlock {
	return taxa.New(n.core)
}

// GetTaxaBlock fetches the TAXA block if it exists.
func (n *Nexus) GetTaxaBlock() (*taxa.TaxaBlock, bool) {
	return core.GetBlock[*taxa.TaxaBlock](n.core)
}

// NewTreesBlock appends and returns a new TREES block.
func (n *Nexus) NewTreesBlock() *trees.TreesBlock {
	return trees.New(n.core)
}

// GetTreesBlock fetches the TREES block if it exists.
func (n *Nexus) GetTreesBlock() (*trees.TreesBlock, bool) {
	return core.GetBlock[*trees.TreesBlock](n.core)
}

// NewUnknownBlock creates and appends a new block with the given name, returning it as a generic block.
func (n *Nexus) NewUnknownBlock(name string) core.Block {
	return generic.New(n.core, name)
}

// GetBlockByName allows fetching any block by its name.
func (n *Nexus) GetBlockByName(name string) (core.Block, bool) {
	for _, block := range n.core.Blocks {
		if block.GetName() == name {
			return block, true
		}
	}
	return nil, false
}

// Export writes the Nexus data to a standard io.Writer in NEXUS format.
func (n *Nexus) Export(w io.Writer) error {
	return n.core.Export(w)
}
