package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"wtfsiw/internal/ai"
	"wtfsiw/internal/tmdb"
)

// State represents the current view state
type State int

const (
	StateInput State = iota
	StateLoading
	StateResults
	StateDetail
	StateError
)

// Model is the main Bubble Tea model
type Model struct {
	state       State
	input       textinput.Model
	spinner     spinner.Model
	results     []ai.Recommendation
	summary     string // AI summary of what was searched for
	selected    int
	err         error
	statusMsg   string
	width       int
	height      int
	aiProvider  ai.Provider
	tmdbClient  *tmdb.Client // nil if TMDb not configured
	query       string
}

// Messages
type searchCompleteMsg struct {
	results []ai.Recommendation
	summary string
}

type searchErrorMsg struct {
	err error
}

type statusMsg string

// NewModel creates a new TUI model
func NewModel(aiProvider ai.Provider, tmdbClient *tmdb.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "e.g., something dark and psychological like Breaking Bad"
	ti.Focus()
	ti.CharLimit = 500
	ti.Width = 60

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		state:      StateInput,
		input:      ti,
		spinner:    s,
		aiProvider: aiProvider,
		tmdbClient: tmdbClient,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = min(60, msg.Width-10)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case searchCompleteMsg:
		m.results = msg.results
		m.summary = msg.summary
		m.selected = 0
		if len(msg.results) == 0 {
			m.state = StateError
			m.err = fmt.Errorf("no results found for your query")
		} else {
			m.state = StateResults
		}
		return m, nil

	case searchErrorMsg:
		m.state = StateError
		m.err = msg.err
		return m, nil

	case statusMsg:
		m.statusMsg = string(msg)
		return m, nil
	}

	// Update text input
	if m.state == StateInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.state == StateInput || m.state == StateError {
			return m, tea.Quit
		}
		// In other states, go back
		if m.state == StateDetail {
			m.state = StateResults
		} else if m.state == StateResults {
			m.state = StateInput
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "esc":
		if m.state == StateDetail {
			m.state = StateResults
		} else if m.state == StateResults {
			m.state = StateInput
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink
		} else if m.state == StateError {
			m.state = StateInput
			m.err = nil
			m.input.Focus()
			return m, textinput.Blink
		}
		return m, nil

	case "enter":
		if m.state == StateInput && m.input.Value() != "" {
			m.query = m.input.Value()
			m.state = StateLoading
			m.statusMsg = "Analyzing your request..."
			return m, tea.Batch(m.spinner.Tick, m.performSearch())
		}
		if m.state == StateResults && len(m.results) > 0 {
			m.state = StateDetail
		}
		return m, nil

	case "up", "k":
		if m.state == StateResults && m.selected > 0 {
			m.selected--
		}
		return m, nil

	case "down", "j":
		if m.state == StateResults && m.selected < len(m.results)-1 {
			m.selected++
		}
		return m, nil
	}

	// Pass to text input if in input state
	if m.state == StateInput {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) performSearch() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		// If TMDb is not configured, use AI directly
		if m.tmdbClient == nil {
			return m.searchWithAI(ctx)
		}

		// Otherwise, use TMDb with AI for search params
		return m.searchWithTMDb(ctx)
	}
}

func (m Model) searchWithAI(ctx context.Context) tea.Msg {
	resp, err := m.aiProvider.GetRecommendations(ctx, m.query, 10)
	if err != nil {
		return searchErrorMsg{err: fmt.Errorf("AI recommendation failed: %w", err)}
	}

	return searchCompleteMsg{
		results: resp.Recommendations,
		summary: resp.Summary,
	}
}

