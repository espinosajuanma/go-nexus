package taxa

import (
	"bytes"
	"text/template"
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
	var buf bytes.Buffer
	// Execute the template, passing the TaxaBlock struct (.) as the data payload
	if err := taxaTmpl.Execute(&buf, t); err != nil {
		// If template execution fails, return a NEXUS comment with the error
		return "[ERROR rendering TAXA block: " + err.Error() + "]\n"
	}
	return buf.String()
}
