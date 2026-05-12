package xread

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/espinosajuanma/nexus/blocks/characters"
	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/templater"
)

//go:embed nona.tmpl
var nonaTemplate string

//go:embed tnt.tmpl
var tntTemplate string

// Variant defines the target software dialect.
type Variant string

const (
	NONA Variant = "NONA"
	TNT  Variant = "TNT"
)

type Row struct {
	Name     string
	Sequence string
}

type CharLabel struct {
	Index       int
	Name        string
	StateLabels []string
}

// Exporter unifies the export configuration for xread-based formats.
type Exporter struct {
	Nexus      *core.Nexus
	Variant    Variant
	Project    string
	Author     string
	Commands   []string // TNT specific
	UseTaxname bool     // TNT specific
}

// New creates a new Exporter wrapping the Nexus AST for a specific Variant.
func New(nex *core.Nexus, variant Variant) *Exporter {
	return &Exporter{
		Nexus:      nex,
		Variant:    variant,
		UseTaxname: variant == TNT, // TNT defaults to true to avoid truncation
	}
}

// SetProject adds project metadata to the header comment.
func (e *Exporter) SetProject(project string) *Exporter {
	e.Project = project
	return e
}

// SetAuthor adds author metadata to the header comment.
func (e *Exporter) SetAuthor(author string) *Exporter {
	e.Author = author
	return e
}

// SetTaxname toggles the "taxname=;" command injection (TNT only).
func (e *Exporter) SetTaxname(use bool) *Exporter {
	e.UseTaxname = use
	return e
}

// AddCommand appends an operational command before the matrix (TNT only).
func (e *Exporter) AddCommand(cmd string) *Exporter {
	e.Commands = append(e.Commands, cmd)
	return e
}

// -- Template Helper Methods --

func (e *Exporter) charBlock() (*characters.CharactersBlock, error) {
	cb, ok := core.GetBlock[*characters.CharactersBlock](e.Nexus)
	if !ok {
		return nil, fmt.Errorf("no CHARACTERS block found in the Nexus file")
	}
	return cb, nil
}

func (e *Exporter) Title() string {
	cb, err := e.charBlock()
	if err != nil {
		return ""
	}
	return cb.Title
}

func (e *Exporter) NChar() int {
	cb, err := e.charBlock()
	if err != nil || cb == nil {
		return 0
	}
	return cb.Dimensions
}

func (e *Exporter) NTax() int {
	cb, err := e.charBlock()
	if err != nil || cb == nil {
		return 0
	}
	return len(cb.Matrix.Taxa)
}

func (e *Exporter) Rows() []Row {
	cb, err := e.charBlock()
	if err != nil {
		return nil
	}

	maxTaxonLen := 0
	for _, taxon := range cb.Matrix.Taxa {
		encodedName := core.EncodeName(taxon.Name)
		if len(encodedName) > maxTaxonLen {
			maxTaxonLen = len(encodedName)
		}
	}

	var rows []Row
	for _, taxon := range cb.Matrix.Taxa {
		encodedName := core.EncodeName(taxon.Name)
		paddedName := fmt.Sprintf("%-*s", maxTaxonLen, encodedName)

		var seqBuilder strings.Builder
		for _, char := range cb.Matrix.Characters {
			state := taxon.GetState(char)

			var symbol string
			switch state.Type {
			case characters.StateMissing:
				symbol = "?"
			case characters.StateGap:
				symbol = "-"
			case characters.StateSingle:
				// Ensure we use Observations per your domain logic
				if len(state.Values) > 0 {
					symbol = state.Values[0].Symbol
				} else {
					symbol = "?"
				}
			case characters.StatePolymorphic, characters.StateUncertain:
				symbols := make([]string, len(state.Values))
				for i, val := range state.Values {
					symbols[i] = val.Symbol
				}
				symbol = "[" + strings.Join(symbols, "") + "]"
			default:
				symbol = "?"
			}
			seqBuilder.WriteString(symbol)
		}
		rows = append(rows, Row{Name: paddedName, Sequence: seqBuilder.String()})
	}
	return rows
}

// HasLabels checks if any character has a name or state labels, which determines if the CHARLABELS block should be rendered.
func (e *Exporter) HasLabels() bool {
	cb, err := e.charBlock()
	if err != nil {
		return false
	}
	for _, char := range cb.Matrix.Characters {
		if char.Name != "" || len(char.StateLabels) > 0 {
			return true
		}
	}
	return false
}

// Characters returns a slice of CharLabel structs for characters that have names or state labels, which are used to render the CHARLABELS block.
func (e *Exporter) Characters() []CharLabel {
	cb, err := e.charBlock()
	if err != nil {
		return nil
	}

	var chars []CharLabel
	for i, char := range cb.Matrix.Characters {
		if char.Name != "" || len(char.StateLabels) > 0 {
			charName := char.Name
			if charName == "" {
				charName = fmt.Sprintf("Char_%d", i)
			}

			var stateLabels []string
			for _, sl := range char.StateLabels {
				stateLabels = append(stateLabels, core.EncodeName(sl))
			}

			chars = append(chars, CharLabel{
				Index:       i,
				Name:        core.EncodeName(charName),
				StateLabels: stateLabels,
			})
		}
	}
	return chars
}

// Render executes the appropriate template based on the Variant.
func (e *Exporter) Render() (string, error) {
	var tmplStr string
	if e.Variant == TNT {
		tmplStr = tntTemplate
	} else {
		tmplStr = nonaTemplate
	}

	tmpl, err := templater.New(string(e.Variant), tmplStr)
	if err != nil {
		return "", err
	}

	return tmpl.Render(e)
}
