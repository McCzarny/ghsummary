package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/McCzarny/ghsummary"
)

func SanitizeInputs(username string, outputFile string) bool {
	// Sanitize username for length
	if len(username) > 39 {
		log.Printf("Maximum length of username is 39: %s", username)
		return false
	}

	if outputFile == "" {
		log.Printf("Output file is empty")
		return false
	}

	// Validate allowed characters in username
	for _, char := range username {
		if !(char >= 'a' && char <= 'z') && !(char >= 'A' && char <= 'Z') && !(char >= '0' && char <= '9') && char != '-' {
			log.Printf("Invalid character in username: %c", char)
			return false
		}
	}

	// Allow only relative paths for output file and do not allow paths with ".."
	if outputFile == "/" || outputFile == "." || outputFile == ".." {
		log.Printf("Invalid output file path: %s", outputFile)
		return false
	}
	if outputFile[0] == '/' {
		log.Printf("Absolute path is not allowed: %s", outputFile)
		return false
	}
	// Check if any part of the path contains ".."
	pattern := regexp.MustCompile(`(\.\./|\.\.\\|%2e%2e%2f|%2e%2e%5c|%252e%252e%255c|%c0%2e|%c0%af|%00)`)
	if pattern.MatchString(outputFile) {
		log.Printf("Path traversal detected in output file: %s", outputFile)
		return false
	}

	return true
}

func main() {
	username := flag.String("username", "", "GitHub username")
	outputFile := flag.String("output", "summary.svg", "Output SVG file")
	flag.Parse()

	// Sanitize inputs
	if !SanitizeInputs(*username, *outputFile) {
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
