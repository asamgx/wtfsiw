package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Chat container
	chatContainerStyle = lipgloss.NewStyle().
				Padding(1, 2)

	// Chat header
	chatHeaderStyle = lipgloss.NewStyle().
			Foreground(mauve).
			Bold(true).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(surface1).
			PaddingBottom(1).
			MarginBottom(1)

	// Message styles
	userMsgStyle = lipgloss.NewStyle().
			Foreground(text).
			PaddingLeft(2)

	userLabelStyle = lipgloss.NewStyle().
			Foreground(sapphire).
			Bold(true)

	assistantMsgStyle = lipgloss.NewStyle().
				Foreground(text).
				PaddingLeft(2)

	assistantLabelStyle = lipgloss.NewStyle().
				Foreground(lavender).
				Bold(true)

	toolMsgStyle = lipgloss.NewStyle().
			Foreground(subtext0).
			Italic(true).
			PaddingLeft(4)

	toolLabelStyle = lipgloss.NewStyle().
			Foreground(peach).
			Bold(true)

	systemMsgStyle = lipgloss.NewStyle().
			Foreground(overlay1).
			Italic(true).
			Align(lipgloss.Center)

	// Input area
	chatInputContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), true, false, false, false).
				BorderForeground(surface1).
				PaddingTop(1).
				MarginTop(1)

	chatInputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surface2).
			Padding(0, 1)

	// Thinking/loading indicator
	thinkingStyle = lipgloss.NewStyle().
			Foreground(lavender).
			Italic(true).
			PaddingLeft(2)

	// Tool execution indicator
	toolExecutingStyle = lipgloss.NewStyle().
				Foreground(peach).
				Bold(true).
				PaddingLeft(4)

	// Chat footer/help
	chatHelpStyle = lipgloss.NewStyle().
			Foreground(overlay1).
			MarginTop(1).
			Align(lipgloss.Center)

	// Scroll indicator
	scrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(overlay0).
				Align(lipgloss.Right)

	// Viewport focus style (highlighted border when scrolling)
	viewportFocusStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(mauve)
)

// FormatUserMessage formats a user message for display
func FormatUserMessage(content string) string {
	return userLabelStyle.Render("You: ") + userMsgStyle.Render(content)
}

// FormatAssistantMessage formats an assistant message for display
func FormatAssistantMessage(content string) string {
	return assistantLabelStyle.Render("AI: ") + assistantMsgStyle.Render(content)
}

// FormatToolCall formats a tool call for display
func FormatToolCall(name string) string {
	return toolLabelStyle.Render("  → ") + toolMsgStyle.Render(name)
}

// FormatToolResult formats a tool result summary for display
func FormatToolResult(name string, success bool) string {
	if success {
		checkStyle := lipgloss.NewStyle().Foreground(green)
		return checkStyle.Render("  ✓ ") + toolMsgStyle.Render(name)
	}
	crossStyle := lipgloss.NewStyle().Foreground(red)
	return crossStyle.Render("  ✗ ") + toolMsgStyle.Render(name)
}

// FormatThinking formats the thinking indicator
func FormatThinking() string {
	return thinkingStyle.Render("Thinking...")
}

// FormatSystemMessage formats a system message
func FormatSystemMessage(content string) string {
	return systemMsgStyle.Render(content)
}
