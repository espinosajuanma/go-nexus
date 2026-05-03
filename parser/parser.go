package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/scanner"
)

// Parse reads a NEXUS format file from an io.Reader and populates the Nexus struct.
func Parse(r io.Reader) (*core.Nexus, error) {
	scanner := scanner.NewScanner(r)
	nex := &core.Nexus{
		Blocks: make([]core.Block, 0),
	}

	token, err := scanner.NextToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	if strings.ToUpper(token) != "#NEXUS" {
		return nil, fmt.Errorf("invalid file format: expected #NEXUS, got %s", token)
	}

	for {
		token, err = scanner.NextToken()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if strings.ToUpper(token) == "BEGIN" {
			blockName, err := scanner.NextToken()
			if err != nil {
				return nil, fmt.Errorf("expected block name after BEGIN: %w", err)
			}

			if err := ExpectSemicolon(scanner); err != nil {
				return nil, err
			}

			factory, exists := core.BlockRegistry[strings.ToUpper(blockName)]
			if exists {
				// Create a new instance of the block
				block := factory()

				// Tell the block to parse its own contents
				if err := block.Parse(scanner); err != nil {
					return nil, fmt.Errorf("error parsing %s block: %w", blockName, err)
				}
				nex.Blocks = append(nex.Blocks, block)
			} else {
				// If not registered, modularity allows us to safely ignore it.
				if err := SkipBlock(scanner); err != nil {
					return nil, fmt.Errorf("failed to skip unknown block %s: %w", blockName, err)
				}
			}
		}
	}

	return nex, nil
}
