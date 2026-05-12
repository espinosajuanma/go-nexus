package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/espinosajuanma/nexus"
	"github.com/espinosajuanma/nexus/exports/xread"
)

func main() {
	fileName := filepath.Join("examples", "export_tnt", "test.nex")

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open NEXUS file: %v", err)
	}
	defer file.Close()

	nex, err := nexus.Parse(file)
	if err != nil {
		log.Fatalf("Failed to parse NEXUS file: %v", err)
	}

	tnt := xread.New(nex, xread.TNT).
		SetProject("Example Project").
		SetAuthor("Juanma Espinosa")

	output, err := tnt.Render()
	if err != nil {
		log.Fatalf("Failed to render TNT: %v", err)
	}
	fmt.Println(output)
}
