package ai

import (
	"context"
	"fmt"

	"wtfsiw/internal/config"
)

// SearchParams represents the extracted search parameters from a natural language query
type SearchParams struct {
	Keywords     []string `json:"keywords"`
	Genres       []string `json:"genres"`
	SimilarTo    []string `json:"similar_to"`
	MediaType    string   `json:"media_type"` // movie, tv, or all
	YearFrom     int      `json:"year_from,omitempty"`
	YearTo       int      `json:"year_to,omitempty"`
	MinRating    float64  `json:"min_rating,omitempty"`
	Mood         string   `json:"mood,omitempty"`
	MaxRuntime   int      `json:"max_runtime,omitempty"` // in minutes
	OriginalLang string   `json:"original_language,omitempty"`
}

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

const systemPromptExtract = `You are a movie/TV show search assistant. Your job is to extract structured search parameters from natural language queries.

Given a user's description of what they want to watch, extract the following parameters:
- keywords: relevant search terms
- genres: movie/TV genres (action, comedy, drama, thriller, sci-fi, horror, romance, documentary, animation, etc.)
- similar_to: titles the user mentioned as reference points
- media_type: "movie", "tv", or "all" based on what they're looking for
- year_from/year_to: year range if mentioned (e.g., "recent" = last 2-3 years, "80s" = 1980-1989)
- min_rating: minimum rating threshold if quality is emphasized
- mood: overall mood/tone (dark, light, intense, relaxing, thought-provoking, etc.)
- max_runtime: if they mention wanting something short/long (short = ~90 min, long = 150+ min)
- original_language: if they specify (e.g., "Korean" = "ko", "Japanese" = "ja")

Respond with ONLY a valid JSON object, no markdown, no explanation. Example:
{"keywords":["heist","clever"],"genres":["thriller","crime"],"similar_to":["Ocean's Eleven"],"media_type":"movie","mood":"fun","min_rating":7.0}`

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
