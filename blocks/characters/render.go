package characters

import (
	"strings"
	"text/template"
)

// The template for the CHARACTERS block
const charsTmplStr = `BEGIN CHARACTERS;
{{- if .Title}}
	TITLE {{.Title}};
{{- end}}
	DIMENSIONS NCHAR={{.Dimensions.NChar}};
{{- if or .Format.DataType .Format.Missing .Format.Gap .Format.Symbols}}
	FORMAT{{if .Format.DataType}} DATATYPE={{.Format.DataType}}{{end}}{{if .Format.Missing}} MISSING={{.Format.Missing}}{{end}}{{if .Format.Gap}} GAP={{.Format.Gap}}{{end}}{{if .Format.Symbols}} SYMBOLS="{{.Format.Symbols}}"{{end}};
{{- end}}
{{- if .SortedLabels}}
	CHARSTATELABELS
{{- range .SortedLabels}}
		{{.ID}} {{if .Name}}{{.Name}}{{end}}{{if .States}} / {{.States}}{{end}},
{{- end}}
	;
{{- end}}
	MATRIX
{{- range .Matrix}}
	{{.PaddedName}}{{range .States}}{{.Render}}{{end}}
{{- end}}
	;
END;
`

var charsTmpl = template.Must(template.New("characters").Parse(charsTmplStr))

// Render formats the CharacterState back into its proper NEXUS representation.
func (c CharacterState) Render() string {
	valStr := strings.Join(c.Value, " ")
	switch c.Type {
	case StatePolymorphic:
		return "(" + valStr + ")"
	case StateUncertain:
		return "{" + valStr + "}"
	default:
		return valStr
	}
}
