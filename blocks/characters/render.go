package characters

import (
	_ "embed"
	"strings"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/templater"
)

//go:embed characters.tmpl
var charsTmplStr string

// Render implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Render() (string, error) {
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
						safeStates = append(safeStates, core.QuoteName(st))
					}
				}
				statesStr = strings.Join(safeStates, " ")
			}

			sortedLabels = append(sortedLabels, labelView{
				ID:     char.Index,
				Name:   core.QuoteName(char.Name),
				States: statesStr,
			})
		}
	}

	// Calculate the longest taxon name for matrix alignment
	maxTaxonLen := 0
	for _, taxon := range c.Taxa {
		if len(taxon.Name) > maxTaxonLen {
			maxTaxonLen = len(core.EncodeName(taxon.Name))
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
		taxonName := core.EncodeName(taxon.Name)
		padding := strings.Repeat(" ", (maxTaxonLen-len(taxonName))+2)

		rows = append(rows, templateRow{
			PaddedName: taxonName + padding,
			States:     c.data[i],
		})
	}

	// Populate the final structure for the Go Template engine
	templateData := struct {
		Title        string
		Dimensions   int
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

	tmpl, err := templater.New("characters", charsTmplStr)
	if err != nil {
		return "", err
	}
	rendered, err := tmpl.Render(templateData)

	return rendered, err
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
