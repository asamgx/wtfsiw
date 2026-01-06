package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("#FF6B6B")
	secondaryColor = lipgloss.Color("#4ECDC4")
	accentColor    = lipgloss.Color("#FFE66D")
	mutedColor     = lipgloss.Color("#666666")
	bgColor        = lipgloss.Color("#1a1a2e")
	cardBgColor    = lipgloss.Color("#16213e")

	// App container
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title/header
	titleStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginBottom(1)

	// Subtitle
	subtitleStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Input
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1).
			MarginBottom(1)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	// Results list
	listItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), false, false, false, true).
				BorderForeground(primaryColor).
				PaddingLeft(1).
				Foreground(primaryColor)

	// Media card
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(1, 2).
			MarginBottom(1)

	mediaTitleStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	mediaYearStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	mediaTypeStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Background(lipgloss.Color("#0d1b2a")).
			Padding(0, 1)

	ratingStyle = lipgloss.NewStyle().
			Foreground(accentColor).
			Bold(true)

	overviewStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#cccccc")).
			MarginTop(1)

	// Providers
	providerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff")).
			Background(primaryColor).
			Padding(0, 1).
			MarginRight(1)

	// Status/loading
	spinnerStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	statusStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	// Error
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff4444")).
			Bold(true)

	// Help
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)
)

// RenderRating returns a formatted rating string with stars for detail view
func RenderRating(rating float64) string {
	return ratingStyle.Render(renderStars(rating) + " " + formatRating(rating))
}

// RenderRatingCompact returns a compact rating for list view (stars + number)
func RenderRatingCompact(rating float64) string {
	if rating == 0 {
		return ratingStyle.Render("☆ N/A")
	}
	return ratingStyle.Render(renderStars(rating) + " " + formatFloat(rating))
}

func renderStars(rating float64) string {
	// Convert 0-10 scale to 0-5 stars
	stars := int(rating / 2)
	halfStar := (rating/2 - float64(stars)) >= 0.5

	result := ""
	for i := 0; i < 5; i++ {
		if i < stars {
			result += "★"
		} else if i == stars && halfStar {
			result += "✦" // half star
		} else {
			result += "☆"
		}
	}
	return result
}

func formatRating(r float64) string {
	if r == 0 {
		return "N/A"
	}
	return formatFloat(r) + "/10"
}

func formatFloat(f float64) string {
	whole := int(f)
	frac := int((f - float64(whole)) * 10)
	return intToStr(whole) + "." + intToStr(frac)
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}
	result := ""
	for n > 0 {
		result = string(digits[n%10]) + result
		n /= 10
	}
	return result
}
