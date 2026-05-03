package taxa

import (
	_ "embed"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/templater"
)

//go:embed taxa.tmpl
var taxaTmplStr string

// Render implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Render() (string, error) {
	encodedLabels := make([]string, len(t.TaxLabels))
	for i, label := range t.TaxLabels {
		encodedLabels[i] = core.EncodeName(label)
	}

	templateData := struct {
		Title      string
		Dimensions NTax
		TaxLabels  []string
	}{
		Title:      core.QuoteName(t.Title),
		Dimensions: t.Dimensions,
		TaxLabels:  encodedLabels,
	}

	tmpl, err := templater.New("taxa", taxaTmplStr)
	if err != nil {
		return "", err
	}
	return tmpl.Render(templateData)
}
