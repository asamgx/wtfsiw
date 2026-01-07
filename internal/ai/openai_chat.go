package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"wtfsiw/internal/ai/tools"
)

// OpenAIChatProvider implements ChatProvider using OpenAI's API
type OpenAIChatProvider struct {
	client *openai.Client
}

// NewOpenAIChatProvider creates a new OpenAI chat provider
func NewOpenAIChatProvider(apiKey string) *OpenAIChatProvider {
	client := openai.NewClient(apiKey)
	return &OpenAIChatProvider{client: client}
}

// SendMessage sends messages to OpenAI and returns the response
func (p *OpenAIChatProvider) SendMessage(ctx context.Context, messages []ChatMessage, toolDefs []tools.ToolDefinition) (*ChatResponse, error) {
	// Convert messages to OpenAI format
	oaiMessages := make([]openai.ChatCompletionMessage, 0, len(messages)+1)

	// Add system message
	oaiMessages = append(oaiMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: chatSystemPrompt,
	})

	// Convert chat messages
	for _, msg := range messages {
		oaiMsg := convertToOpenAIMessage(msg)
		oaiMessages = append(oaiMessages, oaiMsg)
	}

	// Convert tools
	oaiTools := tools.ToOpenAITools(toolDefs)

	// Make API call
	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: oaiMessages,
		Tools:    oaiTools,
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from OpenAI")
	}

	choice := resp.Choices[0]

	// Check if there are tool calls
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]tools.ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
				args = make(map[string]interface{})
			}
			toolCalls[i] = tools.ToolCall{
				ID:        tc.ID,
				Name:      tc.Function.Name,
				Arguments: args,
			}
		}
		return &ChatResponse{
			Content:    choice.Message.Content,
			ToolCalls:  toolCalls,
			StopReason: "tool_use",
		}, nil
	}

	// Regular text response
	return &ChatResponse{
		Content:    choice.Message.Content,
		StopReason: "end_turn",
	}, nil
}

func convertToOpenAIMessage(msg ChatMessage) openai.ChatCompletionMessage {
	switch msg.Role {
	case "user":
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Content,
		}

	case "assistant":
		oaiMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: msg.Content,
		}
		// If this assistant message had tool calls, include them
		if len(msg.ToolCalls) > 0 {
			oaiMsg.ToolCalls = make([]openai.ToolCall, len(msg.ToolCalls))
			for i, tc := range msg.ToolCalls {
				argsJSON, _ := json.Marshal(tc.Arguments)
				oaiMsg.ToolCalls[i] = openai.ToolCall{
					ID:   tc.ID,
					Type: openai.ToolTypeFunction,
					Function: openai.FunctionCall{
						Name:      tc.Name,
						Arguments: string(argsJSON),
					},
				}
			}
		}
		return oaiMsg

	case "tool":
		return openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    msg.Content,
			ToolCallID: msg.ToolCallID,
		}

	default:
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: msg.Content,
		}
	}
}
