package generic

import (
	"github.com/espinosajuanma/nexus/core"
)

// init automatically registers the GENERIC block with the core parser.
func init() {
	core.RegisterBlock("GENERIC", func(name string) core.Block {
		return &GenericBlock{
			Name: name,
		}
	})
}

// New creates, appends, and returns a new generic block.
func New(n *core.Nexus) *GenericBlock {
	gb := &GenericBlock{nexus: n}
	n.Blocks = append(n.Blocks, gb)
	return gb
}

// GenericBlock is a flexible block type that can hold any content. It is
// designed to be used for blocks that don't fit into the standard categories or
// for custom user-defined blocks.
type GenericBlock struct {
	nexus   *core.Nexus
	Name    string
	Content string
}

// SetNexus implements the NexusAware interface.
func (t *GenericBlock) SetNexus(n *core.Nexus) {
	t.nexus = n
}
