package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/espinosajuanma/go-nexus"
)

func main() {
	// A sample NEXUS string with a rich TAXA block
	nexusData := `#NEXUS
BEGIN TAXA;
	TITLE 'Amphibians and Reptiles';
	DIMENSIONS NTAX=4;
	TAXLABELS frog toad snake lizard;
	TAXSET amphibians = frog toad;
	TAXSET reptiles = snake lizard;
	TAXPARTITION clades = clade_a: frog toad, clade_b: snake lizard;
END;`

	fmt.Println("=== Parsing NEXUS Data ===")
	nex, err := nexus.Parse(strings.NewReader(nexusData))
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
