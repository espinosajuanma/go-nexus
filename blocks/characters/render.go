package characters

import (
	_ "embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/espinosajuanma/nexus/templater"
)

//go:embed characters.tmpl
var charsTmplStr string

const name = "CHARACTERS"

// Render implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Render() (string, error) {
	// Define block-specific template helpers
	customFuncs := template.FuncMap{
		"formatState": formatState,
		"printFormat": printFormat,
	}

	// Inject the custom functions into the generic templater
	tmpl, err := templater.New(name, charsTmplStr, customFuncs)
	if err != nil {
		return "", err
	}

	return tmpl.Render(c)
}

// Data is exported so the text/template can access the private matrix slice.
func (c *CharactersBlock) Data() [][]CharacterState {
	return c.data
}

// CalculateChunks returns a slice of starting indices [0, 70, 140...] for interleaved rendering.
func (c *CharactersBlock) CalculateChunks(nchar, chunkSize int) []int {
	var chunks []int
	for i := 0; i < nchar; i += chunkSize {
		chunks = append(chunks, i)
	}
	return chunks
}

// CalculateEndCol safely calculates the end slice index for interleaved segments.
func (c *CharactersBlock) CalculateEndCol(start, size, max int) int {
	end := start + size
	if end > max {
		return max
	}
	return end
}

// -----------------------------------------------------------------------------
// Template Helpers
// -----------------------------------------------------------------------------

// formatState translates internal CharacterState sentinels back to file-ready text.
func formatState(state CharacterState, format Format) string {
	var resolved []string
	for _, v := range state.Value {
		switch v {
		case InternalMissing:
			resolved = append(resolved, format.Missing)
		case InternalGap:
			resolved = append(resolved, format.Gap)
		default:
			resolved = append(resolved, v)
		}
	}

	if state.Type == StatePolymorphic {
		return "(" + strings.Join(resolved, " ") + ")"
	}
	if state.Type == StateUncertain {
		return "{" + strings.Join(resolved, " ") + "}"
	}

	return resolved[0]
}

// printFormat constructs the FORMAT string, only printing non-default values.
func printFormat(f Format) string {
	var parts []string

	if f.DataType != "" {
		parts = append(parts, fmt.Sprintf("DATATYPE=%s", f.DataType))
	}
	if f.Missing != "" && f.Missing != "?" {
		parts = append(parts, fmt.Sprintf("MISSING=%s", f.Missing))
	}
	if f.Gap != "" && f.Gap != "-" {
		parts = append(parts, fmt.Sprintf("GAP=%s", f.Gap))
	}
	if f.MatchChar != "" {
		parts = append(parts, fmt.Sprintf("MATCHCHAR=%s", f.MatchChar))
	}
	if f.RespectCase {
		parts = append(parts, "RESPECTCASE")
	}
	if f.Interleave {
		parts = append(parts, "INTERLEAVE")
	}
	if f.Tokens {
		if f.DataType != Standard {
			parts = append(parts, "TOKENS")
		}
	}
	if !f.Labels {
		parts = append(parts, "LABELS=NO")
	}

	if len(parts) == 0 {
		return ""
	}
	return "\tFORMAT " + strings.Join(parts, " ") + ";"
}

// MaxTaxonNameLength calculates the length of the longest taxon name for padding matrix rows.
func (c *CharactersBlock) MaxTaxonNameLength() int {
	max := 0
	for _, t := range c.Taxa {
		// Length of the raw name (snake casting won't change string length)
		if len(t.Name) > max {
			max = len(t.Name)
		}
	}
	return max
}
