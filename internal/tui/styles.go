package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Catppuccin Mocha color palette
var (
	// Accent colors
	rosewater = lipgloss.Color("#f5e0dc")
	flamingo  = lipgloss.Color("#f2cdcd")
	pink      = lipgloss.Color("#f5c2e7")
	mauve     = lipgloss.Color("#cba6f7")
	red       = lipgloss.Color("#f38ba8")
	maroon    = lipgloss.Color("#eba0ac")
	peach     = lipgloss.Color("#fab387")
	yellow    = lipgloss.Color("#f9e2af")
	green     = lipgloss.Color("#a6e3a1")
	teal      = lipgloss.Color("#94e2d5")
	sky       = lipgloss.Color("#89dceb")
	sapphire  = lipgloss.Color("#74c7ec")
	blue      = lipgloss.Color("#89b4fa")
	lavender  = lipgloss.Color("#b4befe")

	// Text colors
	text     = lipgloss.Color("#cdd6f4")
	subtext1 = lipgloss.Color("#bac2de")
	subtext0 = lipgloss.Color("#a6adc8")

	// Overlay colors
	overlay2 = lipgloss.Color("#9399b2")
	overlay1 = lipgloss.Color("#7f849c")
	overlay0 = lipgloss.Color("#6c7086")

	// Surface colors
	surface2 = lipgloss.Color("#585b70")
	surface1 = lipgloss.Color("#45475a")
	surface0 = lipgloss.Color("#313244")

	// Base colors
	base   = lipgloss.Color("#1e1e2e")
	mantle = lipgloss.Color("#181825")
	crust  = lipgloss.Color("#11111b")
)

// Semantic color aliases
var (
	primaryColor   = mauve
	secondaryColor = teal
	accentColor    = yellow
	mutedColor     = overlay1
	bgColor        = base
	cardBgColor    = surface0

	// App container
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	// Title/header
	titleStyle = lipgloss.NewStyle().
			Foreground(mauve).
			Bold(true).
			MarginBottom(1)

	// Subtitle
	subtitleStyle = lipgloss.NewStyle().
			Foreground(subtext0).
			Italic(true)

	// Input
	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surface2).
			Padding(0, 1).
			MarginBottom(1)

	inputPromptStyle = lipgloss.NewStyle().
				Foreground(teal).
				Bold(true)

	// Results list
	listItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	selectedItemStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), false, false, false, true).
				BorderForeground(mauve).
				PaddingLeft(1).
				Foreground(lavender)

	// Media card
	cardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surface2).
			Padding(1, 2).
			MarginBottom(1)

	mediaTitleStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	mediaYearStyle = lipgloss.NewStyle().
			Foreground(subtext0)

	mediaTypeStyle = lipgloss.NewStyle().
			Foreground(base).
			Background(mauve).
			Padding(0, 1)

	ratingStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	overviewStyle = lipgloss.NewStyle().
			Foreground(text).
			MarginTop(1)

	// Providers
	providerStyle = lipgloss.NewStyle().
			Foreground(base).
			Background(teal).
			Padding(0, 1).
			MarginRight(1)

	// Status/loading
	spinnerStyle = lipgloss.NewStyle().
			Foreground(mauve)

	statusStyle = lipgloss.NewStyle().
			Foreground(subtext0).
			Italic(true)

	// Error
	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	// Help
	helpStyle = lipgloss.NewStyle().
			Foreground(overlay1).
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
