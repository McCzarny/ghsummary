package ghsummary

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

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

	log.Printf("Client created. Generating summary...")
	result, err := client.Models.GenerateContent(ctx,
		"gemini-2.5-flash",
		genai.Text(activity),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: fmt.Sprintf(SystemPromptSummary, pronounValue)}}},
		},
	)
	log.Printf("Summary generation completed.")

	if err != nil {
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

func GenerateCommitSummary(content string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")

	log.Printf("Generating commit summary...")
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
