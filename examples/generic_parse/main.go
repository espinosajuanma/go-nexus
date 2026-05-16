package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/go-nexus"
)

func main() {
	fileName := filepath.Join("examples", "generic_parse", "test.nex")

	fmt.Printf("=== Reading file: %v\n", fileName)
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

	fmt.Println("\n=== Exporting NEXUS Data ===")
	err = nex.Export(os.Stdout)
	if err != nil {
		log.Fatalf("Failed to export NEXUS file: %v", err)
	}
}
