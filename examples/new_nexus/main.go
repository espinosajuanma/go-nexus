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
)

func main() {
	fmt.Println("\n=== Creating NEXUS Structure ===")
	nex := nexus.New()

	// Create a TAXA block to set a specific title.
	tb := taxa.New(nex)
	tb.SetTitle("Database_Export")

	// Create the CHARACTERS block
	cb := characters.New(nex, characters.Standard)
	cb.SetTitle("Morphology Matrix")

	taxons := []string{"fish", "frog", "snake", "mouse"}

	// Add taxa to the block
	for _, taxon := range taxons {
		tb.AddTaxon(taxon)
		cb.AddTaxon(taxon)
	}

	// Add characters to the block
	eyeColor := cb.AddCharacter("eye color", "light red", "blue", "green")
	tailLength := cb.AddCharacter("tail length", "short", "long")
	unnamedChar := cb.AddCharacter("", "absent", "present")

	charsList := []*characters.Character{eyeColor, tailLength, unnamedChar}

	for _, char := range charsList {
		for _, taxon := range cb.Taxa {
			randState := char.StateLabels[rand.Intn(len(char.StateLabels))]
			taxon.SetState(char, randState)
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

	fmt.Println("\n=== Exporting NEXUS Structure ===")
	err := nex.Export(os.Stdout)
	if err != nil {
		log.Fatalf("Failed to export NEXUS file: %v", err)
	}
}
