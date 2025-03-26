package ghsummary

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <username>", os.Args[0])
	}
	username := os.Args[1] // Replace with dynamic input if needed

	// Fetch GitHub activity
	activity, err := GetUserActivity(username)
	if err != nil {
		log.Fatalf("Error fetching GitHub activity: %v", err)
	}

	// Generate summary using LLM
	summary, err := GenerateSummary(activity)
	if err != nil {
		log.Fatalf("Error generating summary: %v", err)
	}

	// Generate SVG from summary
	outputFile := "summary.svg"
	err = GenerateSVGFile(summary, outputFile)
	if err != nil {
		log.Fatalf("Error generating SVG: %v", err)
	}

	fmt.Printf("Summary SVG generated: %s\n", outputFile)
}
