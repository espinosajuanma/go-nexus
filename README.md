# Go Nexus

A modular Go library to parse, manipulate, and generate NEXUS (`.nex`)
phylogenetic files.

## Installation

```bash
go get github.com/espinosajuanma/nexus
```

## Usage

The library provides a simple API to read existing files or build new ones from
scratch.

### Parsing an existing file

```go
package main

import (
	"fmt"
	"os"

	"https://github.com/espinosajuanma/nexus"
	// Blank imports are required to register block parsers
	_ "https://github.com/espinosajuanma/nexus/blocks/characters"
	_ "https://github.com/espinosajuanma/nexus/blocks/taxa"
	_ "https://github.com/espinosajuanma/nexus/blocks/trees"
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

	fmt.Printf("Successfully parsed %d blocks\n", len(nex.Blocks))
}
```

### Generating a new file

```go
package main

import (
	"os"

	"https://github.com/espinosajuanma/nexus"
	"https://github.com/espinosajuanma/nexus/blocks/taxa"
)

func main() {
	nex := nexus.New()

	// Create and attach a TAXA block
	tb := taxa.New(nex)
	tb.SetTitle("My_Taxa")
	tb.AddTaxon("fish")
	tb.AddTaxon("frog")

	// Export to stdout (or any io.Writer)
	nex.Export(os.Stdout)
}
```

## Supported Blocks

The following NEXUS blocks are currently supported and modularized:

- `TAXA`
- `CHARACTERS`
- `TREES`

Unsupported blocks are safely skipped during parsing.