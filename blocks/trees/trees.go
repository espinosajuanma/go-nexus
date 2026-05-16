package trees

import (
	"github.com/espinosajuanma/nexus/core"
)

func init() {
	core.RegisterBlock("TREES", func(name string) core.Block {
		return &TreesBlock{
			Translate: make(map[string]string),
			Name:      name,
		}
	})
}

// New creates, appends, and returns a new TREES block.
func New(n *core.Core) *TreesBlock {
	tb := &TreesBlock{
		nexus:     n,
		Translate: make(map[string]string),
		Name:      "TREES",
	}
	n.Blocks = append(n.Blocks, tb)
	return tb
}

// TreesBlock stores information about trees.
type TreesBlock struct {
	nexus     *core.Core
	Name      string
	Translate map[string]string
	Trees     []Tree
}

// SetCore implements the CoreAware interface.
func (t *TreesBlock) SetCore(n *core.Core) {
	t.nexus = n
}

// GetName returns the name of the block, fulfilling the Block interface.
func (t *TreesBlock) GetName() string {
	return t.Name
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
