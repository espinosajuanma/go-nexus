package characters

import (
	"strconv"
	"strings"

	"github.com/espinosajuanma/nexus/core"
)

// AddCharacter registers a new character (column) to the block.
func (c *CharactersBlock) AddCharacter(name string, states ...string) *Character {
	// Sanitize state labels
	sanitizedStates := make([]string, len(states))
	for i, st := range states {
		sanitizedStates[i] = core.DecodeName(st)
	}

	char := &Character{
		Index:       len(c.Characters) + 1,
		Name:        core.DecodeName(name),
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
	sanitizedName := core.DecodeName(name)

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
func (t *TaxonReference) AddCharacterState(char *Character, value string) *TaxonReference {
	if char.Index <= 0 || char.Index > len(t.parent.Characters) {
		return t
	}

	colIdx := char.Index - 1
	rowIdx := t.Index

	// Helper: Translate a label (e.g., "light red") into its symbol (e.g., "0")
	resolveSymbol := func(input string) string {
		sanitizedInput := core.DecodeName(input)
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
