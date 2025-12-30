package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/genai"
	"swarmtest/internal/models"
)

// GeminiClient provides AI-powered decision making for agents
type GeminiClient interface {
	DecideNextAction(ctx context.Context, mission *models.Mission, agent *models.Agent, page *models.StrippedPage) (*models.GeminiDecisionResponse, error)
}

// GeminiService implements GeminiClient
type GeminiService struct {
	client *genai.Client
}

func NewGeminiService(client *genai.Client) *GeminiService {
	return &GeminiService{client: client}
}

func (s *GeminiService) DecideNextAction(ctx context.Context, mission *models.Mission, agent *models.Agent, page *models.StrippedPage) (*models.GeminiDecisionResponse, error) {
	// Construct prompt
	prompt := buildPrompt(mission, agent, page)

	// Call Gemini
	// Fix type mismatches: val creates *T. GenAI expects specific types.
	// Looking at error: cannot use val(0.2) (type *float64) as *float32.
	// So Temperature is *float32.
	// MaxOutputTokens is int32 (error said int32).
	
	temp := float32(0.2)
	maxTokens := int32(1024)
	
	resp, err := s.client.Models.GenerateContent(ctx, "gemini-2.0-flash-exp", genai.Text(prompt), &genai.GenerateContentConfig{
		Temperature:     &temp,
		MaxOutputTokens: maxTokens, 
		ResponseMIMEType: "application/json", 
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call Gemini: %v", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	// Parse Logic
	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		// part is *genai.Part struct.
		if part.Text != "" {
			responseText += part.Text
		}
	}
	
	// Clean markdown json if present
	responseText = strings.TrimSpace(responseText)
	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
	}

	var decision models.GeminiDecisionResponse
	if err := json.Unmarshal([]byte(responseText), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %v. Response: %s", err, responseText)
	}

	return &decision, nil
}

func buildPrompt(mission *models.Mission, agent *models.Agent, page *models.StrippedPage) string {
	
	elementsJSON, _ := json.MarshalIndent(page.InteractiveElements, "", "  ")

	systemPrompt := mission.InitialSystemPrompt
	if systemPrompt == "" {
		systemPrompt = "You are an AI agent."
	}

	return fmt.Sprintf(`%s

Current Goal: %s
Current URL: %s

Page Content:
%s

Interactable Elements:
%s

Agent History (Last 5 actions):
%s

Instructions:
1. Analyze the page and history.
2. Decide the next best action to assume to achieve the goal.
3. If the goal is achieved, return action="completed".
4. If stuck or error, return action="failed" or try "go_back".
5. Respond strictly in JSON format matching this schema:
{
  "reasoning": "Reasoning ...",
  "action": "click" | "type" | "wait" | "go_back" | "visit" | "completed" | "failed",
  "selector": "css_selector",
  "text_input": "text to type (optional)"
}
`, systemPrompt, mission.Goal, agent.CurrentURL, page.TextContent, string(elementsJSON), formatHistory(agent.ActionHistory))
}

func formatHistory(history []string) string {
	start := 0
	if len(history) > 5 {
		start = len(history) - 5
	}
	return strings.Join(history[start:], "\n")
}
