# Go Nexus

> ⚠️ Notice: This package is currently in an early development state. The API is
> unstable and subject to breaking changes without prior notice. Use with caution
> in production environments.

A modular Go library to parse, manipulate, and generate NEXUS (.nex)
phylogenetic files.

## Installation

```bash
go get github.com/espinosajuanma/go-nexus
```

## Usage

The library provides a simple API to read existing files or build new ones from
scratch. It features a smart wrapper to easily access and manipulate standard blocks.

### Parsing an existing file

```go
package main

import (
	"fmt"
	"os"

	"github.com/espinosajuanma/go-nexus"
)

func main() {
	file, err := os.Open("example.nex")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	nex, err := nexus.Parse(file)
	if err != nil {
		panic(err)
	}

	if taxa, ok := nex.GetTaxaBlock(); ok {
		fmt.Printf("Successfully parsed TAXA block with %d taxa\n", taxa.Dimensions)
	}
	
	if char, ok := nex.GetCharactersBlock(); ok {
		fmt.Printf("Successfully parsed CHARACTERS block of type %s\n", char.Format.DataType)
	}
}
```

### Generating a new file

```go
package main

import (
	"os"

	"github.com/espinosajuanma/go-nexus"
	"github.com/espinosajuanma/go-nexus/blocks/characters"
)

func main() {
	nex := nexus.New()

	// Create and attach a TAXA block
	tb := nex.NewTaxaBlock()
	tb.SetTitle("My_Taxa")
	tb.AddTaxon("fish")
	tb.AddTaxon("frog")

	// Create and attach a CHARACTERS block
	cb := nex.NewCharactersBlock(characters.Standard)
	cb.SetTitle("Morphology Matrix")
	cb.AddTaxon("fish")
	cb.AddTaxon("frog")

	// Export to stdout (or any io.Writer)
	nex.Export(os.Stdout)
}
```

## Supported Blocks

The following NEXUS blocks are currently supported and natively typed:

- `TAXA`
- `CHARACTERS`
- `TREES`

Unsupported or custom blocks are safely parsed as generic blocks and can be
retrieved using GetBlockByName or created using NewUnknownBlock. 

## Exporters

The library also supports rendering NEXUS data into other phylogenetic formats:

- `TNT`
- `NONA`