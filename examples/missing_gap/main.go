package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/go-nexus"
	"github.com/espinosajuanma/go-nexus/blocks/characters"
)

func main() {
	file, _ := os.Open(filepath.Join("examples", "missing_gap", "missing_gap.nex"))
	defer file.Close()

	nex, _ := nexus.Parse(file)

	if char, ok := nex.GetCharactersBlock(); ok {
		fmt.Println("=== Missing & Gap Example ===")

		fish := char.Matrix.GetTaxon("fish")
		if fish == nil {
			fmt.Println("Error: 'fish' taxon not found.")
			return
		}
		frog := char.Matrix.GetTaxon("frog")
		char1 := char.Matrix.GetCharacterByIndex(1)
		char3 := char.Matrix.GetCharacterByIndex(3)

		// Check Missing
		stateFish1 := fish.GetState(char1)
		if stateFish1.Type == characters.StateMissing {
			fmt.Println("Success: Parser correctly flagged 'fish' character 1 as Missing.")
		}

		// Check Gap
		stateFrog1 := frog.GetState(char1)
		if stateFrog1.Type == characters.StateGap {
			fmt.Println("Success: Parser correctly flagged 'frog' character 1 as Gap.")
		}

		// Manipulate to Gap/Missing
		// Set fish character 3 (currently Gap) to Missing
		fish.SetState(char3, characters.StateSingle, "0")
		if fish.GetState(char3).Type == characters.StateMissing {
			fmt.Println("Success: Programmatically changed 'fish' character 3 to Missing.")
		}
	}

	nex.Export(os.Stdout)
}
