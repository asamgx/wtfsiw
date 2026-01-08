package tui

import (
	"encoding/json"
	"strings"
)

// DisplayItemType represents the type of display item
type DisplayItemType int

const (
	DisplayItemText DisplayItemType = iota
	DisplayItemCards
)

// DisplayItem represents either a plain text message or a media card group
type DisplayItem struct {
	Type       DisplayItemType
	Text       string      // For text messages
	MediaCards []MediaCard // For card groups from tool results
	ToolName   string      // Which tool produced these cards
}

// MediaCard represents a single movie/TV show card
type MediaCard struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Year      string   `json:"year"`
	MediaType string   `json:"media_type"`
	Rating    float64  `json:"rating"`
	VoteCount int      `json:"vote_count"`
	Providers []string `json:"providers"`
	WhyWatch  string   `json:"why_watch"`
	Overview  string   `json:"overview"`
}

// CardSelection tracks which card is currently selected
type CardSelection struct {
	ItemIndex  int // Which DisplayItem contains the cards
	CardIndex  int // Which card within that group is selected
	TotalCards int // Total cards in current group
}

// MediaTools lists tools that return media results
var MediaTools = map[string]bool{
	"search_media":             true,
	"get_similar":              true,
	"search_by_title":          true,
	"generate_recommendations": true,
}

// IsMediaTool checks if a tool name returns media results
func IsMediaTool(name string) bool {
	return MediaTools[name]
}

// tmdbMediaResult represents the JSON format from TMDb tool results
type tmdbMediaResult struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Name      string   `json:"name"` // TV shows use "name"
	Year      string   `json:"year"`
	MediaType string   `json:"media_type"`
	Rating    float64  `json:"rating"`
	VoteCount int      `json:"vote_count"`
	Overview  string   `json:"overview"`
	Providers []string `json:"providers"`
}

// aiRecommendationResult represents the JSON format from AI recommendation tool
type aiRecommendationResult struct {
	Summary         string `json:"summary"`
	Recommendations []struct {
		Title     string   `json:"title"`
		Year      string   `json:"year"`
		MediaType string   `json:"media_type"`
		Rating    float64  `json:"rating"`
		Genres    []string `json:"genres"`
		Overview  string   `json:"overview"`
		WhyWatch  string   `json:"why_watch"`
		Providers []string `json:"providers"`
	} `json:"recommendations"`
}

// ParseMediaCards attempts to parse JSON tool result into MediaCards
// It handles both TMDb array format and AI recommendation format
func ParseMediaCards(jsonStr string) ([]MediaCard, error) {
	jsonStr = strings.TrimSpace(jsonStr)

	// Try parsing as TMDb array format first
	var tmdbResults []tmdbMediaResult
	if err := json.Unmarshal([]byte(jsonStr), &tmdbResults); err == nil && len(tmdbResults) > 0 {
		cards := make([]MediaCard, 0, len(tmdbResults))
		for _, r := range tmdbResults {
			title := r.Title
			if title == "" {
				title = r.Name // Use Name for TV shows
			}
			cards = append(cards, MediaCard{
				ID:        r.ID,
				Title:     title,
				Year:      r.Year,
				MediaType: r.MediaType,
				Rating:    r.Rating,
				VoteCount: r.VoteCount,
				Overview:  r.Overview,
				Providers: r.Providers,
			})
		}
		return cards, nil
	}

	// Try parsing as AI recommendation format
	var aiResult aiRecommendationResult
	if err := json.Unmarshal([]byte(jsonStr), &aiResult); err == nil && len(aiResult.Recommendations) > 0 {
		cards := make([]MediaCard, 0, len(aiResult.Recommendations))
		for _, r := range aiResult.Recommendations {
			cards = append(cards, MediaCard{
				Title:     r.Title,
				Year:      r.Year,
				MediaType: r.MediaType,
				Rating:    r.Rating,
				Overview:  r.Overview,
				WhyWatch:  r.WhyWatch,
				Providers: r.Providers,
			})
		}
		return cards, nil
	}

	// Not a recognized format, return nil (not an error - just not media data)
	return nil, nil
}

// NewTextDisplayItem creates a DisplayItem for plain text
func NewTextDisplayItem(text string) DisplayItem {
	return DisplayItem{
		Type: DisplayItemText,
		Text: text,
	}
}

// NewCardsDisplayItem creates a DisplayItem for media cards
func NewCardsDisplayItem(cards []MediaCard, toolName string) DisplayItem {
	return DisplayItem{
		Type:       DisplayItemCards,
		MediaCards: cards,
		ToolName:   toolName,
	}
}
