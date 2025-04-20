package ghsummary

import (
	"log"
)

func GenerateSummarySVG(username string, max_events int, mode string) (string, bool) {
	// Fetch GitHub activity
	activity, err := GetUserActivity(username, max_events, mode)
	if err != nil {
		log.Printf("Error fetching GitHub activity: %v", err)
		return "", false
	}

	// Generate summary using LLM
	summary, err := GenerateSummary(activity)
	if err != nil {
		log.Printf("Error generating summary: %v", err)
		return "", false
	}

	// Generate SVG content
	svgContent, err := GenerateSVG(summary, "")
	if err != nil {
		log.Printf("Error generating SVG: %v", err)
		return "", false
	}

	return svgContent, true
}
