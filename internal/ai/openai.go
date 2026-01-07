package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
}

func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	client := openai.NewClient(apiKey)
	return &OpenAIProvider{client: client}
}

// cleanNumericFields fixes common JSON issues where empty strings are used instead of 0 for numeric fields
func cleanNumericFields(jsonStr string) string {
	// Replace empty strings with 0 for known numeric fields
	numericFields := []string{
		"year_from", "year_to", "min_rating", "max_runtime", "vote_count",
		"min_vote_count", // new field
	}
	for _, field := range numericFields {
		// Match "field_name":"" and replace with "field_name":0
		pattern := regexp.MustCompile(`"` + field + `"\s*:\s*""`)
		jsonStr = pattern.ReplaceAllString(jsonStr, `"`+field+`":0`)
	}
	return jsonStr
}

func (p *OpenAIProvider) ExtractSearchParams(ctx context.Context, query string) (*SearchParams, error) {
	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: getSystemPromptExtract(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: query,
			},
		},
		MaxTokens: 1024,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	responseText := resp.Choices[0].Message.Content

	// Clean up common JSON issues (empty strings for numeric fields)
	responseText = cleanNumericFields(responseText)

	// Parse JSON response
	var params SearchParams
	if err := json.Unmarshal([]byte(responseText), &params); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response as JSON: %w\nResponse: %s", err, responseText)
	}

	// Set defaults if not specified
	if params.MediaType == "" {
		params.MediaType = "all"
	}

	return &params, nil
}

func (p *OpenAIProvider) GetRecommendations(ctx context.Context, query string, count int) (*RecommendationResponse, error) {
	userPrompt := fmt.Sprintf("Please recommend %d movies or TV shows based on this request: %s", count, query)

	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPromptRecommend,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		},
		MaxTokens: 4096,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("openai API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	responseText := resp.Choices[0].Message.Content

	// Parse JSON response
	var result RecommendationResponse
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response as JSON: %w\nResponse: %s", err, responseText)
	}

	// Mark all recommendations as from AI
	for i := range result.Recommendations {
		result.Recommendations[i].FromAI = true
	}

	return &result, nil
}
