package main

import (
	"fmt"
	"log"
	"os"

	"github.com/espinosajuanma/nexus"
)

func main() {
	fmt.Println("=== 1. Parsing NEXUS Data from ./test.nex ===")

	// Open the local file
	file, err := os.Open("test.nex")
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	fmt.Println("=== 1. Parsing NEXUS Data ===")

	// Parse the file using our custom package
	nex, err := nexus.Parse(file)
	if err != nil {
		log.Fatalf("Failed to parse NEXUS file: %v", err)
	}
	fmt.Println("Successfully parsed the NEXUS file into Go structs!")

	if taxa, ok := nexus.GetBlock[*nexus.TaxaBlock](nex); ok {
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

	fmt.Println("\n=== 2. Exporting NEXUS Data ===")
	err = nex.Export(os.Stdout)
	if err != nil {
		log.Fatalf("Failed to export NEXUS file: %v", err)
	}
}
