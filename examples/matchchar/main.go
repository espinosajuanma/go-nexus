package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/nexus"
	"github.com/espinosajuanma/nexus/blocks/characters"
	"github.com/espinosajuanma/nexus/core"
)

func main() {
	file, _ := os.Open(filepath.Join("examples", "matchchar", "matchchar.nex"))
	defer file.Close()

	nex, _ := nexus.Parse(file)

	if char, ok := core.GetBlock[*characters.CharactersBlock](nex); ok {
		fmt.Println("=== Matchchar Example ===")

		chimp := char.Matrix.GetTaxon("chimp")
		char1 := char.Matrix.GetCharacterByIndex(1)

		// Chimp's first character is '.', which should have been copied from 'human' (which is '1')
		state := chimp.GetState(char1)

		if len(state.Values) > 0 && state.Values[0].Symbol == "1" {
			fmt.Println("Success: Matchchar '.' was successfully resolved to '1' from the first taxon.")
		} else {
			fmt.Println("Failed: Matchchar was not resolved correctly.")
		}

		// Exporting to see the final matrix
		fmt.Println("\n-- Exported Matrix --")
		nex.Export(os.Stdout)
	}
}
