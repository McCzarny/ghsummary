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
   flagSet := flag.NewFlagSet("args", flag.ExitOnError)
   username := flagSet.String("username", "", "GitHub username")
   outputFile := flagSet.String("output", "summary.svg", "Output SVG file")
   maxEvents := flagSet.Int("max-events", 100, "Maximum number of events to fetch")
   mode := flagSet.String("mode", "fast", "Mode of operation (fast, strict)")
   pronouns := flagSet.String("pronouns", "he/him", "Pronouns to use for the user (e.g. he/him, she/her, they/them)")
   flagSet.Parse(os.Args[1:])

	// Sanitize inputs
	if !utils.SanitizeInputs(*username, *outputFile) {
		log.Fatalf("Usage: %s --username <username> --output <outputFile> --max-events <maxEvents>", os.Args[0])
	}

   log.Printf("Running app with username: %s, output file: %s, max events: %d, pronouns: %s", *username, *outputFile, *maxEvents, *pronouns)

	// Fetch GitHub activity
	activity, err := ghsummary.GetUserActivity(*username, *maxEvents, *mode)
	if err != nil {
		log.Fatalf("Error fetching GitHub activity: %v", err)
	}

   // Generate summary using LLM
   summary, err := ghsummary.GenerateSummary(activity, *pronouns)
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
