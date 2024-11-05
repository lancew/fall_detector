package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/yourusername/gpx-analyzer/analyzer"
)

func main() {
	flag.Parse()
	if flag.NArg() != 1 {
		fmt.Println("Usage: gpx-analyzer <gpx-file>")
		os.Exit(1)
	}

	gpxFile := flag.Arg(0)
	if filepath.Ext(gpxFile) != ".gpx" {
		fmt.Println("Error: File must have .gpx extension")
		os.Exit(1)
	}

	falls, err := analyzer.AnalyzeGPXFile(gpxFile)
	if err != nil {
		log.Fatalf("Error analyzing GPX file: %v", err)
	}

	if len(falls) == 0 {
		fmt.Println("No falls detected in the activity")
		return
	}

	fmt.Printf("Detected %d potential falls:\n", len(falls))
	for i, fall := range falls {
		fmt.Printf("%d. Fall detected at position (%.6f, %.6f) at time %v\n",
			i+1, fall.Latitude, fall.Longitude, fall.Timestamp)
	}
}
