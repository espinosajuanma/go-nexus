package characters

import (
	_ "embed"

	"github.com/espinosajuanma/go-nexus/templater"
)

//go:embed characters.tmpl
var charsTmplStr string

const name = "CHARACTERS"

// Render implements the Block interface for CharactersBlock.
func (c *CharactersBlock) Render() (string, error) {
	tmpl, err := templater.New(name, charsTmplStr)
	if err != nil {
		return "", err
	}

	return tmpl.Render(c)
}

// HasCharStateLabels determines if the CHARSTATELABELS command is needed.
func (c *CharactersBlock) HasCharStateLabels() bool {
	for _, char := range c.Matrix.Characters {
		if char.Name != "" || len(char.StateLabels) > 0 {
			return true
		}
	}
	return false
}

// MaxTaxonNameLength calculates the length of the longest taxon name for padding matrix rows.
func (c *CharactersBlock) MaxTaxonNameLength() int {
	max := 0
	for _, t := range c.Matrix.Taxa {
		if len(t.Name) > max {
			max = len(t.Name)
		}
	}
	return max
}

// InterleavedChunks calculates the exact segment bounds for interleaved matrices.
func (c *CharactersBlock) InterleavedChunks(chunkSize int) []MatrixChunk {
	var chunks []MatrixChunk
	for i := 0; i < c.Dimensions; i += chunkSize {
		end := i + chunkSize
		if end > c.Dimensions {
			end = c.Dimensions
		}
		chunks = append(chunks, MatrixChunk{Start: i, End: end})
	}
	return chunks
}
