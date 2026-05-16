package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/go-nexus"
)

func main() {
	file, _ := os.Open(filepath.Join("examples", "interleave", "interleave.nex"))
	defer file.Close()

	nex, _ := nexus.Parse(file)

	if char, ok := nex.GetCharactersBlock(); ok {
		fmt.Println("=== Interleave Example ===")
		fmt.Printf("Taxon amount in matrix: %d\n", len(char.Matrix.Taxa))
		fmt.Printf("Character amount in matrix: %d\n", len(char.Matrix.Characters))

		taxon1 := char.Matrix.GetTaxon("taxon1")

		// Let's verify that characters from both chunks are seamlessly stored in taxon1's row
		char1 := char.Matrix.GetCharacterByIndex(1) // From Chunk 1 (A)
		char6 := char.Matrix.GetCharacterByIndex(6) // From Chunk 2 (T)

		fmt.Printf("Taxon1 Char 1: %s\n", taxon1.GetState(char1).Values[0].Symbol)
		fmt.Printf("Taxon1 Char 6: %s\n", taxon1.GetState(char6).Values[0].Symbol)

		fmt.Println("\n-- Exported Matrix (Should be interleaved) --")

		// The export will respect the interleave size, so we set it to 3 for demonstration purposes
		char.Format.InterleaveSize = 3
		nex.Export(os.Stdout)
	}
}
