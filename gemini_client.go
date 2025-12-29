package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/genai"
)

// GeminiClient defines the interface for AI decision making
type GeminiClient interface {
	DecideNextAction(ctx context.Context, req GeminiDecisionRequest) (GeminiDecisionResponse, error)
}

// GeminiService wraps the Google Gen AI client
type GeminiService struct {
	client *genai.Client
	model  string
}

// NewGeminiService creates a new Gemini service
func NewGeminiService(client *genai.Client) *GeminiService {
	return &GeminiService{
		client: client,
		model:  "gemini-2.0-flash",
	}
}

// DecideNextAction asks Gemini to decide the next action for an agent
func (s *GeminiService) DecideNextAction(ctx context.Context, req GeminiDecisionRequest) (GeminiDecisionResponse, error) {
	const maxRetries = 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Create timeout context
		timeoutCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		// Build the prompt
		systemPrompt := s.buildSystemPrompt()
		userPrompt := s.buildUserPrompt(req)

		// Log request size
		requestSize := len(systemPrompt) + len(userPrompt)
		log.Printf("[Gemini] Agent request size: %d bytes (attempt %d)", requestSize, attempt+1)

		// Call Gemini API
		resp, err := s.client.Models.GenerateContent(timeoutCtx, s.model, []*genai.Content{
			{
				Role:  "user",
				Parts: []*genai.Part{{Text: userPrompt}},
			},
		}, &genai.GenerateContentOptions{
			SystemInstruction: &genai.Content{
				Role:  "system",
				Parts: []*genai.Part{{Text: systemPrompt}},
			},
			Temperature:     genai.Ptr[float32](0.7),
			MaxOutputTokens: genai.Ptr[int32](1024),
		})

		if err != nil {
			log.Printf("[Gemini] Request failed: %v", err)
			lastErr = fmt.Errorf("gemini request failed: %w", err)
			continue
		}

		// Extract response text
		if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
			lastErr = fmt.Errorf("empty response from gemini")
			continue
		}

		responseText := resp.Candidates[0].Content.Parts[0].Text
		log.Printf("[Gemini] Raw response size: %d bytes", len(responseText))

		// Clean response - remove markdown code blocks if present
		responseText = strings.TrimSpace(responseText)
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
		responseText = strings.TrimSpace(responseText)

		// Parse JSON
		var decision GeminiDecisionResponse
		if err := json.Unmarshal([]byte(responseText), &decision); err != nil {
			log.Printf("[Gemini] JSON parse error on attempt %d: %v, response was: %s", attempt+1, err, responseText)
			lastErr = fmt.Errorf("invalid json response: %w", err)
			continue
		}

		// Validate response
		if err := s.validateResponse(decision); err != nil {
			log.Printf("[Gemini] Validation error on attempt %d: %v", attempt+1, err)
			lastErr = err
			continue
		}

		log.Printf("[Gemini] Successful decision: action=%s, selector=%s", decision.Action, decision.Selector)
		return decision, nil
	}

	return GeminiDecisionResponse{}, fmt.Errorf("failed after %d attempts: %w", maxRetries, lastErr)
}

// buildSystemPrompt creates the system instruction for Gemini
func (s *GeminiService) buildSystemPrompt() string {
	return `You are an autonomous web testing agent. Your goal is to navigate websites and achieve specific objectives.

CRITICAL OUTPUT REQUIREMENTS:
1. You MUST respond with valid JSON ONLY
2. Do NOT include markdown formatting (no \`\`\`json or \`\`\`)
3. Do NOT include explanatory text outside the JSON
4. Your entire response must be a single valid JSON object

Available actions:
- click: Click on an interactive element (button, link, etc.)
- type: Fill in an input field and submit the nearest form
- wait: Wait and observe (use when page needs time to load)
- go_back: Go back to the previous page

Your response must follow this exact JSON structure:
{
  "reasoning": "Brief explanation of why you chose this action",
  "action": "click|type|wait|go_back",
  "selector": "CSS selector of the element (for click/type)",
  "text_input": "text to enter (for type action only)",
  "expected_next_state": "what you expect to happen"
}

Choose actions that efficiently accomplish the mission goal. Be strategic and deliberate.`
}

// buildUserPrompt creates the user prompt with mission context
func (s *GeminiService) buildUserPrompt(req GeminiDecisionRequest) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("MISSION GOAL: %s\n\n", req.MissionGoal))
	sb.WriteString(fmt.Sprintf("CURRENT PAGE: %s\n", req.CurrentPage.URL))
	sb.WriteString(fmt.Sprintf("TITLE: %s\n", req.CurrentPage.Title))

	if req.CurrentPage.Description != "" {
		sb.WriteString(fmt.Sprintf("DESCRIPTION: %s\n", req.CurrentPage.Description))
	}

	sb.WriteString("\nINTERACTIVE ELEMENTS:\n")
	for i, elem := range req.CurrentPage.InteractiveElements {
		sb.WriteString(fmt.Sprintf("  %d. [%s] %s - selector: %s\n",
			i+1, elem.Type, elem.Text, elem.Selector))
		if elem.Href != "" {
			sb.WriteString(fmt.Sprintf("     href: %s\n", elem.Href))
		}
		if elem.Placeholder != "" {
			sb.WriteString(fmt.Sprintf("     placeholder: %s\n", elem.Placeholder))
		}
	}

	if len(req.ActionHistory) > 0 {
		sb.WriteString("\nRECENT ACTIONS:\n")
		start := len(req.ActionHistory) - 10
		if start < 0 {
			start = 0
		}
		for i, action := range req.ActionHistory[start:] {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, action))
		}
	}

	sb.WriteString("\nDecide the next action to accomplish the mission goal.")
	sb.WriteString("\nRespond with JSON ONLY (no markdown formatting).")

	return sb.String()
}

// validateResponse validates the Gemini response
func (s *GeminiService) validateResponse(resp GeminiDecisionResponse) error {
	validActions := map[string]bool{
		"click":   true,
		"type":    true,
		"wait":    true,
		"go_back": true,
	}

	if !validActions[resp.Action] {
		return fmt.Errorf("invalid action: %s", resp.Action)
	}

	if resp.Action == "click" || resp.Action == "type" {
		if resp.Selector == "" {
			return fmt.Errorf("selector required for action: %s", resp.Action)
		}
	}

	if resp.Action == "type" && resp.TextInput == "" {
		return fmt.Errorf("text_input required for type action")
	}

	return nil
}

const requestTimeout = 5 * time.Second
