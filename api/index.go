package handler

import (
	"fmt"
	"log"
	"net/http"

	"github.com/McCzarny/ghsummary"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to SVG
	w.Header().Set("Content-Type", "image/svg+xml")

	// Extract username from query parameters
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Missing 'username' query parameter", http.StatusBadRequest)
		return
	}

	// Fetch GitHub activity
	activity, err := ghsummary.GetUserActivity(username)
	if err != nil {
		log.Printf("Error fetching GitHub activity: %v", err)
		http.Error(w, "Failed to fetch GitHub activity", http.StatusInternalServerError)
		return
	}

	// Generate summary using LLM
	summary, err := ghsummary.GenerateSummary(activity)
	if err != nil {
		log.Printf("Error generating summary: %v", err)
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	// Generate SVG content
	svgContent, err := ghsummary.GenerateSVG(summary, "")
	if err != nil {
		log.Printf("Error generating SVG: %v", err)
		http.Error(w, "Failed to generate SVG", http.StatusInternalServerError)
		return
	}

	// Write SVG content to response
	fmt.Fprint(w, svgContent)
}
