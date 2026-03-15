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
	nexus      *Nexus
	Title      string
	Dimensions NTax
	TaxLabels  []string
}

// NewTaxaBlock creates, appends, and returns a new TAXA block.
func (n *Nexus) NewTaxaBlock() *TaxaBlock {
	tb := &TaxaBlock{nexus: n}
	n.Blocks = append(n.Blocks, tb)
	return tb
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

// SetNexus implements the NexusAware interface.
func (t *TaxaBlock) SetNexus(n *Nexus) {
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
