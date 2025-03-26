package ghsummary

import (
	"context"
	"errors"
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
		"gemini-2.0-flash-exp",
		genai.Text(activity),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: "Generate a brief, max 10 sentence summary of github activity of the user. Go strait to the summary without explanation what it is. User pronouces are to help build sentences, use pronouces only when needed, do not specify them in parentis. You can start with \"<Username> recently...\""}}},
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
