package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/McCzarny/ghsummary"
	"github.com/McCzarny/ghsummary/utils"
)

func main() {
	username := flag.String("username", "", "GitHub username")
	outputFile := flag.String("output", "summary.svg", "Output SVG file")
	maxEvents := flag.Int("max-events", 100, "Maximum number of events to fetch")
	flag.Parse()

	// Sanitize inputs
	if !utils.SanitizeInputs(*username, *outputFile) {
		log.Fatalf("Usage: %s --username <username> --output <outputFile> --max-events <maxEvents>", os.Args[0])
	}

	log.Printf("Running app with username: %s, output file: %s, max events: %d", *username, *outputFile, *maxEvents)

	// Fetch GitHub activity
	activity, err := ghsummary.GetUserActivity(*username, *maxEvents)
	if err != nil {
		log.Fatalf("Error fetching GitHub activity: %v", err)
	}

	// Generate summary using LLM
	summary, err := ghsummary.GenerateSummary(activity)
	if err != nil {
		log.Fatalf("Error generating summary: %v", err)
	}

	// Generate SVG from summary
	err = ghsummary.GenerateSVGFile(summary, *outputFile)
	if err != nil {
		log.Fatalf("Error generating SVG: %v", err)
	}

	fmt.Printf("Summary SVG generated: %s\n", *outputFile)
}
