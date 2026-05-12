package characters

import (
	"fmt"
	"strings"
)

// StateType tracks if a character is a single state, polymorphic, or uncertain.
type StateType string

const (
	StateSingle      StateType = "SINGLE"
	StatePolymorphic StateType = "POLYMORPHIC"
	StateUncertain   StateType = "UNCERTAIN"
	StateMissing     StateType = "MISSING"
	StateGap         StateType = "GAP"
)

// CharacterState holds individualized information about a single character state.
type CharacterState struct {
	Index  int
	Type   StateType
	Values []StateValue
}

// StateValue holds a parsed state symbol and its associated value
// Handles STATESFORMAT=COUNT (e.g., "A:2") or FREQUENCY (e.g., "B:0.40").
type StateValue struct {
	Symbol string
	Weight float64 // Defaults to 1.0. Holds count or frequency if provided.
}

// String translates the AST CharacterState back to file-ready text.
func (cs CharacterState) String(format Format) string {
	// Handle absolute flags first
	if cs.Type == StateMissing {
		return format.Missing
	}
	if cs.Type == StateGap {
		return format.Gap
	}

	var resolved []string
	for _, val := range cs.Values {
		str := val.Symbol
		// If dealing with frequencies/counts, format them back
		if val.Weight != 1.0 && val.Weight != 0.0 {
			str = fmt.Sprintf("%s:%v", val.Symbol, val.Weight)
		}
		resolved = append(resolved, str)
	}

	if cs.Type == StatePolymorphic {
		return "(" + strings.Join(resolved, " ") + ")"
	}
	if cs.Type == StateUncertain {
		return "{" + strings.Join(resolved, " ") + "}"
	}

	if len(resolved) > 0 {
		return resolved[0]
	}
	return format.Missing
}
