package nexus

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"
)

// init automatically registers the TAXA block with the core parser.
func init() {
	RegisterBlock("TAXA", func() Block {
		return &TaxaBlock{}
	})
}

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

// TaxaBlock specifies information about taxa.
type TaxaBlock struct {
	Title      string
	Dimensions NTax
	TaxLabels  []string
}

type NTax struct {
	Count int
}

// Parse implements the Block interface for TaxaBlock.
func (t *TaxaBlock) Parse(s *Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}

		cmd := strings.ToUpper(token)
		// Blocks end with an END or ENDBLOCK command
		if cmd == "END" || cmd == "ENDBLOCK" {
			return expectSemicolon(s)
		}

		switch cmd {
		case "TITLE":
			tokens, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			if len(tokens) > 0 {
				t.Title = strings.Join(tokens, " ")
			}
		case "DIMENSIONS":
			tokens, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			for i, tok := range tokens {
				if strings.ToUpper(tok) == "NTAX" {
					valIdx := i + 1
					if valIdx < len(tokens) && tokens[valIdx] == "=" {
						valIdx++
					}
					if valIdx < len(tokens) {
						count, _ := strconv.Atoi(tokens[valIdx])
						t.Dimensions.Count = count
					}
				}
			}
		case "TAXLABELS":
			labels, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			t.TaxLabels = labels
		default:
			if _, err := readUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

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
