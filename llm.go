package ghsummary

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

type LLMRequest struct {
	Prompt string `json:"prompt"`
}

type LLMResponse struct {
	Summary string `json:"summary"`
}

type GeminiRequest struct {
	Prompt string `json:"prompt"`
}

type GeminiResponse struct {
	Summary string `json:"summary"`
}

// Constants for Gemini API
const (
	SystemPromptSummary = `Generate a concise summary (max 10 sentences) of the user's recent GitHub activity based on the provided data.
You can start the summary directly with "<Username> recently...".
Use the user's pronouns (%s) naturally only when needed for sentence structure; do not state the pronouns themselves.
The output must be plain text only, with absolutely no formatting (no markdown, newlines, etc.), suitable for direct use within an SVG <text> element.
Focus on key actions like commits, pull requests, and issues. Avoid any introductory or explanatory text.`
	SystemPromptSummaryCommit = `Generate a brief, max 4 sentence summary of commit content.`
)

func GenerateSummary(activity string, pronouns ...string) (string, error) {
	return GenerateSummaryWithRetry(activity, 0, pronouns...)
}

func GenerateSummaryWithRetry(activity string, attempt int, pronouns ...string) (string, error) {
	const maxRetries = 5
	// If pronouns are provided, use them; otherwise default to "he/him"
	pronounValue := "he/him"
	if len(pronouns) > 0 && pronouns[0] != "" {
		pronounValue = pronouns[0]
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", errors.New("GEMINI_API_KEY is not set in the environment")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Client created. Generating summary... (attempt %d)", attempt+1)
	result, err := client.Models.GenerateContent(ctx,
		"gemini-2.5-flash",
		genai.Text(activity),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: fmt.Sprintf(SystemPromptSummary, pronounValue)}}},
		},
	)
	log.Printf("Summary generation completed.")

	if err != nil {
		// Check for server overload (503) or rate limit errors
		errMsg := err.Error()
		isOverloaded := strings.Contains(errMsg, "503") || strings.Contains(errMsg, "overloaded")
		isRateLimit := strings.Contains(errMsg, "PerMinute")

		if (isOverloaded || isRateLimit) && attempt < maxRetries {
			// Exponential backoff: 2^attempt seconds (32s, 64s, 128s, 256s, 512s)
			waitDuration := time.Duration(1<<uint(attempt)) * 32 * time.Second
			if isRateLimit {
				// For rate limits, wait at least 1 minute
				waitDuration = time.Minute
			}
			log.Printf("Error: %s. Retrying attempt %d/%d after %v", errMsg, attempt+1, maxRetries, waitDuration)
			time.Sleep(waitDuration)
			return GenerateSummaryWithRetry(activity, attempt+1, pronouns...)
		}
		log.Fatal(err)
	}

	summary := ""
	for _, part := range result.Candidates[0].Content.Parts {
		summary += part.Text
	}
	if len(summary) == 0 {
		return "", errors.New("no summary generated")
	}
	// Print the summary
	log.Printf("Summary: %s", summary)

	return summary, nil
}

func GenerateCommitSummary(content string, attempt int) (string, error) {
	const maxRetries = 5
	apiKey := os.Getenv("GEMINI_API_KEY")

	log.Printf("Generating commit summary... (attempt %d)", attempt+1)
	if apiKey == "" {
		return "", errors.New("GEMINI_API_KEY is not set in the environment")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(ctx,
		"gemini-2.5-flash-lite",
		genai.Text(content),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: SystemPromptSummaryCommit}}},
		},
	)
	log.Printf("Commit summary generation completed.")

	if err != nil {
		errMsg := err.Error()
		isOverloaded := strings.Contains(errMsg, "503") || strings.Contains(errMsg, "overloaded")
		isRateLimit := strings.Contains(errMsg, "PerMinute")

		if (isOverloaded || isRateLimit) && attempt < maxRetries {
			// Exponential backoff: 2^attempt seconds (2s, 4s, 8s, 16s, 32s)
			waitDuration := time.Duration(1<<uint(attempt)) * 2 * time.Second
			if isRateLimit {
				// For rate limits, wait at least 1 minute
				waitDuration = time.Minute
			}
			log.Printf("Error: %s. Retrying attempt %d/%d after %v", errMsg, attempt+1, maxRetries, waitDuration)
			time.Sleep(waitDuration)
			return GenerateCommitSummary(content, attempt+1)
		}
		log.Fatal(err)
	}

	summary := ""
	for _, part := range result.Candidates[0].Content.Parts {
		summary += part.Text
	}
	if len(summary) == 0 {
		return "", errors.New("no commit summary generated")
	}
	// Print the summary
	log.Printf("Commit Summary: %s", summary)

	return summary, nil
}
