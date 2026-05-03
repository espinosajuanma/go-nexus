package taxa

import (
	"bytes"
	"text/template"

	"github.com/espinosajuanma/nexus/core"
)

// The template for the TAXA block
const taxaTmplStr = `BEGIN TAXA;
{{- if .Title}}
	TITLE {{.Title}};
{{- end}}
	DIMENSIONS NTAX={{.Dimensions.Count}};
	TAXLABELS
{{- range .TaxLabels}}
		{{.}}
{{- end}}
	;
END;
`

var taxaTmpl = template.Must(template.New("taxa").Parse(taxaTmplStr))

// Render implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Render() string {
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

	var buf bytes.Buffer
	if err := taxaTmpl.Execute(&buf, templateData); err != nil {
		// If template execution fails, return a NEXUS comment with the error
		return "[ERROR rendering TAXA block: " + err.Error() + "]\n"
	}
	return buf.String()
}
