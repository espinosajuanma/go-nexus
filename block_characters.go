package nexus

import (
	"bytes"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

// The template for the CHARACTERS block [cite: 348-377]
const charsTmplStr = `BEGIN CHARACTERS;
	DIMENSIONS NCHAR={{.Dimensions.NChar}};
{{- if or .Format.DataType .Format.Missing .Format.Gap .Format.Symbols}}
	FORMAT{{if .Format.DataType}} DATATYPE={{.Format.DataType}}{{end}}{{if .Format.Missing}} MISSING={{.Format.Missing}}{{end}}{{if .Format.Gap}} GAP={{.Format.Gap}}{{end}}{{if .Format.Symbols}} SYMBOLS="{{.Format.Symbols}}"{{end}};
{{- end}}
{{- if .SortedLabels}}
	CHARSTATELABELS
{{- range .SortedLabels}}
		{{.ID}} {{.Name}},
{{- end}}
	;
{{- end}}
	MATRIX
{{- range .Matrix}}
	{{.TaxonName}}	{{.Data}}
{{- end}}
	;
END;
`

var charsTmpl = template.Must(template.New("characters").Parse(charsTmplStr))

// init automatically registers the CHARACTERS block with the core parser.
func init() {
	RegisterBlock("CHARACTERS", func() Block {
		return &CharactersBlock{
			CharStateLabels: make(map[int]string), // Safely initialize the map
		}
	})
}

// CharactersBlock defines characters and includes character data.
type CharactersBlock struct {
	Dimensions      Dimensions
	Format          Format
	CharStateLabels map[int]string
	Matrix          []MatrixRow
}

type Dimensions struct {
	NChar int
}

// Format specifies the format of the data MATRIX.
type Format struct {
	DataType string // e.g., "STANDARD", "DNA", "CONTINUOUS"
	Missing  string // Default is "?"
	Gap      string // e.g., "-"
	Symbols  string // e.g., "0 1 2"
}

// MatrixRow holds the sequence/data for a single taxon.
type MatrixRow struct {
	TaxonName string
	Data      string
}

// Parse implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Parse(s *Scanner) error {
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
		case "DIMENSIONS":
			tokens, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			for i, tok := range tokens {
				if strings.ToUpper(tok) == "NCHAR" {
					valIdx := i + 1
					if valIdx < len(tokens) && tokens[valIdx] == "=" {
						valIdx++
					}
					if valIdx < len(tokens) {
						count, _ := strconv.Atoi(tokens[valIdx])
						c.Dimensions.NChar = count
					}
				}
			}
		case "FORMAT":
			tokens, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			c.Format = parseFormatCommand(tokens)
		case "CHARSTATELABELS":
			if err := parseCharStateLabels(s, c.CharStateLabels); err != nil {
				return err
			}
		case "MATRIX":
			if err := parseMatrix(s, c); err != nil {
				return err
			}
		default:
			if _, err := readUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

// Render implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Render() string {
	// Sort the map keys to ensure deterministic output
	type labelPair struct {
		ID   int
		Name string
	}
	var sortedLabels []labelPair
	for id, name := range c.CharStateLabels {
		sortedLabels = append(sortedLabels, labelPair{ID: id, Name: name})
	}
	sort.Slice(sortedLabels, func(i, j int) bool {
		return sortedLabels[i].ID < sortedLabels[j].ID
	})

	// 2. Create an anonymous struct to feed the template
	templateData := struct {
		Dimensions   Dimensions
		Format       Format
		SortedLabels []labelPair
		Matrix       []MatrixRow
	}{
		Dimensions:   c.Dimensions,
		Format:       c.Format,
		SortedLabels: sortedLabels,
		Matrix:       c.Matrix,
	}

	// 3. Render the template
	var buf bytes.Buffer
	if err := charsTmpl.Execute(&buf, templateData); err != nil {
		return "[ERROR rendering CHARACTERS block: " + err.Error() + "]\n"
	}
	return buf.String()
}
