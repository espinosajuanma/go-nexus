package characters

import (
	"strconv"
	"strings"

	"github.com/espinosajuanma/go-nexus/utils"
)

type Matrix struct {
	Characters []*Character
	Taxa       []*Taxon
	data       [][]CharacterState
	parent     *CharactersBlock
}

// GetStateByIndex retrieves the CharacterState for a given taxon and character index.
func (m *Matrix) GetStateByIndex(taxonIndex, charIndex int) CharacterState {
	if taxonIndex >= 0 && taxonIndex < len(m.data) && charIndex >= 0 && charIndex < len(m.data[taxonIndex]) {
		return m.data[taxonIndex][charIndex]
	}
	return CharacterState{
		Index: charIndex + 1,
		Type:  StateSingle,
	}
}

// SetStateByIndex allows direct manipulation of the matrix grid using taxon and character indices.
func (m *Matrix) SetStateByIndex(taxonIndex, charIndex int, cs CharacterState) {
	if taxonIndex >= 0 && taxonIndex < len(m.data) && charIndex >= 0 && charIndex < len(m.data[taxonIndex]) {
		m.data[taxonIndex][charIndex] = cs
	}
}

// GetState is a convenient wrapper that allows retrieving states using taxon and character references.
func (m *Matrix) GetState(taxon *Taxon, char *Character) CharacterState {
	return m.GetStateByIndex(taxon.Index, char.Index-1)
}

// SetState is a convenient wrapper that allows setting states using taxon and character references.
func (m *Matrix) SetState(taxon *Taxon, char *Character, cs CharacterState) {
	m.SetStateByIndex(taxon.Index, char.Index-1, cs)
}

// Character represents a single column in the NEXUS matrix.
type Character struct {
	Index       int
	Name        string
	StateLabels []string
}

// Taxon represents a single row in the NEXUS matrix.
type Taxon struct {
	Index  int
	Name   string
	matrix *Matrix // Back-pointer for fluent API
	block  *CharactersBlock
}

func (t *Taxon) SetState(char *Character, stateType StateType, states ...string) error {
	cs := CharacterState{
		Index: char.Index,
		Type:  stateType,
	}

	for _, s := range states {
		sym := t.block.Matrix.ResolveStateSymbol(char, s)
		cs.Values = append(cs.Values, StateValue{
			Symbol: sym,
			Weight: 1.0,
		})
	}

	t.matrix.SetState(t, char, cs)
	return nil
}

// MatrixChunk defines the start and end column indices for an interleaved block.
type MatrixChunk struct {
	Start int
	End   int
}

// AddCharacter registers a new character and initializes it as missing data.
func (m *Matrix) AddCharacter(name string, states ...string) *Character {
	char := &Character{
		Index:       len(m.Characters) + 1,
		Name:        utils.DecodeName(name),
		StateLabels: states,
	}
	m.Characters = append(m.Characters, char)
	m.parent.Dimensions = len(m.Characters)

	// Initialize the new character for all taxa as missing
	for i := range m.data {
		m.data[i] = append(m.data[i], CharacterState{
			Index: char.Index,
			Type:  StateMissing,
		})
	}
	return char
}

// AddTaxon registers a new taxon and prepares its matrix row.
func (m *Matrix) AddTaxon(name string) *Taxon {
	sanitizedName := utils.DecodeName(name)

	taxon := &Taxon{
		Index:  len(m.Taxa),
		Name:   sanitizedName,
		matrix: m, // Set back-pointer for fluent API
		block:  m.parent,
	}
	m.Taxa = append(m.Taxa, taxon)

	// Pre-fill the row based on current character count
	newRow := make([]CharacterState, m.parent.Dimensions)
	for i := 0; i < m.parent.Dimensions; i++ {
		newRow[i] = CharacterState{
			Index: i + 1,
			Type:  StateMissing,
		}
	}
	m.data = append(m.data, newRow)

	return taxon
}

// GetTaxon retrieves a Taxon by its name (case-insensitive).
func (m *Matrix) GetTaxon(name string) *Taxon {
	normalizedName := normalizeTaxonName(name)
	for _, t := range m.Taxa {
		if normalizeTaxonName(t.Name) == normalizedName {
			return t
		}
	}
	return nil
}

// GetCharacter retrieves a Character by its 1-based Index.
func (m *Matrix) GetCharacterByIndex(index int) *Character {
	if index > 0 && index <= len(m.Characters) {
		return m.Characters[index-1]
	}
	return nil
}

// GetState returns the current CharacterState for this taxon at the given character.
func (t *Taxon) GetState(char *Character) CharacterState {
	return t.matrix.GetState(t, char)
}

// ResolveStateSymbol translates labels to their symbol equivalent if needed.
func (m *Matrix) ResolveStateSymbol(char *Character, state string) string {
	for i, label := range char.StateLabels {
		if strings.EqualFold(label, state) {
			if i < len(m.parent.Format.Symbols) {
				return m.parent.Format.Symbols[i]
			}
			return strconv.Itoa(i)
		}
	}
	return state
}
