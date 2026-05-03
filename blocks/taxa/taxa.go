package taxa

import (
	"github.com/espinosajuanma/nexus/core"
)

// init automatically registers the TAXA block with the core parser.
func init() {
	core.RegisterBlock("TAXA", func() core.Block {
		return &TaxaBlock{}
	})
}

// New creates, appends, and returns a new TAXA block.
func New(n *core.Nexus) *TaxaBlock {
	tb := &TaxaBlock{nexus: n}
	n.Blocks = append(n.Blocks, tb)
	return tb
}

// TaxaBlock specifies information about taxa.
type TaxaBlock struct {
	nexus      *core.Nexus
	Title      string
	Dimensions NTax
	TaxLabels  []string
}

type NTax struct {
	Count int
}

// SetNexus implements the NexusAware interface.
func (t *TaxaBlock) SetNexus(n *core.Nexus) {
	t.nexus = n
}

// SetTitle applies a title to the block.
func (t *TaxaBlock) SetTitle(title string) {
	t.Title = title
}

// AddTaxon appends a taxon to the block if it doesn't already exist .
// It automatically updates the NTAX dimension count.
func (t *TaxaBlock) AddTaxon(name string) {
	for _, label := range t.TaxLabels {
		if label == name {
			return // Already exists, no need to add
		}
	}
	t.TaxLabels = append(t.TaxLabels, name)
	t.Dimensions.Count = len(t.TaxLabels)
}
