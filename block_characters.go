package nexus

import (
	"fmt"
	"strconv"
	"strings"
)

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
	var b strings.Builder
	b.WriteString("BEGIN CHARACTERS;\n")
	b.WriteString(fmt.Sprintf("\tDIMENSIONS NCHAR=%d;\n", c.Dimensions.NChar))

	// Format command setup
	var formatArgs []string
	if c.Format.DataType != "" {
		formatArgs = append(formatArgs, fmt.Sprintf("DATATYPE=%s", c.Format.DataType))
	}
	if c.Format.Missing != "" {
		formatArgs = append(formatArgs, fmt.Sprintf("MISSING=%s", c.Format.Missing))
	}
	if c.Format.Gap != "" {
		formatArgs = append(formatArgs, fmt.Sprintf("GAP=%s", c.Format.Gap))
	}
	if c.Format.Symbols != "" {
		formatArgs = append(formatArgs, fmt.Sprintf("SYMBOLS=\"%s\"", c.Format.Symbols))
	}
	if len(formatArgs) > 0 {
		b.WriteString(fmt.Sprintf("\tFORMAT %s;\n", strings.Join(formatArgs, " ")))
	}

	// Charstatelabels
	if len(c.CharStateLabels) > 0 {
		b.WriteString("\tCHARSTATELABELS\n")
		for idx, name := range c.CharStateLabels {
			b.WriteString(fmt.Sprintf("\t\t%d %s,\n", idx, name))
		}
		b.WriteString("\t;\n")
	}

	// Matrix
	b.WriteString("\tMATRIX\n")
	for _, row := range c.Matrix {
		b.WriteString(fmt.Sprintf("\t%s\t%s\n", row.TaxonName, row.Data))
	}
	b.WriteString("\t;\nEND;\n\n")

	return b.String()
}
