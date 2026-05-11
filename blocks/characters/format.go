package characters

import (
	"fmt"
	"strings"
)

// Format specifies the format of the data MATRIX.
type Format struct {
	DataType       DataType
	Missing        string
	Gap            string
	Symbols        []string
	Equate         map[string]string
	MatchChar      string
	RespectCase    bool
	Interleave     bool
	InterleaveSize int
	Tokens         bool
	Labels         bool
	Transpose      bool
	Items          string
	StatesFormat   string
	NStates        int
}

// String translates the Format AST back into a NEXUS string.
func (f Format) String() string {
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
