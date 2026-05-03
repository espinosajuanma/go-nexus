package nexus

import (
	"bytes"
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
	CharStateLabels map[int]string // to be deprecated
	Matrix          []MatrixRow    // to be deprecated

	Characters []*Character
	Taxa       []*TaxonReference

	data [][]CharacterState
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

// AddCharacter registers a new character (column) to the block.
func (c *CharactersBlock) AddCharacter(name string, states ...string) *Character {
	// Sanitize state labels
	sanitizedStates := make([]string, len(states))
	for i, st := range states {
		sanitizedStates[i] = SanitizeName(st)
	}

	char := &Character{
		Index:       len(c.Characters) + 1,
		Name:        SanitizeName(name),
		StateLabels: sanitizedStates,
	}
	c.Characters = append(c.Characters, char)
	c.Dimensions.NChar = len(c.Characters)

	// Expand every existing taxon row to include a "Missing" state for this new character
	for i := range c.data {
		c.data[i] = append(c.data[i], CharacterState{
			Index: char.Index,
			Type:  StateSingle,
			Value: []string{c.Format.Missing},
		})
	}
	return char
}

// AddTaxon registers a new taxon (row) and returns a reference for adding states.
func (c *CharactersBlock) AddTaxon(name string) *TaxonReference {
	sanitizedName := SanitizeName(name)

	// Sync with global TAXA block
	if c.nexus != nil {
		c.nexus.RegisterTaxon(sanitizedName)
	}

	// Create the Taxon Reference
	taxon := &TaxonReference{
		Index:  len(c.Taxa),
		Name:   sanitizedName,
		parent: c,
	}
	c.Taxa = append(c.Taxa, taxon)

	// Initialize the matrix row with default "Missing" values for all current characters
	newRow := make([]CharacterState, len(c.Characters))
	for i, char := range c.Characters {
		newRow[i] = CharacterState{
			Index: char.Index,
			Type:  StateSingle,
			Value: []string{c.Format.Missing},
		}
	}
	c.data = append(c.data, newRow)

	return taxon
}

// AddCharacterState allows setting a value at a specific character for this taxon.
// AddCharacterState allows setting a value at a specific character for this taxon.
func (t *TaxonReference) AddCharacterState(char *Character, value string) *TaxonReference {
	if char.Index <= 0 || char.Index > len(t.parent.Characters) {
		return t
	}

	colIdx := char.Index - 1
	rowIdx := t.Index

	// Helper: Translate a label (e.g., "light red") into its symbol (e.g., "0")
	resolveSymbol := func(input string) string {
		sanitizedInput := SanitizeName(input)
		for i, label := range char.StateLabels {
			if sanitizedInput == label {
				return strconv.Itoa(i) // Found the label, return its index as the symbol
			}
		}
		// If not found in labels, assume it's already a valid symbol (like "?", "-", or "0")
		return sanitizedInput
	}

	cs := CharacterState{
		Index: char.Index,
		Label: char.Name,
	}

	// Clean and split the incoming value
	cleanVal := strings.Trim(value, "(){} ")

	// Use strings.Split for comma-separated or strings.Fields for space-separated
	// Assuming fields (spaces) based on the current logic
	var rawValues []string
	if strings.Contains(cleanVal, " ") {
		rawValues = strings.Fields(cleanVal)
	} else {
		// Handle smushed symbols if the user passed something like "01" instead of "0 1"
		// If it's a known label like "light_red", it won't hit this branch unless it has no spaces
		rawValues = []string{cleanVal}
	}

	// Resolve all extracted values into their proper symbols
	var resolvedSymbols []string
	for _, rv := range rawValues {
		resolvedSymbols = append(resolvedSymbols, resolveSymbol(rv))
	}

	// Determine the state type based on original brackets
	if strings.HasPrefix(value, "(") {
		cs.Type = StatePolymorphic
	} else if strings.HasPrefix(value, "{") {
		cs.Type = StateUncertain
	} else if len(resolvedSymbols) > 1 {
		// Catch-all: if multiple symbols are resolved without brackets, assume polymorphism
		cs.Type = StatePolymorphic
	} else {
		cs.Type = StateSingle
	}

	cs.Value = resolvedSymbols
	t.parent.data[rowIdx][colIdx] = cs

	return t
}

// Dimensions captures the NCHAR value for the CHARACTERS block.
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

// Character represents a single column in the NEXUS matrix.
type Character struct {
	Index       int
	Name        string
	StateLabels []string
}

// TaxonReference represents a single row in the NEXUS matrix.
type TaxonReference struct {
	Index  int
	Name   string
	parent *CharactersBlock // Back-pointer for fluent API
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
			c.Title = strings.Join(tokens, " ")

		case "DIMENSIONS":
			// Extract NCHAR to initialize our Character objects
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
						// Pre-populate Character objects
						for j := 1; j <= count; j++ {
							c.Characters = append(c.Characters, &Character{Index: j})
						}
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
			parseCharStateLabelsRelational(tokens, c)

		case "MATRIX":
			if err := parseMatrixRelational(s, c); err != nil {
				return err
			}

		default:
			// Skip unrecognized commanas
			if _, err := readUntilSemicolon(s); err != nil {
				return err
			}
		}
	}
}

