package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"
)

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	model := "gemini-2.0-flash"
	prompt := "Hello, are you working?"

	fmt.Printf("Sending prompt to %s: %s\n", model, prompt)

	resp, err := client.Models.GenerateContent(ctx, model, []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: prompt}},
		},
	}, nil)

	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		fmt.Printf("Response: %s\n", resp.Candidates[0].Content.Parts[0].Text)
		fmt.Println("\nSUCCESS: API key is working!")
	} else {
		fmt.Println("No response content received.")
	}
}
