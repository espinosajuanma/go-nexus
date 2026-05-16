package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/nexus"
)

func main() {
	fileName := filepath.Join("examples", "taxa_vector", "test.nex")

	fmt.Printf("=== Reading file: %v", fileName)
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open NEXUS file: %v", err)
	}
	defer file.Close()

	fmt.Println("=== Parsing NEXUS Data ===")
	nex, err := nexus.Parse(file)
	if err != nil {
		log.Fatalf("Failed to parse NEXUS file: %v", err)
	}

	// Retrieve the TAXA block
	if taxaBlock, ok := nex.GetTaxaBlock(); ok {
		fmt.Println("-- Found a TAXA Block --")
		fmt.Printf("Title: %s\n", taxaBlock.Title)
		fmt.Printf("Taxa Count (NTAX): %d\n", taxaBlock.Dimensions)
		fmt.Printf("Taxa Labels: %v\n", taxaBlock.TaxLabels)

		fmt.Println("\n-- Taxa Sets --")
		for name, set := range taxaBlock.TaxSets {
			fmt.Printf("  Set '%s' (Format: %s): %v\n", name, set.Format, set.TaxaList)
		}

		fmt.Println("\n-- Taxa Partitions --")
		for name, partition := range taxaBlock.TaxPartitions {
			fmt.Printf("  Partition '%s' (Format: %s):\n", name, partition.Format)
			for subsetName, subsetTaxa := range partition.Subsets {
				fmt.Printf("    Subset '%s': %v\n", subsetName, subsetTaxa)
			}
		}

		fmt.Println("\n=== Exporting NEXUS Data ===")
		nex.Export(os.Stdout)
	} else {
		fmt.Println("-- No TAXA Block found --")
	}
}
