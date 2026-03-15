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

// StateType tracks if a character is a single state, polymorphic, or uncertain.
type StateType string

const (
	StateSingle      StateType = "SINGLE"
	StatePolymorphic StateType = "POLYMORPHIC"
	StateUncertain   StateType = "UNCERTAIN"
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
	{{.PaddedName}}{{range .States}}{{.Render}}{{end}}
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
				Missing: "?",
				Gap:     "-",
			},
			CharStateLabels: make(map[int]string),
		}
	})
}

// CharactersBlock defines characters and includes character data.
type CharactersBlock struct {
	nexus           *Nexus
	Title           string
	Dimensions      Dimensions
	Format          Format
	CharStateLabels map[int]string
	Matrix          []MatrixRow
}

// NewCharactersBlock creates, appends, and returns a new CHARACTERS block.
func (n *Nexus) NewCharactersBlock(dt DataType) *CharactersBlock {
	cb := &CharactersBlock{
		nexus: n,
		Format: Format{
			DataType: dt,
			Missing:  "?",
			Gap:      "-",
		},
		CharStateLabels: make(map[int]string),
	}
	n.Blocks = append(n.Blocks, cb)
	return cb
}

// SetNexus implements the NexusAware interface.
func (c *CharactersBlock) SetNexus(n *Nexus) {
	c.nexus = n
}

// SetTitle applies a title to the block.
func (c *CharactersBlock) SetTitle(title string) {
	c.Title = title
}

// AddCharStateLabel registers a label for a specific character index [cite: 518-523].
// It automatically expands NCHAR if the index is larger than the current dimension.
func (c *CharactersBlock) AddCharStateLabel(index int, name string) {
	c.CharStateLabels[index] = name
	if index > c.Dimensions.NChar {
		c.Dimensions.NChar = index
	}
}

// AddRow adds a new sequence row to the matrix and auto-registers the taxon [cite: 560-565].
// States can be passed individually (e.g., "A", "C", "(AG)").
func (c *CharactersBlock) AddRow(taxonName string, states ...string) {
	// 1. Auto-sync with the parent TAXA block [cite: 344-345]
	if c.nexus != nil {
		c.nexus.RegisterTaxon(taxonName)
	}

	// 2. Prepare the new matrix row
	row := MatrixRow{
		TaxonName: taxonName,
		States:    make([]CharacterState, 0, len(states)),
	}

	// 3. Smartly parse the incoming states
	for i, stateStr := range states {
		index := i + 1 // NEXUS is 1-indexed

		cs := CharacterState{
			Index: index,
			Label: c.CharStateLabels[index],
		}

		// Clean up the string in case user passed brackets
		cleanStr := strings.Trim(stateStr, "(){} ")

		// Infer state type [cite: 577, 591-592]
		if strings.HasPrefix(stateStr, "(") {
			cs.Type = StatePolymorphic
			cs.Value = strings.Split(cleanStr, "")
		} else if strings.HasPrefix(stateStr, "{") {
			cs.Type = StateUncertain
			cs.Value = strings.Split(cleanStr, "")
		} else if len(cleanStr) > 1 {
			// If they pass "01" without brackets, we assume polymorphism
			cs.Type = StatePolymorphic
			cs.Value = strings.Split(cleanStr, "")
		} else {
			cs.Type = StateSingle
			cs.Value = []string{cleanStr}
		}

		row.States = append(row.States, cs)
	}

	c.Matrix = append(c.Matrix, row)

	// 4. Ensure NCHAR matches the longest sequence [cite: 380]
	if len(states) > c.Dimensions.NChar {
		c.Dimensions.NChar = len(states)
	}
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
	Index int
	Type  StateType
	Value []string
	Label string
}

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

// MatrixRow holds the sequence/data for a single taxon.
type MatrixRow struct {
	TaxonName string
	States    []CharacterState
}

// Parse implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Parse(s *Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}

		cmd := strings.ToUpper(token)
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

	maxLen := 0
	for _, row := range c.Matrix {
		if len(row.TaxonName) > maxLen {
			maxLen = len(row.TaxonName)
		}
	}

	type templateMatrixRow struct {
		PaddedName string
		States     []CharacterState
	}

	var tmplMatrix []templateMatrixRow
	for _, row := range c.Matrix {
		padding := strings.Repeat(" ", (maxLen-len(row.TaxonName))+2)
		tmplMatrix = append(tmplMatrix, templateMatrixRow{
			PaddedName: row.TaxonName + padding,
			States:     row.States,
		})
	}

	templateData := struct {
		Title        string
		Dimensions   Dimensions
		Format       Format
		SortedLabels []labelPair
		Matrix       []templateMatrixRow
	}{
		Title:        c.Title,
		Dimensions:   c.Dimensions,
		Format:       c.Format,
		SortedLabels: sortedLabels,
		Matrix:       tmplMatrix,
	}

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

func parseCharStateLabels(tokens []string, labels map[int]string) {
	var currentID int
	for _, t := range tokens {
		if t == "," {
			continue
		}
		if id, err := strconv.Atoi(t); err == nil {
			currentID = id
		} else if currentID != 0 {
			if _, exists := labels[currentID]; !exists {
				name := strings.Split(t, "/")[0]
				labels[currentID] = name
			}
		}
	}
}

// parseMatrix maps individual character state runes into the CharacterState struct
// while gracefully tracking polymorphism () and uncertainty {}.
func parseMatrix(s *Scanner, chars *CharactersBlock) error {
	nchar := chars.Dimensions.NChar
	for {
		taxonToken, err := s.NextToken()
		if err != nil {
			return err
		}
		if taxonToken == ";" {
			break
		}

		row := MatrixRow{
			TaxonName: taxonToken,
			States:    make([]CharacterState, 0, nchar),
		}

		parsedStates := 0
		for parsedStates < nchar {
			tok, err := s.NextToken()
			if err != nil {
				return err
			}

			// Handle Polymorphism and Uncertainty
			if tok == "(" || tok == "{" {
				stateType := StatePolymorphic
				closingTok := ")"
				if tok == "{" {
					stateType = StateUncertain
					closingTok = "}"
				}

				var values []string
				for {
					innerTok, err := s.NextToken()
					if err != nil {
						return err
					}
					if innerTok == closingTok {
						break
					}
					// Deconstruct smushed tokens inside the brackets (e.g. "AB" -> "A", "B")
					for _, r := range innerTok {
						values = append(values, string(r))
					}
				}

				index := parsedStates + 1
				row.States = append(row.States, CharacterState{
					Index: index,
					Type:  stateType,
					Value: values,
					Label: chars.CharStateLabels[index],
				})
				parsedStates++
			} else {
				// Handle standard tokens (which might be long strings like "ACATA")
				for _, r := range tok {
					if parsedStates >= nchar {
						break // Protect against malformed files exceeding NCHAR
					}
					index := parsedStates + 1
					row.States = append(row.States, CharacterState{
						Index: index,
						Type:  StateSingle,
						Value: []string{string(r)},
						Label: chars.CharStateLabels[index],
					})
					parsedStates++
				}
			}
		}
		chars.Matrix = append(chars.Matrix, row)
	}
	return nil
}
