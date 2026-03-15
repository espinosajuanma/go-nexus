package nexus

import (
	"bytes"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

// DataType acts as an enum for standard NEXUS format data types.
type DataType string

const (
	Standard   DataType = "STANDARD"
	DNA        DataType = "DNA"
	RNA        DataType = "RNA"
	Nucleotide DataType = "NUCLEOTIDE"
	Protein    DataType = "PROTEIN"
	Continuous DataType = "CONTINUOUS"
)

// The template for the CHARACTERS block [cite: 348-377]
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
		{{.ID}} {{.Name}},
{{- end}}
	;
{{- end}}
	MATRIX
{{- range .Matrix}}
	{{.TaxonName}}	{{range .States}}{{.Value}}{{end}}
{{- end}}
	;
END;
`

var charsTmpl = template.Must(template.New("characters").Parse(charsTmplStr))

// init automatically registers the CHARACTERS block with the core parser.
func init() {
	RegisterBlock("CHARACTERS", func() Block {
		return &CharactersBlock{
			Format: Format{
				Missing: "?", // Default missing symbol
				Gap:     "-", // User-requested default gap symbol
			},
			CharStateLabels: make(map[int]string), // Safely initialize the map
		}
	})
}

// CharactersBlock defines characters and includes character data.
type CharactersBlock struct {
	Title           string
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
	DataType DataType
	Missing  string
	Gap      string
	Symbols  string
}

// CharacterState holds individualized information about a single character state.
type CharacterState struct {
	Index int    // 1-based character number
	Value string // The actual symbol/data (e.g., "A", "C", "-")
	Label string // The associated CharStateLabel, if any
}

// MatrixRow holds the sequence/data for a single taxon.
type MatrixRow struct {
	TaxonName string
	States    []CharacterState
}

func (r MatrixRow) String() string {
	var buf bytes.Buffer
	buf.WriteString(r.TaxonName)
	buf.WriteString("\t")
	for _, state := range r.States {
		buf.WriteString(state.Value)
	}
	return buf.String()
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
		case "TITLE":
			tokens, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			if len(tokens) > 0 {
				c.Title = strings.Join(tokens, " ")
			}
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
			parseFormatCommand(tokens, &c.Format)
		case "CHARSTATELABELS":
			tokens, err := readUntilSemicolon(s)
			if err != nil {
				return err
			}
			parseCharStateLabels(tokens, c.CharStateLabels)
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

	// Create an anonymous struct to feed the template
	templateData := struct {
		Title        string
		Dimensions   Dimensions
		Format       Format
		SortedLabels []labelPair
		Matrix       []MatrixRow
	}{
		Title:        c.Title,
		Dimensions:   c.Dimensions,
		Format:       c.Format,
		SortedLabels: sortedLabels,
		Matrix:       c.Matrix,
	}

	// Render the template
	var buf bytes.Buffer
	if err := charsTmpl.Execute(&buf, templateData); err != nil {
		return "[ERROR rendering CHARACTERS block: " + err.Error() + "]\n"
	}
	return buf.String()
}

// -----------------------------------------------------------------------------
// Parsing Helpers
// -----------------------------------------------------------------------------

func parseFormatCommand(tokens []string, format *Format) {
	for i := 0; i < len(tokens); i++ {
		key := strings.ToUpper(tokens[i])

		valIdx := i + 1
		if valIdx < len(tokens) && tokens[valIdx] == "=" {
			valIdx++
			i = valIdx
		}

		if valIdx < len(tokens) {
			val := tokens[valIdx]
			switch key {
			case "DATATYPE":
				format.DataType = DataType(strings.ToUpper(val))
			case "MISSING":
				format.Missing = val
			case "GAP":
				format.Gap = val
			case "SYMBOLS":
				format.Symbols = strings.Trim(val, "\"'")
			}
		}
	}
}

// parseCharStateLabels extracts the character name, ignoring states after the slash.
func parseCharStateLabels(tokens []string, labels map[int]string) {
	var currentID int
	for _, t := range tokens {
		if t == "," {
			continue // Skip commas
		}
		if id, err := strconv.Atoi(t); err == nil {
			currentID = id
		} else if currentID != 0 {
			if _, exists := labels[currentID]; !exists {
				// Strip states (e.g., "eye_color/red" -> "eye_color")
				name := strings.Split(t, "/")[0]
				labels[currentID] = name
			}
		}
	}
}

// parseMatrix maps individual character state runes into the CharacterState struct.
func parseMatrix(s *Scanner, chars *CharactersBlock) error {
	for {
		taxonToken, err := s.NextToken()
		if err != nil {
			return err
		}
		if taxonToken == ";" {
			break
		}

		dataToken, err := s.NextToken()
		if err != nil {
			return err
		}

		row := MatrixRow{
			TaxonName: taxonToken,
			States:    make([]CharacterState, 0, len(dataToken)),
		}

		// Individualize each state value
		for i, r := range dataToken {
			index := i + 1 // NEXUS characters are 1-indexed
			row.States = append(row.States, CharacterState{
				Index: index,
				Value: string(r),
				Label: chars.CharStateLabels[index],
			})
		}

		chars.Matrix = append(chars.Matrix, row)
	}
	return nil
}
