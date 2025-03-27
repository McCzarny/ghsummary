package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/McCzarny/ghsummary"
)

func main() {
	username := flag.String("username", "", "GitHub username")
	outputFile := flag.String("output", "summary.svg", "Output SVG file")
	flag.Parse()

	if *username == "" {
		log.Fatalf("Usage: %s --username <username> --output <outputFile>", os.Args[0])
	}

	// Fetch GitHub activity
	activity, err := ghsummary.GetUserActivity(*username)
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
