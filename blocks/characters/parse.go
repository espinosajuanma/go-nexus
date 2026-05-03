package characters

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/espinosajuanma/nexus/parser"
	"github.com/espinosajuanma/nexus/scanner"
)

// Parse implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Parse(s *scanner.Scanner) error {
	for {
		token, err := s.NextToken()
		if err != nil {
			return err
		}

		cmd := strings.ToUpper(token)
		if cmd == "END" || cmd == "ENDBLOCK" {
			return parser.ExpectSemicolon(s)
		}

		switch cmd {
		case "TITLE":
			tokens, err := parser.ReadUntilSemicolon(s)
			if err != nil {
				return err
			}
			c.Title = strings.Join(tokens, " ")

		case "DIMENSIONS":
			// Extract NCHAR to initialize our Character objects
			tokens, err := parser.ReadUntilSemicolon(s)
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
			tokens, err := parser.ReadUntilSemicolon(s)
			if err != nil {
				return err
			}
			parseFormatCommand(tokens, &c.Format)

		case "CHARSTATELABELS":
			tokens, err := parser.ReadUntilSemicolon(s)
			if err != nil {
				return err
			}
			parseCharStateLabelsRelational(tokens, c)

		case "MATRIX":
			if err := parseMatrixRelational(s, c); err != nil {
				return err
			}

		default:
			// Skip unrecognized commands
			if _, err := parser.ReadUntilSemicolon(s); err != nil {
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
func parseMatrixRelational(s *scanner.Scanner, c *CharactersBlock) error {
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
