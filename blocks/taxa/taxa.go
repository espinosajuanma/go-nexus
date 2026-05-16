package taxa

import (
	"fmt"
	"slices"

	"github.com/espinosajuanma/go-nexus/core"
	"github.com/espinosajuanma/go-nexus/parser"
)

// init automatically registers the TAXA block with the core parser.
func init() {
	core.RegisterBlock("TAXA", func(name string) core.Block {
		return &TaxaBlock{
			Name: name,
		}
	})
}

// New creates, appends, and returns a new TAXA block.
func New(n *core.Core) *TaxaBlock {
	tb := &TaxaBlock{
		nexus: n,
		Name:  "TAXA",
	}
	n.Blocks = append(n.Blocks, tb)
	return tb
}

// TaxaBlock specifies information about taxa.
type TaxaBlock struct {
	nexus         *core.Core
	Name          string
	Title         string
	Dimensions    int
	TaxLabels     []string
	TaxPartitions map[string]TaxPartition
	TaxSets       map[string]TaxSet
}

// SetCore implements the CoreAware interface.
func (t *TaxaBlock) SetCore(n *core.Core) {
	t.nexus = n
}

// SetTitle applies a title to the block.
func (t *TaxaBlock) SetTitle(title string) {
	t.Title = title
}

// GetName returns the name of the block, fulfilling the Block interface.
func (t *TaxaBlock) GetName() string {
	return t.Name
}

// ContainsTaxon checks if a taxon exists in the block
func (t *TaxaBlock) ContainsTaxon(name string) bool {
	normalized := normalizeTaxonName(name)
	return slices.ContainsFunc(t.TaxLabels, func(label string) bool {
		return normalizeTaxonName(label) == normalized
	})
}

// AddTaxon appends a taxon to the block if it doesn't already exist.
func (t *TaxaBlock) AddTaxon(name string) error {
	normalized := normalizeTaxonName(name)

	// NEXUS rule: Cannot consist entirely of digits
	if parser.IsAllDigits(normalized) {
		return fmt.Errorf("invalid taxon name '%s': cannot consist entirely of digits", name)
	}

	// NEXUS rule: Homonym check
	if t.ContainsTaxon(name) {
		return fmt.Errorf("taxon '%s' already exists", name)
	}

	t.TaxLabels = append(t.TaxLabels, name)
	t.Dimensions = len(t.TaxLabels)
	return nil
}

// RemoveTaxon removes a taxon from the block if it exists.
func (t *TaxaBlock) RemoveTaxon(name string) error {
	if !t.ContainsTaxon(name) {
		return fmt.Errorf("taxon '%s' not found", name)
	}

	normalized := normalizeTaxonName(name)
	t.TaxLabels = slices.DeleteFunc(t.TaxLabels, func(label string) bool {
		return normalizeTaxonName(label) == normalized
	})

	t.Dimensions = len(t.TaxLabels)
	return nil
}

// AddTaxSet adds a new TAXSET to the block.
func (t *TaxaBlock) AddTaxSet(name string, format SetFormat, taxaList []string) error {
	for _, taxon := range taxaList {
		if !t.ContainsTaxon(taxon) {
			return fmt.Errorf("cannot add taxon '%s' to set '%s': taxon does not exist in TAXA block", taxon, name)
		}
	}

	if t.TaxSets == nil {
		t.TaxSets = make(map[string]TaxSet)
	}

	t.TaxSets[name] = TaxSet{
		Format:   format,
		TaxaList: taxaList,
	}
	return nil
}

// RemoveTaxSet removes a TAXSET from the block if it exists.
func (t *TaxaBlock) RemoveTaxSet(name string) error {
	if _, exists := t.TaxSets[name]; !exists {
		return fmt.Errorf("tax set '%s' not found", name)
	}
	delete(t.TaxSets, name)
	return nil
}

// AddTaxPartition adds a new TAXPARTITION to the block.
func (t *TaxaBlock) AddTaxPartition(name string, format SetFormat, subsets map[string][]string) error {
	if t.TaxPartitions == nil {
		t.TaxPartitions = make(map[string]TaxPartition)
	}

	t.TaxPartitions[name] = TaxPartition{
		Format:  format,
		Subsets: subsets,
	}
	return nil
}

// RemoveTaxPartition removes a TAXPARTITION from the block if it exists.
func (t *TaxaBlock) RemoveTaxPartition(name string) error {
	if _, exists := t.TaxPartitions[name]; !exists {
		return fmt.Errorf("tax partition '%s' not found", name)
	}
	delete(t.TaxPartitions, name)
	return nil
}
