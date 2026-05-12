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
	fileName := filepath.Join("examples", "export_nona", "test.nex")

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Failed to open NEXUS file: %v", err)
	}
	defer file.Close()

	nex, err := nexus.Parse(file)
	if err != nil {
		log.Fatalf("Failed to parse NEXUS file: %v", err)
	}

	nona := xread.New(nex, xread.NONA)
	nona.SetAuthor("Juanma Espinosa")
	nona.SetProject("Go Nexus Export Example")

	str, err := nona.Render()
	if err != nil {
		log.Fatalf("Failed to export to NONA format: %v", err)
	}
	fmt.Print(str)

}
