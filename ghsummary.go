package ghsummary

import (
	"fmt"
	"log"
)

func GenerateSummarySVG(username string) (string, error) {
	// Fetch GitHub activity
	activity, err := GetUserActivity(username)
	if err != nil {
		log.Printf("Error fetching GitHub activity: %v", err)
		return "", fmt.Errorf("failed to fetch GitHub activity: %w", err)
	}

	// Generate summary using LLM
	summary, err := GenerateSummary(activity)
	if err != nil {
		log.Printf("Error generating summary: %v", err)
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	// Generate SVG content
	svgContent, err := GenerateSVG(summary, "")
	if err != nil {
		log.Printf("Error generating SVG: %v", err)
		return "", fmt.Errorf("failed to generate SVG: %w", err)
	}

	return svgContent, nil
}
