package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ClaudeProvider struct {
	client anthropic.Client
}

func NewClaudeProvider(apiKey string) *ClaudeProvider {
	return &ClaudeProvider{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
	}
}

func (p *ClaudeProvider) ExtractSearchParams(ctx context.Context, query string) (*SearchParams, error) {
	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5Haiku20241022,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: getSystemPromptExtract()},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(query)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	responseText := extractTextFromResponse(message)
	if responseText == "" {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Parse JSON response
	var params SearchParams
	if err := json.Unmarshal([]byte(responseText), &params); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response as JSON: %w\nResponse: %s", err, responseText)
	}

	// Set defaults if not specified
	if params.MediaType == "" {
		params.MediaType = "all"
	}

	return &params, nil
}

func (p *ClaudeProvider) GetRecommendations(ctx context.Context, query string, count int) (*RecommendationResponse, error) {
	userPrompt := fmt.Sprintf("Please recommend %d movies or TV shows based on this request: %s", count, query)

	message, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5Haiku20241022,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: systemPromptRecommend},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userPrompt)),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	responseText := extractTextFromResponse(message)
	if responseText == "" {
		return nil, fmt.Errorf("empty response from Claude")
	}

	// Parse JSON response
	var resp RecommendationResponse
	if err := json.Unmarshal([]byte(responseText), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response as JSON: %w\nResponse: %s", err, responseText)
	}

	// Mark all recommendations as from AI
	for i := range resp.Recommendations {
		resp.Recommendations[i].FromAI = true
	}

	return &resp, nil
}

func extractTextFromResponse(message *anthropic.Message) string {
	if len(message.Content) == 0 {
		return ""
	}
	for _, block := range message.Content {
		if block.Type == "text" {
			return block.Text
		}
	}
	return ""
}
