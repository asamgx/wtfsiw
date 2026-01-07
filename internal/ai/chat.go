package ai

import (
	"context"
	"fmt"
	"time"

	"wtfsiw/internal/ai/tools"
	"wtfsiw/internal/config"
)

// ChatMessage represents a message in the conversation
type ChatMessage struct {
	Role       string            `json:"role"`        // "user", "assistant", "tool"
	Content    string            `json:"content"`
	ToolCalls  []tools.ToolCall  `json:"tool_calls,omitempty"`  // For assistant messages requesting tool use
	ToolCallID string            `json:"tool_call_id,omitempty"` // For tool result messages
	Timestamp  time.Time         `json:"timestamp"`
}

// ChatResponse represents the AI's response
type ChatResponse struct {
	Content    string            // Text content of the response
	ToolCalls  []tools.ToolCall  // Tools the AI wants to call
	StopReason string            // "end_turn", "tool_use", "max_tokens"
}

// ChatProvider defines the interface for chat-based AI providers with tool use
type ChatProvider interface {
	// SendMessage sends conversation messages and returns the response (may include tool calls)
	SendMessage(ctx context.Context, messages []ChatMessage, toolDefs []tools.ToolDefinition) (*ChatResponse, error)
}

// NewChatProvider creates a new chat provider based on config
func NewChatProvider() (ChatProvider, error) {
	cfg := config.Get()

	switch cfg.AI.Provider {
	case "claude":
		if cfg.AI.ClaudeAPIKey == "" {
			return nil, fmt.Errorf("Claude API key not configured. Set ANTHROPIC_API_KEY or run: wtfsiw config set ai.claude_api_key YOUR_KEY")
		}
		return NewClaudeChatProvider(cfg.AI.ClaudeAPIKey), nil
	case "openai":
		if cfg.AI.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OpenAI API key not configured. Set OPENAI_API_KEY or run: wtfsiw config set ai.openai_api_key YOUR_KEY")
		}
		return NewOpenAIChatProvider(cfg.AI.OpenAIAPIKey), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.AI.Provider)
	}
}

// Chat system prompt
const chatSystemPrompt = `You are a helpful movie and TV show recommendation assistant called "wtfsiw" (What The Fuck Should I Watch).

You have access to tools to help users find content to watch:
- search_media: Search TMDb for movies/TV shows with filters (genre, year, rating, language, streaming service, actors, studios)
- get_media_details: Get detailed info about a specific title
- get_streaming_providers: Check where something is available to watch
- get_similar: Find similar movies/shows to a given title
- search_by_title: Find a specific title by name
- get_trakt_watchlist: View the user's Trakt watchlist (if connected)
- get_trakt_history: View the user's watch history (if connected)
- generate_recommendations: Generate AI recommendations directly for complex/mood-based requests

When helping users:
1. Use search_media for discovery requests with specific criteria
2. Use search_by_title first when users mention a specific title, then get_similar for recommendations
3. Use get_streaming_providers to show where they can watch something
4. Use generate_recommendations for subjective requests that don't map well to filters

Format your responses clearly:
- Use numbered lists for multiple recommendations
- Include ratings (out of 10) and where to watch
- Explain why each recommendation matches their request
- Keep descriptions concise but helpful

If you're unsure what the user wants, ask clarifying questions.
Be conversational and helpful. You can remember context from earlier in the conversation.`
