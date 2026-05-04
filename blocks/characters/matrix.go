package characters

import (
	"strconv"
	"strings"

	"github.com/espinosajuanma/nexus/core"
)

// AddCharacter registers a new character and initializes it with the InternalMissing sentinel.
func (c *CharactersBlock) AddCharacter(name string, states ...string) *Character {
	char := &Character{
		Index:       len(c.Characters) + 1,
		Name:        core.DecodeName(name),
		StateLabels: states,
	}
	c.Characters = append(c.Characters, char)
	c.Dimensions = len(c.Characters)

	// Fill existing rows with the Missing sentinel
	for i := range c.data {
		c.data[i] = append(c.data[i], CharacterState{
			Index: char.Index,
			Type:  StateSingle,
			Value: []string{InternalMissing}, // Decoupled sentinel
		})
	}
	return char
}

// AddTaxon registers a new taxon and prepares its matrix row.
func (c *CharactersBlock) AddTaxon(name string) *TaxonReference {
	sanitizedName := core.DecodeName(name)

	taxon := &TaxonReference{
		Index:  len(c.Taxa),
		Name:   sanitizedName,
		parent: c,
	}
	c.Taxa = append(c.Taxa, taxon)

	// Pre-fill the row based on current character count
	newRow := make([]CharacterState, c.Dimensions)
	for i := 0; i < c.Dimensions; i++ {
		newRow[i] = CharacterState{
			Index: i + 1,
			Type:  StateSingle,
			Value: []string{InternalMissing},
		}
	}
	c.data = append(c.data, newRow)

	return taxon
}

// ResolveStateSymbol takes a user input (which could be a label or a direct symbol) and returns the correct internal symbol for storage.
func (c *CharactersBlock) ResolveStateSymbol(char *Character, state string) string {
	// Check if it's one of our internal sentinels
	if state == InternalMissing || state == InternalGap {
		return state
	}

	// Search the character's defined StateLabels
	for i, label := range char.StateLabels {
		if strings.EqualFold(label, state) {
			// Found it! Map its index to the allowed FORMAT symbols
			if i < len(c.Format.Symbols) {
				return string(c.Format.Symbols[i])
			}
			// Fallback if symbols aren't explicitly defined
			return strconv.Itoa(i)
		}
	}

	// If no label matched, assume the user passed a raw symbol directly (e.g., "A", "0")
	return state
}

// SetState safely registers character states for a taxon, translating labels to symbols.
// It accepts multiple states to support Polymorphism (e.g., passing "blue", "green").
func (t *TaxonReference) SetState(char *Character, states ...string) error {
	var resolved []string

	// Translate every provided state string into its proper matrix symbol
	for _, state := range states {
		resolved = append(resolved, t.parent.ResolveStateSymbol(char, state))
	}

	stateType := StateSingle
	if len(resolved) > 1 {
		stateType = StatePolymorphic
	}

	// Update the matrix grid
	t.parent.data[t.Index][char.Index-1] = CharacterState{
		Index: char.Index,
		Type:  stateType,
		Value: resolved,
	}
	return nil
}
