package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/espinosajuanma/nexus"
	"github.com/espinosajuanma/nexus/blocks/characters"
	"github.com/espinosajuanma/nexus/blocks/taxa"
	"github.com/espinosajuanma/nexus/blocks/trees"
	"github.com/espinosajuanma/nexus/core"
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

	if t, ok := core.GetBlock[*taxa.TaxaBlock](nex); ok {
		t.AddTaxon("Sarasa_1")
		fmt.Println("-- Found a TAXA Block --")
		fmt.Printf("Taxa Count: %d\n", t.Dimensions.Count)
		fmt.Printf("Taxa Labels: %v\n", t.TaxLabels)
	} else {
		fmt.Println("-- No TAXA Block found --")
	}

	if c, ok := core.GetBlock[*characters.CharactersBlock](nex); ok {
		fmt.Println("-- Found a CHARACTERS Block --")
		fmt.Printf("Characters Count: %d\n", c.Dimensions.NChar)
		fmt.Printf("Data Type: %s\n", c.Format.DataType)
	} else {
		fmt.Println("-- No CHARACTERS Block found --")
	}

	if tr, ok := core.GetBlock[*trees.TreesBlock](nex); ok {
		fmt.Println("-- Found a TREES Block --")
		fmt.Printf("Trees Count: %d\n", len(tr.Trees))
	} else {
		fmt.Println("-- No TREES Block found --")
	}

	return nex, nil
}

func newNexus() (*nexus.Nexus, error) {
	nex := nexus.New()

	// Create a TAXA block to set a specific title.
	tb := taxa.New(nex)
	tb.SetTitle("Database_Export")

	// Create the CHARACTERS block
	cb := characters.New(nex, characters.Standard)
	cb.SetTitle("Morphology_Matrix")

	// Add taxa to the block
	cb.AddTaxon("fish fish")
	cb.AddTaxon("frog")
	cb.AddTaxon("snake")
	cb.AddTaxon("mouse")

	// Add characters to the block
	eyeColor := cb.AddCharacter("eye color", "light red", "blue", "green")
	tailLength := cb.AddCharacter("tail length", "short", "long")
	unnamedChar := cb.AddCharacter("", "absent", "present")

	charsList := []*characters.Character{eyeColor, tailLength, unnamedChar}

	for _, char := range charsList {
		for _, taxon := range cb.Taxa {
			randState := char.StateLabels[rand.Intn(len(char.StateLabels))]
			taxon.AddCharacterState(char, randState)
		}
	}

	// Create the TREES block
	tr := trees.New(nex)

	// Map tokens to taxa (This AUTO-CREATES the TAXA block if not present!)
	tr.AddTranslate("1", "fish")
	tr.AddTranslate("2", "frog")
	tr.AddTranslate("3", "snake")
	tr.AddTranslate("4", "mouse")

	// Build the Tree AST programmatically using the fluent builder
	// We are building this topology: [&R] (1:0.5, (2:0.2, (3:0.1, 4:0.1):0.3):0.4)
	rootNode := trees.NewNode().AddComment("[&R]").
		AddChild(trees.NewNode().SetName("1").SetBranchLength("0.5")).
		AddChild(trees.NewNode().SetBranchLength("0.4").
			AddChild(trees.NewNode().SetName("2").SetBranchLength("0.2")).
			AddChild(trees.NewNode().SetBranchLength("0.3").
				AddChild(trees.NewNode().SetName("3").SetBranchLength("0.1")).
				AddChild(trees.NewNode().SetName("4").SetBranchLength("0.1"))))

	// Attach the parsed root node to the block
	tr.AddTree("my_database_tree", true, rootNode)

	return nex, nil
}