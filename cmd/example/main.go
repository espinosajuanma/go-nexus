package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/espinosajuanma/nexus"
)

func main() {
	fileName := "./test.nex"
	fmt.Printf("=== 1. Parsing NEXUS Data from %s ===\n", fileName)
	existingNexus, err := parseNexus(fileName)
	if err != nil {
		log.Fatalf("Failed to parse NEXUS file: %v", err)
	}
	fmt.Println("\n=== 2. Exporting NEXUS Data ===")
	err = existingNexus.Export(os.Stdout)
	if err != nil {
		log.Fatalf("Failed to export NEXUS file: %v", err)
	}

	fmt.Println("\n=== 3. Creating NEXUS Structure ===")
	newNexus, err := newNexus()
	if err != nil {
		log.Fatalf("Failed to create NEXUS file: %v", err)
	}
	fmt.Println("\n=== 4. Exporting NEXUS Structure ===")
	err = newNexus.Export(os.Stdout)
	if err != nil {
		log.Fatalf("Failed to export NEXUS file: %v", err)
	}
}

func parseNexus(fileName string) (*nexus.Nexus, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Parse the file using our custom package
	nex, err := nexus.Parse(file)
	if err != nil {
		log.Fatalf("Failed to parse NEXUS file: %v", err)
	}
	fmt.Println("Successfully parsed the NEXUS file into Go structs!")

	if taxa, ok := nexus.GetBlock[*nexus.TaxaBlock](nex); ok {
		taxa.AddTaxon("Sarasa_1")

		fmt.Println("-- Found a TAXA Block --")
		fmt.Printf("Taxa Count: %d\n", taxa.Dimensions.Count)
		fmt.Printf("Taxa Labels: %v\n", taxa.TaxLabels)
	} else {
		fmt.Println("-- No TAXA Block found --")
	}

	if chars, ok := nexus.GetBlock[*nexus.CharactersBlock](nex); ok {
		fmt.Println("-- Found a CHARACTERS Block --")
		fmt.Printf("Characters Count: %d\n", chars.Dimensions.NChar)
		fmt.Printf("Data Type: %s\n", chars.Format.DataType)
	} else {
		fmt.Println("-- No CHARACTERS Block found --")
	}

	if trees, ok := nexus.GetBlock[*nexus.TreesBlock](nex); ok {
		fmt.Println("-- Found a TREES Block --")
		fmt.Printf("Trees Count: %d\n", len(trees.Trees))
	} else {
		fmt.Println("-- No TREES Block found --")
	}
	return nex, nil
}

func newNexus() (*nexus.Nexus, error) {
	nex := nexus.New()

	// Create a TAXA block to set a specific title.
	taxa := nex.NewTaxaBlock()
	taxa.SetTitle("Database_Export")

	// Create the CHARACTERS block
	chars := nex.NewCharactersBlock(nexus.Standard)
	chars.SetTitle("Morphology_Matrix")

	// Add taxa to the block
	chars.AddTaxon("fish fish")
	chars.AddTaxon("frog")
	chars.AddTaxon("snake")
	chars.AddTaxon("mouse")

	// Add characters to the block
	eyeColor := chars.AddCharacter("eye color", "light red", "blue", "green")
	tailLength := chars.AddCharacter("tail length", "short", "long")
	unnamedChar := chars.AddCharacter("", "absent", "present")

	characters := []*nexus.Character{eyeColor, tailLength, unnamedChar}
	for _, char := range characters {
		for _, taxon := range chars.Taxa {
			randState := char.StateLabels[rand.Intn(len(char.StateLabels))]
			taxon.AddCharacterState(char, randState)
		}
	}

	// Create the TREES block
	trees := nex.NewTreesBlock()

	// Map tokens to taxa (This AUTO-CREATES the TAXA block!)
	trees.AddTranslate("1", "fish")
	trees.AddTranslate("2", "frog")
	trees.AddTranslate("3", "snake")
	trees.AddTranslate("4", "mouse")

	// Build the Tree AST programmatically using the fluent builder
	// We are building this topology: [&R] (1:0.5, (2:0.2, (3:0.1, 4:0.1):0.3):0.4)
	rootNode := nexus.NewNode().AddComment("[&R]").
		AddChild(nexus.NewNode().SetName("1").SetBranchLength("0.5")).
		AddChild(nexus.NewNode().SetBranchLength("0.4").
			AddChild(nexus.NewNode().SetName("2").SetBranchLength("0.2")).
			AddChild(nexus.NewNode().SetBranchLength("0.3").
				AddChild(nexus.NewNode().SetName("3").SetBranchLength("0.1")).
				AddChild(nexus.NewNode().SetName("4").SetBranchLength("0.1")),
			),
		)

	// 4. Attach the parsed root node to the block
	trees.AddTree("my_database_tree", true, rootNode)

	return nex, nil
}
