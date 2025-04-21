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
	Pronouns            = "he/him"
	SystemPromptSummary = `Generate a brief, max 10 sentence summary of github activity of the user.
	Go straight to the summary without explanation what it is. Try to be concise and clear. User's pronouns: %s.
	User pronouns are to help build sentences, use pronouns only when needed, do not specify them in parenthesis.
	The text will be saved in svg <text> element, so do not use formatting at all. Just plain text.
	You can start with \"<Username> recently...\"`
	SystemPromptSummaryCommit = `Generate a brief, max 4 sentence summary of commit content.`
)

func GenerateSummary(activity string) (string, error) {
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
		"gemini-2.5-flash-preview-04-17",
		genai.Text(activity),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: fmt.Sprintf(SystemPromptSummary, Pronouns)}}},
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
		"gemini-2.0-flash-lite",
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
