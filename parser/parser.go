package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/espinosajuanma/nexus/core"
	"github.com/espinosajuanma/nexus/scanner"
)

// Parse reads a NEXUS format file from an io.Reader and populates the Core struct.
func Parse(r io.Reader) (*core.Core, error) {
	scanner := scanner.NewScanner(r)
	nex := &core.Core{
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

			name := strings.ToUpper(blockName)

			factory, exists := core.BlockRegistry[name]
			if exists {
				// Create a new instance of the block
				block := factory(name)

				// Tell the block to parse its own contents
				if err := block.Parse(scanner); err != nil {
					return nil, fmt.Errorf("error parsing %s block: %w", blockName, err)
				}
				nex.Blocks = append(nex.Blocks, block)
			} else {
				factory, exist := core.BlockRegistry["GENERIC"]
				if !exist {
					return nil, fmt.Errorf("unknown block type '%s' and no GENERIC block registered", blockName)
				}
				block := factory(name)
				if err := block.Parse(scanner); err != nil {
					return nil, fmt.Errorf("error parsing GENERIC block for unknown block %s: %w", blockName, err)
				}
				nex.Blocks = append(nex.Blocks, block)
			}
		}
	}

	return nex, nil
}
