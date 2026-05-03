package nexus

import (
	"io"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/parser"

	_ "github.com/espinosajuanma/nexus/blocks/characters"
	_ "github.com/espinosajuanma/nexus/blocks/taxa"
	_ "github.com/espinosajuanma/nexus/blocks/trees"
)

// Expose the core types at the root level
type Nexus = core.Nexus
type Block = core.Block

// New creates a new, empty Nexus file container ready for building.
func New() *Nexus {
	return core.New()
}

// Parse reads a NEXUS format file from an io.Reader and populates the Nexus struct.
func Parse(r io.Reader) (*Nexus, error) {
	return parser.Parse(r)
}
