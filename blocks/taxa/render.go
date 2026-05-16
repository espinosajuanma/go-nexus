package taxa

import (
	_ "embed"

	"github.com/espinosajuanma/go-nexus/templater"
)

//go:embed taxa.tmpl
var taxaTmplStr string

const name = "TAXA"

// Render implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Render() (string, error) {
	tmpl, err := templater.New(name, taxaTmplStr)
	if err != nil {
		return "", err
	}
	return tmpl.Render(t)
}
