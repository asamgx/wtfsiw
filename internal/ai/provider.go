package ai

import (
	"context"
	"fmt"
	"time"

	"wtfsiw/internal/config"
	"wtfsiw/internal/tmdb"
)

// SearchParams is an alias for tmdb.SearchParams for backwards compatibility
type SearchParams = tmdb.SearchParams

// Recommendation represents a movie/TV show recommendation (unified format)
type Recommendation struct {
	Title       string   `json:"title"`
	Year        string   `json:"year"`
	MediaType   string   `json:"media_type"` // "movie" or "tv"
	Rating      float64  `json:"rating"`     // 0-10 scale
	Genres      []string `json:"genres"`
	Overview    string   `json:"overview"`
	WhyWatch    string   `json:"why_watch"`  // AI explanation of why this matches the query
	Providers   []string `json:"providers"`  // Streaming services (when known)
	VoteCount   int      `json:"vote_count"` // Number of votes (0 if from AI)
	FromAI      bool     `json:"-"`          // True if recommendation came directly from AI
}

// RecommendationResponse is the structured output from the AI
type RecommendationResponse struct {
	Recommendations []Recommendation `json:"recommendations"`
	Summary         string           `json:"summary"` // Brief summary of what was searched for
}

// Provider defines the interface for AI providers
type Provider interface {
	ExtractSearchParams(ctx context.Context, query string) (*SearchParams, error)
	GetRecommendations(ctx context.Context, query string, count int) (*RecommendationResponse, error)
}

// NewProvider creates a new AI provider based on config
func NewProvider() (Provider, error) {
	cfg := config.Get()

	switch cfg.AI.Provider {
	case "claude":
		if cfg.AI.ClaudeAPIKey == "" {
			return nil, fmt.Errorf("Claude API key not configured. Set ANTHROPIC_API_KEY or run: wtfsiw config set ai.claude_api_key YOUR_KEY")
		}
		return NewClaudeProvider(cfg.AI.ClaudeAPIKey), nil
	case "openai":
		if cfg.AI.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OpenAI API key not configured. Set OPENAI_API_KEY or run: wtfsiw config set ai.openai_api_key YOUR_KEY")
		}
		return NewOpenAIProvider(cfg.AI.OpenAIAPIKey), nil
	default:
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.AI.Provider)
	}
}

// getSystemPromptExtract returns the extraction prompt with current date
func getSystemPromptExtract() string {
	now := time.Now()
	currentYear := now.Year()
	currentDate := now.Format("January 2, 2006")

	return fmt.Sprintf(`You are a movie/TV show search assistant. Extract structured search parameters from natural language queries.

Today's date: %s (current year: %d)

Parameters to extract:

CORE SEARCH:
- keywords: search terms (array of strings, default: [])
- genres: genres like action, comedy, drama, horror, thriller, sci-fi, romance, documentary, animation, fantasy, mystery, crime, war, western, family, history, music (array, default: [])
- similar_to: reference titles mentioned (array, default: [])
- media_type: "movie", "tv", or "all" (default: "all")

DATE/YEAR:
- year_from: start year (integer, default: 0)
- year_to: end year (integer, default: 0)
  Examples: "recent" = %d-%d, "last 5 years" = %d-%d, "90s" = 1990-1999, "2010s" = 2010-2019

RATINGS:
- min_rating: minimum rating 0-10 (number, default: 0). "highly rated" = 7.5+, "critically acclaimed" = 8+
- min_vote_count: minimum votes for quality (integer, default: 0). "well-known" = 1000+, "popular" = 5000+

RUNTIME:
- max_runtime: max minutes (integer, default: 0). "short" = 90, "quick watch" = 100

LANGUAGE:
- original_language: ISO 639-1 code (string, default: ""). Examples: "en", "ko" (Korean), "ja" (Japanese), "fr", "es", "de", "it", "zh" (Chinese), "hi" (Hindi)

PEOPLE/STUDIOS:
- actors: actor names mentioned (array, default: [])
- directors: director names mentioned (array, default: [])
- studios: production companies (array, default: []). Examples: "Pixar", "A24", "Marvel", "DC", "Disney", "Warner Bros", "Universal", "Paramount", "Sony", "Lionsgate", "Blumhouse", "Studio Ghibli"

STREAMING:
- watch_providers: streaming services (array, default: []). Examples: "Netflix", "Amazon Prime Video", "Disney Plus", "HBO Max", "Hulu", "Apple TV Plus", "Paramount Plus", "Peacock"
- monetization_type: "flatrate" (subscription), "free", "rent", "buy" (string, default: "")

CONTENT RATING:
- certification: "G", "PG", "PG-13", "R", "NC-17" for movies; "TV-Y", "TV-G", "TV-PG", "TV-14", "TV-MA" for TV (string, default: "")

TV-SPECIFIC:
- tv_status: "returning" (still airing), "ended", "canceled" (string, default: "")

SORTING:
- sort_by: "popularity", "rating", "release_date", "revenue" (string, default: "")

MOOD (for AI interpretation, not TMDb filter):
- mood: overall tone like "dark", "fun", "thought-provoking", "feel-good", "intense", "relaxing" (string, default: "")

IMPORTANT: For ALL numeric fields, use 0 as default, NOT empty strings.

Respond with ONLY valid JSON, no markdown. Example:
{"keywords":["heist"],"genres":["thriller","crime"],"similar_to":["Ocean's Eleven"],"media_type":"movie","year_from":0,"year_to":0,"min_rating":7.5,"min_vote_count":1000,"max_runtime":0,"original_language":"","actors":[],"directors":["Steven Soderbergh"],"studios":[],"watch_providers":["Netflix"],"monetization_type":"flatrate","certification":"","tv_status":"","sort_by":"rating","mood":"fun"}`,
		currentDate, currentYear,
		currentYear-2, currentYear, // "recent"
		currentYear-5, currentYear) // "last 5 years"
}

const systemPromptRecommend = `You are an expert movie and TV show recommender. Given a user's description of what they want to watch, provide personalized recommendations.

For each recommendation include:
- title: The exact title of the movie or TV show
- year: Release year (e.g., "2019" or "2019-2023" for TV shows)
- media_type: Either "movie" or "tv"
- rating: Your estimated rating out of 10 based on critical consensus and audience reception
- genres: Array of genres that apply
- overview: A brief 1-2 sentence description (no spoilers)
- why_watch: A personalized explanation of why this matches what the user is looking for
- providers: Common streaming services where it's typically available (Netflix, Prime Video, HBO Max, Disney+, Hulu, Apple TV+, etc.) - leave empty if unsure

Respond with ONLY a valid JSON object in this exact format:
{
  "summary": "Brief description of what you searched for",
  "recommendations": [
    {
      "title": "Breaking Bad",
      "year": "2008-2013",
      "media_type": "tv",
      "rating": 9.5,
      "genres": ["drama", "crime", "thriller"],
      "overview": "A high school chemistry teacher turned methamphetamine manufacturer partners with a former student.",
      "why_watch": "Matches your request for dark, psychological content with morally complex characters.",
      "providers": ["Netflix"]
    }
  ]
}`
