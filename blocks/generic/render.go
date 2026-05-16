package generic

import (
	_ "embed"

	"github.com/espinosajuanma/go-nexus/templater"
)

//go:embed generic.tmpl
var genericTmplStr string

const name = "GENERIC"

// Render implements the Block interface for GenericBlock.
func (t *GenericBlock) Render() (string, error) {
	tmpl, err := templater.New(name, genericTmplStr)
	if err != nil {
		return "", err
	}
	return tmpl.Render(t)
}