func (m Model) searchWithTMDb(ctx context.Context) tea.Msg {
	// Extract search params using AI
	params, err := m.aiProvider.ExtractSearchParams(ctx, m.query)
	if err != nil {
		return searchErrorMsg{err: fmt.Errorf("AI analysis failed: %w", err)}
	}

	// Search TMDb
	resp, err := m.tmdbClient.Discover(params)
	if err != nil {
		return searchErrorMsg{err: fmt.Errorf("search failed: %w", err)}
	}

	// Enrich with streaming providers
	m.tmdbClient.EnrichWithProviders(resp.Results)

	// Convert TMDb results to Recommendations
	recommendations := make([]ai.Recommendation, len(resp.Results))
	for i, media := range resp.Results {
		// Get provider names
		providers := make([]string, len(media.Providers))
		for j, p := range media.Providers {
			providers[j] = p.Name
		}

		recommendations[i] = ai.Recommendation{
			Title:     media.GetDisplayTitle(),
			Year:      media.GetDisplayYear(),
			MediaType: media.MediaType,
			Rating:    media.VoteAverage,
			Overview:  media.Overview,
			Providers: providers,
			VoteCount: media.VoteCount,
			FromAI:    false,
		}
	}

	summary := fmt.Sprintf("Searched for: %s", strings.Join(params.Keywords, ", "))
	if len(params.Genres) > 0 {
		summary += fmt.Sprintf(" in genres: %s", strings.Join(params.Genres, ", "))
	}

	return searchCompleteMsg{
		results: recommendations,
		summary: summary,
	}
}

func (m Model) View() string {
	var content string

	switch m.state {
	case StateInput:
		content = m.viewInput()
	case StateLoading:
		content = m.viewLoading()
	case StateResults:
		content = m.viewResults()
	case StateDetail:
		content = m.viewDetail()
	case StateError:
		content = m.viewError()
	}

	return appStyle.Render(content)
}

func (m Model) viewInput() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("wtfsiw"))
	sb.WriteString(" ")
	sb.WriteString(subtitleStyle.Render("What The Fuck Should I Watch?"))
	sb.WriteString("\n\n")

	// Show mode indicator
	if m.tmdbClient == nil {
		sb.WriteString(statusStyle.Render("(AI-only mode - TMDb not configured)"))
		sb.WriteString("\n\n")
	}

	sb.WriteString(inputPromptStyle.Render("What are you in the mood for?"))
	sb.WriteString("\n")
	sb.WriteString(inputStyle.Render(m.input.View()))
	sb.WriteString("\n\n")

	sb.WriteString(helpStyle.Render("Examples:"))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  • something dark and psychological like Breaking Bad"))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  • a feel-good comedy from the 90s"))
	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("  • Korean thriller, recent"))
	sb.WriteString("\n\n")

	sb.WriteString(helpStyle.Render("Press Enter to search • q to quit"))

	return sb.String()
}

func (m Model) viewLoading() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("wtfsiw"))
	sb.WriteString("\n\n")

	sb.WriteString(m.spinner.View())
	sb.WriteString(" ")
	sb.WriteString(statusStyle.Render(m.statusMsg))
	sb.WriteString("\n\n")

	sb.WriteString(subtitleStyle.Render("Query: " + m.query))

	return sb.String()
}

