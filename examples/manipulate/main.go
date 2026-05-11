package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/nexus"
	"github.com/espinosajuanma/nexus/blocks/characters"
	"github.com/espinosajuanma/nexus/core"
)

func main() {
	file, _ := os.Open(filepath.Join("examples", "manipulate", "manipulate.nex"))
	defer file.Close()

	nex, _ := nexus.Parse(file)

	if char, ok := core.GetBlock[*characters.CharactersBlock](nex); ok {
		fmt.Println("=== Manipulation Example ===")

		bat := char.Matrix.GetTaxon("bat")
		char2 := char.Matrix.GetCharacterByIndex(2)

		currentState := bat.GetState(char2)
		fmt.Printf("Original State Type: %s\n", currentState.Type) // Should print SINGLE

		// Change bat's 2nd character to a polymorphic state (0 and 1)
		err := bat.SetState(char2, "0", "1")
		if err != nil {
			log.Fatal(err)
		}

		newState := bat.GetState(char2)
		fmt.Printf("New State Type: %s\n", newState.Type) // Should print POLYMORPHIC

		// Let's print the internal observations to verify
		fmt.Print("Internal Observations: ")
		for _, obs := range newState.Observations {
			fmt.Printf("%s ", obs.Symbol)
		}
		fmt.Println()

		// Exporting to see the final matrix
		fmt.Println("\n-- Exported Matrix --")
		nex.Export(os.Stdout) // You will see `bat 1(0 1)1` in the output
	}
}
