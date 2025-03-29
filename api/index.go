package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/McCzarny/ghsummary"
	"github.com/McCzarny/ghsummary/utils"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// Set the content type to SVG
	w.Header().Set("Content-Type", "image/svg+xml")

	// Extract username from query parameters
	username := r.URL.Query().Get("username")
	max_events_str := r.URL.Query().Get("max-events")
	if max_events_str == "" {
		max_events_str = "100"
	}
	max_events, err := strconv.Atoi(max_events_str)
	if err != nil {
		log.Printf("Error converting max-events to integer: %v", err)
		http.Error(w, "Invalid 'max-events' query parameter", http.StatusBadRequest)
		return
	}

	if username == "" {
		http.Error(w, "Missing 'username' query parameter", http.StatusBadRequest)
		return
	}

	if !utils.SanitizeUsername(username) {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	// Fetch GitHub activity
	activity, err := ghsummary.GetUserActivity(username, max_events)
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
