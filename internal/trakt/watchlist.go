package trakt

import (
	"encoding/json"
	"fmt"
)

// WatchlistItem represents an item in the user's watchlist
type WatchlistItem struct {
	Rank     int    `json:"rank"`
	ListedAt string `json:"listed_at"`
	Type     string `json:"type"`
	Movie    *Movie `json:"movie,omitempty"`
	Show     *Show  `json:"show,omitempty"`
}

// Movie represents a movie in Trakt (extended=full fields included)
type Movie struct {
	Title                 string   `json:"title"`
	Year                  int      `json:"year"`
	IDs                   IDs      `json:"ids"`
	Tagline               string   `json:"tagline,omitempty"`
	Overview              string   `json:"overview,omitempty"`
	Released              string   `json:"released,omitempty"`
	Runtime               int      `json:"runtime,omitempty"`
	Country               string   `json:"country,omitempty"`
	Trailer               string   `json:"trailer,omitempty"`
	Homepage              string   `json:"homepage,omitempty"`
	Status                string   `json:"status,omitempty"`
	Rating                float64  `json:"rating,omitempty"`
	Votes                 int      `json:"votes,omitempty"`
	CommentCount          int      `json:"comment_count,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
	Language              string   `json:"language,omitempty"`
	Languages             []string `json:"languages,omitempty"`
	AvailableTranslations []string `json:"available_translations,omitempty"`
	Genres                []string `json:"genres,omitempty"`
	Certification         string   `json:"certification,omitempty"`
}

// Show represents a TV show in Trakt (extended=full fields included)
type Show struct {
	Title                 string   `json:"title"`
	Year                  int      `json:"year"`
	IDs                   IDs      `json:"ids"`
	Overview              string   `json:"overview,omitempty"`
	FirstAired            string   `json:"first_aired,omitempty"`
	Runtime               int      `json:"runtime,omitempty"`
	Certification         string   `json:"certification,omitempty"`
	Network               string   `json:"network,omitempty"`
	Country               string   `json:"country,omitempty"`
	Trailer               string   `json:"trailer,omitempty"`
	Homepage              string   `json:"homepage,omitempty"`
	Status                string   `json:"status,omitempty"`
	Rating                float64  `json:"rating,omitempty"`
	Votes                 int      `json:"votes,omitempty"`
	CommentCount          int      `json:"comment_count,omitempty"`
	UpdatedAt             string   `json:"updated_at,omitempty"`
	Language              string   `json:"language,omitempty"`
	Languages             []string `json:"languages,omitempty"`
	AvailableTranslations []string `json:"available_translations,omitempty"`
	Genres                []string `json:"genres,omitempty"`
	AiredEpisodes         int      `json:"aired_episodes,omitempty"`
}

// IDs contains various IDs for a media item
type IDs struct {
	Trakt int    `json:"trakt"`
	Slug  string `json:"slug"`
	IMDB  string `json:"imdb"`
	TMDB  int    `json:"tmdb"`
	TVDB  int    `json:"tvdb,omitempty"` // TV shows only
}

// GetDisplayTitle returns the title of the watchlist item
func (w *WatchlistItem) GetDisplayTitle() string {
	if w.Movie != nil {
		return w.Movie.Title
	}
	if w.Show != nil {
		return w.Show.Title
	}
	return ""
}

// GetDisplayYear returns the year of the watchlist item
func (w *WatchlistItem) GetDisplayYear() int {
	if w.Movie != nil {
		return w.Movie.Year
	}
	if w.Show != nil {
		return w.Show.Year
	}
	return 0
}

// GetOverview returns the overview of the watchlist item
func (w *WatchlistItem) GetOverview() string {
	if w.Movie != nil {
		return w.Movie.Overview
	}
	if w.Show != nil {
		return w.Show.Overview
	}
	return ""
}

// GetRating returns the Trakt rating of the watchlist item
func (w *WatchlistItem) GetRating() float64 {
	if w.Movie != nil {
		return w.Movie.Rating
	}
	if w.Show != nil {
		return w.Show.Rating
	}
	return 0
}

// GetGenres returns the genres of the watchlist item
func (w *WatchlistItem) GetGenres() []string {
	if w.Movie != nil {
		return w.Movie.Genres
	}
	if w.Show != nil {
		return w.Show.Genres
	}
	return nil
}

// GetRuntime returns the runtime in minutes
func (w *WatchlistItem) GetRuntime() int {
	if w.Movie != nil {
		return w.Movie.Runtime
	}
	if w.Show != nil {
		return w.Show.Runtime
	}
	return 0
}

// GetWatchlist returns items from the user's watchlist
// mediaType can be "movies", "shows", or empty for all items
func (c *Client) GetWatchlist(mediaType string) ([]WatchlistItem, error) {
	endpoint := "/users/me/watchlist"
	if mediaType != "" {
		endpoint += "/" + mediaType
	}
	// Add extended=full to get all available fields
	endpoint += "?extended=full"

	data, err := c.get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get watchlist: %w", err)
	}

	var items []WatchlistItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("failed to parse watchlist: %w", err)
	}

	return items, nil
}
