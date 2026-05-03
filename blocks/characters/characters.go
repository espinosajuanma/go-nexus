package characters

import (
	"github.com/espinosajuanma/nexus/core"
)

// init automatically registers the CHARACTERS block with the core parser.
func init() {
	core.RegisterBlock("CHARACTERS", func() core.Block {
		return &CharactersBlock{
			Format: Format{
				Missing: "?",
				Gap:     "-",
			},
			CharStateLabels: make(map[int]string),
		}
	})
}

// New creates, appends, and returns a new CHARACTERS block.
func New(n *core.Nexus, dt DataType) *CharactersBlock {
	cb := &CharactersBlock{
		nexus: n,
		Format: Format{
			DataType: dt,
			Missing:  "?",
			Gap:      "-",
		},
		CharStateLabels: make(map[int]string),
	}
	n.Blocks = append(n.Blocks, cb)
	return cb
}

// CharactersBlock defines characters and includes character data.
type CharactersBlock struct {
	nexus           *core.Nexus
	Title           string
	Dimensions      Dimensions
	Format          Format
	CharStateLabels map[int]string // to be deprecated
	Matrix          []MatrixRow    // to be deprecated

	Characters []*Character
	Taxa       []*TaxonReference

	data [][]CharacterState
}

// SetNexus implements the NexusAware interface.
func (c *CharactersBlock) SetNexus(n *core.Nexus) {
	c.nexus = n
}

// SetTitle applies a title to the block.
func (c *CharactersBlock) SetTitle(title string) {
	c.Title = title
}