// Render implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Render() string {
	// Prepare and sort Character Labels for the CHARSTATELABELS command
	type labelView struct {
		ID     int
		Name   string
		States string
	}
	var sortedLabels []labelView

	for _, char := range c.Characters {
		// Only render if the character has a name OR state labels
		if char.Name != "" || len(char.StateLabels) > 0 {
			statesStr := ""
			if len(char.StateLabels) > 0 {
				var safeStates []string
				for _, st := range char.StateLabels {
					if st == "" {
						safeStates = append(safeStates, "_") // Use underscore for unnamed states
					} else {
						safeStates = append(safeStates, st)
					}
				}
				statesStr = strings.Join(safeStates, " ")
			}

			sortedLabels = append(sortedLabels, labelView{
				ID:     char.Index,
				Name:   char.Name,
				States: statesStr,
			})
		}
	}

	// Calculate the longest taxon name for matrix alignment
	maxTaxonLen := 0
	for _, taxon := range c.Taxa {
		if len(taxon.Name) > maxTaxonLen {
			maxTaxonLen = len(taxon.Name)
		}
	}

	// Flatten the 2D data into a View Model for the template
	type templateRow struct {
		PaddedName string
		States     []CharacterState
	}
	var rows []templateRow

	for i, taxon := range c.Taxa {
		// Calculate dynamic padding
		padding := strings.Repeat(" ", (maxTaxonLen-len(taxon.Name))+2)

		rows = append(rows, templateRow{
			PaddedName: taxon.Name + padding,
			States:     c.data[i],
		})
	}

	// Populate the final structure for the Go Template engine
	templateData := struct {
		Title        string
		Dimensions   Dimensions
		Format       Format
		SortedLabels []labelView
		Matrix       []templateRow
	}{
		Title:        c.Title,
		Dimensions:   c.Dimensions,
		Format:       c.Format,
		SortedLabels: sortedLabels,
		Matrix:       rows,
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

// parseCharStateLabelsRelational handles the CHARSTATELABELS command in a relational format.
func parseCharStateLabelsRelational(tokens []string, c *CharactersBlock) {
	var currentID int
	readingStates := false

	for _, t := range tokens {
		// A comma resets the reader for the next character
		if t == "," {
			readingStates = false
			currentID = 0
			continue
		}

		// Catch the character ID
		if id, err := strconv.Atoi(t); err == nil && !readingStates {
			currentID = id
			continue
		}

		// Toggle state-reading mode when we hit a slash
		if t == "/" {
			readingStates = true
			continue
		}

		// Apply names and states to the specific Character object
		if currentID != 0 && currentID <= len(c.Characters) {
			char := c.Characters[currentID-1]

			if !readingStates {
				char.Name = t
			} else {
				stateName := t
				if stateName == "_" {
					stateName = "" // Translate NEXUS underscores back into empty strings internally
				}
				char.StateLabels = append(char.StateLabels, stateName)
			}
		}
	}
}

// parseMatrixRelational handles the MATRIX command in a relational format, supporting polymorphic and uncertain states.
func parseMatrixRelational(s *Scanner, c *CharactersBlock) error {
	nchar := c.Dimensions.NChar
	for {
		taxonToken, err := s.NextToken()
		if err != nil {
			return err
		}
		if taxonToken == ";" {
			break
		}

		// Register taxon and get a reference to populate states
		taxonRef := c.AddTaxon(taxonToken)

		parsedStates := 0
		for parsedStates < nchar {
			tok, err := s.NextToken()
			if err != nil {
				return err
			}

			// Handle Grouped States (Polymorphism () or Uncertainty {})
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
					// Deconstruct tokens inside brackets
					for _, r := range innerTok {
						values = append(values, string(r))
					}
				}

				charObj := c.Characters[parsedStates]
				taxonRef.AddCharacterState(charObj, formatValueForBuilder(stateType, values))
				parsedStates++
			} else {
				// Standard symbols or smushed sequence strings
				for _, r := range tok {
					if parsedStates >= nchar {
						break
					}
					charObj := c.Characters[parsedStates]
					taxonRef.AddCharacterState(charObj, string(r))
					parsedStates++
				}
			}
		}
	}
	return nil
}

// formatValueForBuilder ensures state groups are passed with their brackets intact
// so the AddCharacterState logic can re-parse them correctly.
func formatValueForBuilder(st StateType, vals []string) string {
	s := strings.Join(vals, " ")
	if st == StatePolymorphic {
		return "(" + s + ")"
	}
	if st == StateUncertain {
		return "{" + s + "}"
	}
	return s
}

// parseFormatCommand processes the tokens from a FORMAT command and updates the Format struct accordingly.
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