func (m Model) viewResults() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("Results"))
	sb.WriteString(" ")
	sb.WriteString(subtitleStyle.Render(fmt.Sprintf("(%d found)", len(m.results))))
	sb.WriteString("\n")

	// Show summary if available
	if m.summary != "" {
		sb.WriteString(statusStyle.Render(m.summary))
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	for i, rec := range m.results {
		line := m.renderResultLine(rec, i == m.selected)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(helpStyle.Render("↑/↓ navigate • Enter view details • Esc back • q quit"))

	return sb.String()
}

func (m Model) renderResultLine(rec ai.Recommendation, selected bool) string {
	// Media type badge
	mediaType := "MOVIE"
	if rec.MediaType == "tv" {
		mediaType = "TV"
	}
	badge := mediaTypeStyle.Render(mediaType)

	// Provider badges
	var providerBadges string
	for _, p := range rec.Providers {
		if abbr := providerEmoji(p); abbr != "" {
			providerBadges += providerStyle.Render(abbr) + " "
		}
	}
	if len(rec.Providers) > 0 && providerBadges == "" {
		// Show first provider name if no emoji match
		providerBadges = providerStyle.Render(truncate(rec.Providers[0], 8)) + " "
	}

	// AI indicator
	aiIndicator := ""
	if rec.FromAI {
		aiIndicator = statusStyle.Render(" [AI]")
	}

	line := fmt.Sprintf("%s %s (%s) %s %s%s",
		badge,
		mediaTitleStyle.Render(truncate(rec.Title, 35)),
		mediaYearStyle.Render(rec.Year),
		RenderRatingCompact(rec.Rating),
		providerBadges,
		aiIndicator,
	)

	if selected {
		return selectedItemStyle.Render(line)
	}
	return listItemStyle.Render(line)
}

func (m Model) viewDetail() string {
	if m.selected >= len(m.results) {
		return "No selection"
	}

	rec := m.results[m.selected]
	var sb strings.Builder

	// Title and year
	mediaType := "Movie"
	if rec.MediaType == "tv" {
		mediaType = "TV Show"
	}

	sb.WriteString(mediaTitleStyle.Render(rec.Title))
	sb.WriteString(" ")
	sb.WriteString(mediaYearStyle.Render("(" + rec.Year + ")"))
	sb.WriteString("\n")
	sb.WriteString(mediaTypeStyle.Render(mediaType))
	if rec.FromAI {
		sb.WriteString(" ")
		sb.WriteString(statusStyle.Render("[AI Recommendation]"))
	}
	sb.WriteString("\n\n")

	// Rating
	sb.WriteString(RenderRating(rec.Rating))
	if rec.VoteCount > 0 {
		sb.WriteString(subtitleStyle.Render(fmt.Sprintf(" (%d votes)", rec.VoteCount)))
	} else if rec.FromAI {
		sb.WriteString(subtitleStyle.Render(" (estimated rating)"))
	}
	sb.WriteString("\n\n")

	// Genres
	if len(rec.Genres) > 0 {
		sb.WriteString(inputPromptStyle.Render("Genres: "))
		sb.WriteString(strings.Join(rec.Genres, ", "))
		sb.WriteString("\n\n")
	}

	// Streaming providers
	if len(rec.Providers) > 0 {
		sb.WriteString(inputPromptStyle.Render("Where to Watch:"))
		sb.WriteString("\n")
		for _, p := range rec.Providers {
			sb.WriteString("  • ")
			sb.WriteString(p)
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Overview
	sb.WriteString(inputPromptStyle.Render("Overview:"))
	sb.WriteString("\n")
	overview := rec.Overview
	if overview == "" {
		overview = "No overview available."
	}
	// Word wrap overview
	wrapped := wordWrap(overview, min(70, m.width-10))
	sb.WriteString(overviewStyle.Render(wrapped))
	sb.WriteString("\n\n")

	// Why watch (AI recommendation reason)
	if rec.WhyWatch != "" {
		sb.WriteString(inputPromptStyle.Render("Why Watch:"))
		sb.WriteString("\n")
		wrapped := wordWrap(rec.WhyWatch, min(70, m.width-10))
		sb.WriteString(overviewStyle.Render(wrapped))
		sb.WriteString("\n\n")
	}

	sb.WriteString(helpStyle.Render("Esc back to results • q quit"))

	return cardStyle.Render(sb.String())
}

func (m Model) viewError() string {
	var sb strings.Builder

	sb.WriteString(titleStyle.Render("wtfsiw"))
	sb.WriteString("\n\n")

	sb.WriteString(errorStyle.Render("Error: "))
	sb.WriteString(m.err.Error())
	sb.WriteString("\n\n")

	sb.WriteString(helpStyle.Render("Press Esc to try again • q to quit"))

	return sb.String()
}

// Helper functions
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func wordWrap(s string, width int) string {
	if width <= 0 {
		width = 70
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var lines []string
	var current string

	for _, word := range words {
		if current == "" {
			current = word
		} else if len(current)+1+len(word) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}

	return strings.Join(lines, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// providerEmoji returns a short abbreviation for common streaming providers
func providerEmoji(name string) string {
	switch name {
	case "Netflix":
		return "N"
	case "Amazon Prime Video", "Prime Video":
		return "P"
	case "Disney Plus", "Disney+":
		return "D+"
	case "Hulu":
		return "H"
	case "HBO Max", "Max":
		return "M"
	case "Apple TV Plus", "Apple TV+":
		return "A+"
	case "Peacock", "Peacock Premium":
		return "Pk"
	case "Paramount Plus", "Paramount+":
		return "P+"
	case "Crunchyroll":
		return "CR"
	default:
		return ""
	}
}

// Run starts the TUI application
func Run(aiProvider ai.Provider, tmdbClient *tmdb.Client) error {
	p := tea.NewProgram(
		NewModel(aiProvider, tmdbClient),
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	return err
}
