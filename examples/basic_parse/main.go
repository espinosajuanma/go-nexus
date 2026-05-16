package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/nexus"
)

func main() {
	fileName := filepath.Join("examples", "basic_parse", "test.nex")

	fmt.Printf("=== Reading file: %v", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open NEXUS file: %v", err)
	}
	defer file.Close()

	fmt.Printf("=== Parsing NEXUS Data from %s ===\n", fileName)
	nex, err := nexus.Parse(file)
	if err != nil {
		log.Fatalf("Failed to parse NEXUS file: %v", err)
	}

	if taxa, ok := nex.GetTaxaBlock(); ok {
		fmt.Println("-- Found a TAXA Block --")
		fmt.Printf("Taxa Count: %d\n", taxa.Dimensions)
		fmt.Printf("Taxa Labels: %v\n", taxa.TaxLabels)
	} else {
		fmt.Println("-- No TAXA Block found --")
	}

	if char, ok := nex.GetCharactersBlock(); ok {
		char.AddTaxon("Sarasa 1")
		fmt.Println("-- Found a CHARACTERS Block --")
		fmt.Printf("Characters Count: %d\n", char.Dimensions)
		fmt.Printf("Data Type: %s\n", char.Format.DataType)
	} else {
		fmt.Println("-- No CHARACTERS Block found --")
	}

	if trees, ok := nex.GetTreesBlock(); ok {
		fmt.Println("-- Found a TREES Block --")
		fmt.Printf("Trees Count: %d\n", len(trees.Trees))
	} else {
		fmt.Println("-- No TREES Block found --")
	}

	fmt.Println("\n=== Exporting NEXUS Data ===")
	err = nex.Export(os.Stdout)
	if err != nil {
		log.Fatalf("Failed to export NEXUS file: %v", err)
	}

}
