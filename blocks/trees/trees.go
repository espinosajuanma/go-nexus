package trees

import (
	"github.com/espinosajuanma/nexus/core"
)

func init() {
	core.RegisterBlock("TREES", func() core.Block {
		return &TreesBlock{
			Translate: make(map[string]string),
		}
	})
}

// New creates, appends, and returns a new TREES block.
func New(n *core.Nexus) *TreesBlock {
	tb := &TreesBlock{
		nexus:     n,
		Translate: make(map[string]string),
	}
	n.Blocks = append(n.Blocks, tb)
	return tb
}

// TreesBlock stores information about trees.
type TreesBlock struct {
	nexus     *core.Nexus
	Translate map[string]string
	Trees     []Tree
}

// SetNexus implements the NexusAware interface.
func (t *TreesBlock) SetNexus(n *core.Nexus) {
	t.nexus = n
}

// AddTranslate maps an arbitrary token (like "1") to a valid taxon name .
// It automatically syncs with the TAXA block.
func (t *TreesBlock) AddTranslate(token string, taxonName string) {
	if t.Translate == nil {
		t.Translate = make(map[string]string)
	}
	t.Translate[token] = taxonName

	// Auto-register the taxon in the TAXA block
	if t.nexus != nil {
		t.nexus.RegisterTaxon(taxonName)
	}
}

// AddTree appends a fully built Tree to the block.
func (t *TreesBlock) AddTree(name string, isDefault bool, root *TreeNode) {
	t.Trees = append(t.Trees, Tree{
		Name:      name,
		IsDefault: isDefault,
		Root:      root,
	})
}
