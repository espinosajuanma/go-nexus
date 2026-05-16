package characters

import (
	"github.com/espinosajuanma/go-nexus/core"
)

// init automatically registers the CHARACTERS block with the core parser.
func init() {
	core.RegisterBlock("CHARACTERS", func(name string) core.Block {
		cb := &CharactersBlock{
			Format: Format{
				Missing:        "?",
				Gap:            "-",
				DataType:       Standard,
				Labels:         true,
				Symbols:        DefaultSymbols[Standard],
				RespectCase:    false,
				InterleaveSize: 70,
			},
			Matrix: Matrix{},
			Name:   name,
		}

		cb.Matrix.parent = cb

		return cb
	})
}

// CharactersBlock defines characters and includes character data.
type CharactersBlock struct {
	nexus      *core.Core
	Name       string
	Title      string
	Dimensions int
	Format     Format
	Matrix     Matrix
	Eliminate  map[int]bool
}

// New creates, appends, and returns a new CHARACTERS block.
func New(n *core.Core, dt DataType) *CharactersBlock {
	cb := &CharactersBlock{
		nexus: n,
		Name:  "CHARACTERS",
		Format: Format{
			DataType:       dt,
			Missing:        "?",
			Gap:            "-",
			Labels:         true,
			InterleaveSize: 70,
		},
		Matrix: Matrix{},
	}
	cb.Matrix.parent = cb

	n.Blocks = append(n.Blocks, cb)
	return cb
}

// SetCore implements the CoreAware interface.
func (c *CharactersBlock) SetCore(n *core.Core) {
	c.nexus = n
}

// SetTitle applies a title to the block.
func (c *CharactersBlock) SetTitle(title string) {
	c.Title = title
}

// AddTaxon adds a taxon to the character matrix and returns the new Taxon.
func (c *CharactersBlock) AddTaxon(name string) *Taxon {
	return c.Matrix.AddTaxon(name)
}

// GetName returns the name of the block, fulfilling the Block interface.
func (c *CharactersBlock) GetName() string {
	return c.Name
}
