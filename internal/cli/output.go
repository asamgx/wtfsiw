package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"

	"wtfsiw/internal/ai"
)

// Catppuccin Mocha colors
var (
	// Accent colors
	mauve    = lipgloss.Color("#cba6f7")
	red      = lipgloss.Color("#f38ba8")
	peach    = lipgloss.Color("#fab387")
	yellow   = lipgloss.Color("#f9e2af")
	green    = lipgloss.Color("#a6e3a1")
	teal     = lipgloss.Color("#94e2d5")
	sapphire = lipgloss.Color("#74c7ec")
	lavender = lipgloss.Color("#b4befe")

	// Text colors
	text     = lipgloss.Color("#cdd6f4")
	subtext0 = lipgloss.Color("#a6adc8")

	// Surface colors
	surface2 = lipgloss.Color("#585b70")
	surface1 = lipgloss.Color("#45475a")
	overlay1 = lipgloss.Color("#7f849c")

	// Base colors
	base = lipgloss.Color("#1e1e2e")

	// Semantic aliases
	primaryColor   = mauve
	secondaryColor = teal
	accentColor    = yellow
	mutedColor     = overlay1
	successColor   = green

	// Styles
	headerStyle = lipgloss.NewStyle().
			Foreground(mauve).
			Bold(true)

	queryStyle = lipgloss.NewStyle().
			Foreground(sapphire).
			Italic(true)

	titleStyle = lipgloss.NewStyle().
			Foreground(yellow).
			Bold(true)

	yearStyle = lipgloss.NewStyle().
			Foreground(subtext0)

	ratingStyle = lipgloss.NewStyle().
			Foreground(yellow)

	providerStyle = lipgloss.NewStyle().
			Foreground(base).
			Background(teal).
			Padding(0, 1)

	whyWatchStyle = lipgloss.NewStyle().
			Foreground(green).
			Italic(true)

	overviewStyle = lipgloss.NewStyle().
			Foreground(text)

	summaryStyle = lipgloss.NewStyle().
			Foreground(teal).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surface2).
			Padding(0, 1)

	indexStyle = lipgloss.NewStyle().
			Foreground(mauve).
			Bold(true)

	dividerStyle = lipgloss.NewStyle().
			Foreground(surface1)

	// Spinner frames
	spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
)

// Spinner handles animated loading indicator
type Spinner struct {
	message string
	done    chan bool
	ticker  *time.Ticker
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	return &Spinner{
		message: message,
		done:    make(chan bool),
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.ticker = time.NewTicker(80 * time.Millisecond)
	go func() {
		frame := 0
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				spinner := lipgloss.NewStyle().Foreground(secondaryColor).Render(spinnerFrames[frame])
				fmt.Printf("\r%s %s", spinner, s.message)
				frame = (frame + 1) % len(spinnerFrames)
			}
		}
	}()
}

// Stop ends the spinner animation
func (s *Spinner) Stop() {
	s.ticker.Stop()
	s.done <- true
	// Clear the line
	fmt.Print("\r\033[K")
}

// StopWithMessage ends spinner and shows a completion message
func (s *Spinner) StopWithMessage(msg string) {
	s.ticker.Stop()
	s.done <- true
	fmt.Print("\r\033[K")
	checkmark := lipgloss.NewStyle().Foreground(successColor).Render("✓")
	fmt.Printf("%s %s\n", checkmark, msg)
}

// PrintHeader prints the app header with query
func PrintHeader(query string) {
	fmt.Println()
	header := headerStyle.Render("🎬 What The Fuck Should I Watch?")
	fmt.Println(header)
	fmt.Println()
	fmt.Printf("   %s %s\n\n", lipgloss.NewStyle().Foreground(mutedColor).Render("Searching:"), queryStyle.Render(query))
}

// PrintSummary prints the result summary in a styled box
func PrintSummary(summary string) {
	fmt.Println(summaryStyle.Render("📋 " + summary))
	fmt.Println()
}

// PrintDivider prints a styled divider
func PrintDivider() {
	width := getTerminalWidth()
	if width > 60 {
		width = 60
	}
	divider := strings.Repeat("─", width)
	fmt.Println(dividerStyle.Render(divider))
}

// PrintRecommendation prints a single recommendation with animations
func PrintRecommendation(index int, rec ai.Recommendation, animate bool) {
	// Media type emoji
	mediaEmoji := "🎬"
	if rec.MediaType == "tv" {
		mediaEmoji = "📺"
	}

	// Build the recommendation display
	indexStr := indexStyle.Render(fmt.Sprintf("%d.", index))
	title := titleStyle.Render(rec.Title)
	year := yearStyle.Render(fmt.Sprintf("(%s)", rec.Year))

	// Rating with stars
	stars := renderStars(rec.Rating)
	ratingStr := ratingStyle.Render(fmt.Sprintf("%s %.1f/10", stars, rec.Rating))

	// Print with optional animation
	if animate {
		// Typewriter effect for title
		fmt.Printf("%s %s ", indexStr, mediaEmoji)
		typewriter(title+" "+year, 15*time.Millisecond)
		fmt.Println()
	} else {
		fmt.Printf("%s %s %s %s\n", indexStr, mediaEmoji, title, year)
	}

	fmt.Printf("   %s\n", ratingStr)

	// Providers
	if len(rec.Providers) > 0 {
		providerStr := "   📍 "
		for i, p := range rec.Providers {
			if i > 0 {
				providerStr += " "
			}
			providerStr += providerStyle.Render(p)
		}
		fmt.Println(providerStr)
	}

	// Why watch (AI explanation)
	if rec.WhyWatch != "" {
		why := whyWatchStyle.Render("💡 " + rec.WhyWatch)
		fmt.Printf("   %s\n", why)
	}

	// Overview (truncated)
	if rec.Overview != "" {
		overview := rec.Overview
		maxLen := getTerminalWidth() - 6
		if maxLen > 120 {
			maxLen = 120
		}
		if len(overview) > maxLen {
			overview = overview[:maxLen-3] + "..."
		}
		fmt.Printf("   %s\n", overviewStyle.Render(overview))
	}

	fmt.Println()
}

// PrintResults prints all recommendations
func PrintResults(recommendations []ai.Recommendation, animate bool) {
	for i, rec := range recommendations {
		PrintRecommendation(i+1, rec, animate && i < 3) // Only animate first 3
		if animate && i < len(recommendations)-1 {
			time.Sleep(100 * time.Millisecond) // Small delay between items
		}
	}
}

// PrintNoResults shows a styled "no results" message
func PrintNoResults() {
	msg := lipgloss.NewStyle().
		Foreground(mutedColor).
		Italic(true).
		Render("No results found. Try a different query!")
	fmt.Println(msg)
}

// PrintError shows a styled error message
func PrintError(err error) {
	errStyle := lipgloss.NewStyle().
		Foreground(red).
		Bold(true)
	fmt.Printf("%s %s\n", errStyle.Render("✗"), err.Error())
}

// typewriter prints text with a typewriter effect
func typewriter(text string, delay time.Duration) {
	for _, char := range text {
		fmt.Print(string(char))
		time.Sleep(delay)
	}
}

func renderStars(rating float64) string {
	stars := int(rating / 2)
	halfStar := (rating/2 - float64(stars)) >= 0.5

	result := ""
	for i := 0; i < 5; i++ {
		if i < stars {
			result += "★"
		} else if i == stars && halfStar {
			result += "✦"
		} else {
			result += "☆"
		}
	}
	return result
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // default
	}
	return width
}
