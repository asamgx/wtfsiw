package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"wtfsiw/internal/config"
)

const baseURL = "https://api.themoviedb.org/3"

type Client struct {
	apiKey     string
	httpClient *http.Client
	region     string
	language   string
}

func NewClient() (*Client, error) {
	cfg := config.Get()
	if cfg.TMDB.APIKey == "" {
		return nil, fmt.Errorf("TMDb API key not configured. Set TMDB_API_KEY or run: wtfsiw config set tmdb.api_key YOUR_KEY")
	}

	return &Client{
		apiKey: cfg.TMDB.APIKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		region:   cfg.Preferences.Region,
		language: cfg.Preferences.Language,
	}, nil
}

func (c *Client) get(endpoint string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("api_key", c.apiKey)
	if c.language != "" {
		params.Set("language", c.language)
	}

	fullURL := fmt.Sprintf("%s%s?%s", baseURL, endpoint, params.Encode())

	resp, err := c.httpClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TMDb API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Media represents a movie or TV show
type Media struct {
	ID           int      `json:"id"`
	Title        string   `json:"title,omitempty"`        // for movies
	Name         string   `json:"name,omitempty"`         // for TV shows
	Overview     string   `json:"overview"`
	PosterPath   string   `json:"poster_path"`
	BackdropPath string   `json:"backdrop_path"`
	VoteAverage  float64  `json:"vote_average"`
	VoteCount    int      `json:"vote_count"`
	ReleaseDate  string   `json:"release_date,omitempty"`  // for movies
	FirstAirDate string   `json:"first_air_date,omitempty"` // for TV shows
	GenreIDs     []int    `json:"genre_ids"`
	MediaType    string   `json:"media_type,omitempty"`
	Popularity   float64  `json:"popularity"`
	Runtime      int      `json:"runtime,omitempty"` // only in detail view
	Providers    []Provider `json:"-"` // populated separately
}

// GetDisplayTitle returns the appropriate title based on media type
func (m *Media) GetDisplayTitle() string {
	if m.Title != "" {
		return m.Title
	}
	return m.Name
}

// GetDisplayYear returns the release year
func (m *Media) GetDisplayYear() string {
	date := m.ReleaseDate
	if date == "" {
		date = m.FirstAirDate
	}
	if len(date) >= 4 {
		return date[:4]
	}
	return ""
}

// Provider represents a streaming provider
type Provider struct {
	ID       int    `json:"provider_id"`
	Name     string `json:"provider_name"`
	LogoPath string `json:"logo_path"`
}

// SearchResponse represents the API response for search/discover
type SearchResponse struct {
	Page         int     `json:"page"`
	Results      []Media `json:"results"`
	TotalPages   int     `json:"total_pages"`
	TotalResults int     `json:"total_results"`
}

// Genre represents a genre
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// GenreMap maps genre names to IDs
var GenreMap = map[string]int{
	// Movie genres
	"action":          28,
	"adventure":       12,
	"animation":       16,
	"comedy":          35,
	"crime":           80,
	"documentary":     99,
	"drama":           18,
	"family":          10751,
	"fantasy":         14,
	"history":         36,
	"horror":          27,
	"music":           10402,
	"mystery":         9648,
	"romance":         10749,
	"sci-fi":          878,
	"science fiction": 878,
	"thriller":        53,
	"war":             10752,
	"western":         37,
	// TV genres (some overlap, some different IDs)
	"action & adventure": 10759,
	"kids":               10762,
	"news":               10763,
	"reality":            10764,
	"soap":               10766,
	"talk":               10767,
	"war & politics":     10768,
}

func (c *Client) parseSearchResponse(data []byte) (*SearchResponse, error) {
	var resp SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return &resp, nil
}
