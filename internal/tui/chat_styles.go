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

	// Media card styles
	cardContainerStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(surface2).
				Padding(0, 1).
				MarginLeft(2)

	cardSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(mauve).
				Padding(0, 1).
				MarginLeft(2)

	cardTitleStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	cardYearStyle = lipgloss.NewStyle().
			Foreground(subtext0)

	cardRatingStyle = lipgloss.NewStyle().
			Foreground(yellow)

	cardProviderStyle = lipgloss.NewStyle().
				Foreground(base).
				Background(teal).
				Padding(0, 1).
				MarginRight(1)

	cardWhyWatchStyle = lipgloss.NewStyle().
				Foreground(green).
				Italic(true)

	cardIndexStyle = lipgloss.NewStyle().
			Foreground(mauve).
			Bold(true)

	cardHeaderStyle = lipgloss.NewStyle().
			Foreground(lavender).
			Italic(true).
			MarginBottom(1)
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

// RenderMediaCard renders a single media card in compact format
// Format:
//   [idx] 🎬 Title (Year)  ★★★★☆ 8.2
//        Netflix  Prime
//        💡 Why watch text...
func RenderMediaCard(card MediaCard, index int, selected bool, width int) string {
	// Media type emoji
	emoji := "🎬"
	if card.MediaType == "tv" {
		emoji = "📺"
	}

	// Line 1: Index + emoji + title + year + rating
	indexStr := cardIndexStyle.Render(intToStr(index) + ".")
	title := cardTitleStyle.Render(card.Title)
	year := cardYearStyle.Render("(" + card.Year + ")")
	rating := cardRatingStyle.Render(renderStars(card.Rating) + " " + formatFloat(card.Rating))

	line1 := indexStr + " " + emoji + " " + title + " " + year + "  " + rating

	// Line 2: Providers (if any)
	var line2 string
	if len(card.Providers) > 0 {
		line2 = "   "
		for i, p := range card.Providers {
			if i > 3 { // Limit to 4 providers
				line2 += cardYearStyle.Render("+more")
				break
			}
			line2 += cardProviderStyle.Render(p) + " "
		}
	}

	// Line 3: Why watch (if present, truncated)
	var line3 string
	if card.WhyWatch != "" {
		why := card.WhyWatch
		maxLen := width - 10
		if maxLen < 30 {
			maxLen = 30
		}
		if len(why) > maxLen {
			why = why[:maxLen-3] + "..."
		}
		line3 = "   " + cardWhyWatchStyle.Render("💡 "+why)
	}

	// Build content
	content := line1
	if line2 != "" {
		content += "\n" + line2
	}
	if line3 != "" {
		content += "\n" + line3
	}

	// Apply container style
	if selected {
		return cardSelectedStyle.Render(content)
	}
	return cardContainerStyle.Render(content)
}

// RenderMediaCardGroup renders a group of media cards with optional selection
func RenderMediaCardGroup(cards []MediaCard, selection *CardSelection, itemIndex int, width int) string {
	if len(cards) == 0 {
		return ""
	}

	var result string

	// Header
	countText := intToStr(len(cards))
	if len(cards) == 1 {
		result = cardHeaderStyle.Render("Found " + countText + " result:")
	} else {
		result = cardHeaderStyle.Render("Found " + countText + " results:")
	}
	result += "\n"

	// Render each card
	for i, card := range cards {
		isSelected := selection != nil && selection.ItemIndex == itemIndex && selection.CardIndex == i
		result += RenderMediaCard(card, i+1, isSelected, width)
		if i < len(cards)-1 {
			result += "\n"
		}
	}

	return result
}
