package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"

	"wtfsiw/internal/ai/tools"
)

// ClaudeChatProvider implements ChatProvider using Anthropic's Claude API
type ClaudeChatProvider struct {
	client anthropic.Client
}

// NewClaudeChatProvider creates a new Claude chat provider
func NewClaudeChatProvider(apiKey string) *ClaudeChatProvider {
	return &ClaudeChatProvider{
		client: anthropic.NewClient(option.WithAPIKey(apiKey)),
	}
}

// SendMessage sends messages to Claude and returns the response
func (p *ClaudeChatProvider) SendMessage(ctx context.Context, messages []ChatMessage, toolDefs []tools.ToolDefinition) (*ChatResponse, error) {
	// Convert messages to Claude format
	claudeMessages := make([]anthropic.MessageParam, 0, len(messages))

	for _, msg := range messages {
		claudeMsg := convertToClaudeMessage(msg)
		if claudeMsg != nil {
			claudeMessages = append(claudeMessages, *claudeMsg)
		}
	}

	// Convert tools to Claude format
	claudeTools := toClaudeTools(toolDefs)

	// Make API call
	resp, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5Haiku20241022,
		MaxTokens: 4096,
		System: []anthropic.TextBlockParam{
			{Text: chatSystemPrompt},
		},
		Messages: claudeMessages,
		Tools:    claudeTools,
	})
	if err != nil {
		return nil, fmt.Errorf("Claude API error: %w", err)
	}

	// Parse response
	return parseClaudeResponse(resp)
}

func convertToClaudeMessage(msg ChatMessage) *anthropic.MessageParam {
	switch msg.Role {
	case "user":
		m := anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
		return &m

	case "assistant":
		// Build content blocks
		var blocks []anthropic.ContentBlockParamUnion

		// Add text content if present
		if msg.Content != "" {
			blocks = append(blocks, anthropic.NewTextBlock(msg.Content))
		}

		// Add tool use blocks if present
		for _, tc := range msg.ToolCalls {
			blocks = append(blocks, anthropic.ContentBlockParamUnion{
				OfToolUse: &anthropic.ToolUseBlockParam{
					ID:    tc.ID,
					Name:  tc.Name,
					Input: tc.Arguments,
				},
			})
		}

		if len(blocks) == 0 {
			return nil
		}

		m := anthropic.NewAssistantMessage(blocks...)
		return &m

	case "tool":
		// Tool results in Claude are sent as user messages with tool_result blocks
		m := anthropic.MessageParam{
			Role: anthropic.MessageParamRoleUser,
			Content: []anthropic.ContentBlockParamUnion{
				{
					OfToolResult: &anthropic.ToolResultBlockParam{
						ToolUseID: msg.ToolCallID,
						Content: []anthropic.ToolResultBlockParamContentUnion{
							{
								OfText: &anthropic.TextBlockParam{
									Text: msg.Content,
								},
							},
						},
					},
				},
			},
		}
		return &m

	default:
		return nil
	}
}

func toClaudeTools(toolDefs []tools.ToolDefinition) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(toolDefs))
	for i, tool := range toolDefs {
		// Build input schema from parameters
		properties := make(map[string]interface{})
		required := make([]string, 0)

		for _, p := range tool.Parameters {
			properties[p.Name] = paramToAnthropicSchema(p)
			if p.Required {
				required = append(required, p.Name)
			}
		}

		inputSchema := anthropic.ToolInputSchemaParam{
			Properties: properties,
			Required:   required,
		}

		result[i] = anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: inputSchema,
			},
		}
	}
	return result
}

func paramToAnthropicSchema(p tools.ToolParameter) map[string]interface{} {
	schema := map[string]interface{}{
		"type":        p.Type,
		"description": p.Description,
	}

	if len(p.Enum) > 0 {
		schema["enum"] = p.Enum
	}

	if p.Type == "array" && p.Items != nil {
		schema["items"] = paramToAnthropicSchema(*p.Items)
	}

	return schema
}

func parseClaudeResponse(resp *anthropic.Message) (*ChatResponse, error) {
	var textContent string
	var toolCalls []tools.ToolCall

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			textContent += block.Text
		case "tool_use":
			// Parse tool use block - Input is json.RawMessage
			args := make(map[string]interface{})
			if len(block.Input) > 0 {
				if err := json.Unmarshal(block.Input, &args); err != nil {
					// Log error but continue
					args = make(map[string]interface{})
				}
			}
			toolCalls = append(toolCalls, tools.ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: args,
			})
		}
	}

	stopReason := "end_turn"
	if len(toolCalls) > 0 {
		stopReason = "tool_use"
	} else if resp.StopReason == anthropic.StopReasonMaxTokens {
		stopReason = "max_tokens"
	}

	return &ChatResponse{
		Content:    textContent,
		ToolCalls:  toolCalls,
		StopReason: stopReason,
	}, nil
}
