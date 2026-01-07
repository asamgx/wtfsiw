package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"wtfsiw/internal/ai/tools"
	"wtfsiw/internal/tmdb"
	"wtfsiw/internal/trakt"
)

// ToolExecutor executes tool calls using the available clients
type ToolExecutor struct {
	tmdbClient  *tmdb.Client
	traktClient *trakt.Client
	aiProvider  Provider
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor(tmdbClient *tmdb.Client, traktClient *trakt.Client, aiProvider Provider) *ToolExecutor {
	return &ToolExecutor{
		tmdbClient:  tmdbClient,
		traktClient: traktClient,
		aiProvider:  aiProvider,
	}
}

// Execute runs a tool call and returns the result
func (e *ToolExecutor) Execute(ctx context.Context, call tools.ToolCall) tools.ToolResult {
	var content string
	var err error

	switch call.Name {
	case "search_media":
		content, err = e.searchMedia(ctx, call)
	case "get_media_details":
		content, err = e.getMediaDetails(ctx, call)
	case "get_streaming_providers":
		content, err = e.getStreamingProviders(ctx, call)
	case "get_similar":
		content, err = e.getSimilar(ctx, call)
	case "search_by_title":
		content, err = e.searchByTitle(ctx, call)
	case "get_trakt_watchlist":
		content, err = e.getTraktWatchlist(ctx, call)
	case "get_trakt_history":
		content, err = e.getTraktHistory(ctx, call)
	case "generate_recommendations":
		content, err = e.generateRecommendations(ctx, call)
	default:
		return tools.ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Unknown tool: %s", call.Name),
			IsError:    true,
		}
	}

	if err != nil {
		return tools.ToolResult{
			ToolCallID: call.ID,
			Content:    fmt.Sprintf("Error: %s", err.Error()),
			IsError:    true,
		}
	}

	return tools.ToolResult{
		ToolCallID: call.ID,
		Content:    content,
		IsError:    false,
	}
}

func (e *ToolExecutor) searchMedia(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.tmdbClient == nil {
		return "", fmt.Errorf("TMDb is not configured")
	}

	// Build search params from tool arguments
	params := &SearchParams{
		Keywords:       call.GetStringArray("keywords"),
		Genres:         call.GetStringArray("genres"),
		MediaType:      call.GetString("media_type"),
		YearFrom:       call.GetInt("year_from"),
		YearTo:         call.GetInt("year_to"),
		MinRating:      call.GetFloat("min_rating"),
		OriginalLang:   call.GetString("language"),
		WatchProviders: call.GetStringArray("providers"),
		Actors:         call.GetStringArray("actors"),
		Studios:        call.GetStringArray("studios"),
	}

	if params.MediaType == "" {
		params.MediaType = "all"
	}

	resp, err := e.tmdbClient.Discover(params)
	if err != nil {
		return "", err
	}

	// Enrich with providers
	e.tmdbClient.EnrichWithProviders(resp.Results)

	// Format results
	return formatMediaResults(resp.Results), nil
}

func (e *ToolExecutor) getMediaDetails(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.tmdbClient == nil {
		return "", fmt.Errorf("TMDb is not configured")
	}

	id := call.GetInt("id")
	mediaType := call.GetString("media_type")

	if id == 0 {
		return "", fmt.Errorf("id is required")
	}
	if mediaType == "" {
		return "", fmt.Errorf("media_type is required")
	}

	// Use search to get details (TMDb client doesn't have a dedicated details method yet)
	// For now, return basic info - could be enhanced later
	providers, _, err := e.tmdbClient.GetWatchProviders(mediaType, id)
	if err != nil {
		providers = nil
	}

	result := map[string]interface{}{
		"id":         id,
		"media_type": mediaType,
		"providers":  formatProviders(providers),
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

func (e *ToolExecutor) getStreamingProviders(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.tmdbClient == nil {
		return "", fmt.Errorf("TMDb is not configured")
	}

	id := call.GetInt("id")
	mediaType := call.GetString("media_type")

	if id == 0 {
		return "", fmt.Errorf("id is required")
	}
	if mediaType == "" {
		return "", fmt.Errorf("media_type is required")
	}

	providers, link, err := e.tmdbClient.GetWatchProviders(mediaType, id)
	if err != nil {
		return "", err
	}

	result := map[string]interface{}{
		"providers": formatProviders(providers),
		"link":      link,
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

func (e *ToolExecutor) getSimilar(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.tmdbClient == nil {
		return "", fmt.Errorf("TMDb is not configured")
	}

	id := call.GetInt("id")
	mediaType := call.GetString("media_type")

	if id == 0 {
		return "", fmt.Errorf("id is required")
	}
	if mediaType == "" {
		return "", fmt.Errorf("media_type is required")
	}

	// Use the existing findSimilar through a search
	params := &SearchParams{
		SimilarTo: []string{fmt.Sprintf("%d", id)}, // This won't work directly, need to enhance
		MediaType: mediaType,
	}

	resp, err := e.tmdbClient.Discover(params)
	if err != nil {
		return "", err
	}

	e.tmdbClient.EnrichWithProviders(resp.Results)
	return formatMediaResults(resp.Results), nil
}

func (e *ToolExecutor) searchByTitle(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.tmdbClient == nil {
		return "", fmt.Errorf("TMDb is not configured")
	}

	title := call.GetString("title")
	if title == "" {
		return "", fmt.Errorf("title is required")
	}

	resp, err := e.tmdbClient.Search(title)
	if err != nil {
		return "", err
	}

	// Limit to first 5 results
	results := resp.Results
	if len(results) > 5 {
		results = results[:5]
	}

	return formatMediaResults(results), nil
}

func (e *ToolExecutor) getTraktWatchlist(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.traktClient == nil {
		return "", fmt.Errorf("Trakt is not configured. Run 'wtfsiw trakt auth' to connect your account.")
	}

	mediaType := call.GetString("media_type")

	items, err := e.traktClient.GetWatchlist(mediaType)
	if err != nil {
		return "", err
	}

	// Format watchlist items
	var results []map[string]interface{}
	for _, item := range items {
		entry := map[string]interface{}{
			"type":     item.Type,
			"title":    item.GetDisplayTitle(),
			"year":     item.GetDisplayYear(),
			"rating":   item.GetRating(),
			"overview": truncateStr(item.GetOverview(), 200),
			"genres":   item.GetGenres(),
		}
		results = append(results, entry)
	}

	jsonBytes, _ := json.MarshalIndent(results, "", "  ")
	return string(jsonBytes), nil
}

func (e *ToolExecutor) getTraktHistory(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.traktClient == nil {
		return "", fmt.Errorf("Trakt is not configured. Run 'wtfsiw trakt auth' to connect your account.")
	}

	// History endpoint not yet implemented - return placeholder
	return `{"message": "Trakt history feature not yet implemented"}`, nil
}

func (e *ToolExecutor) generateRecommendations(ctx context.Context, call tools.ToolCall) (string, error) {
	if e.aiProvider == nil {
		return "", fmt.Errorf("AI provider is not configured")
	}

	description := call.GetString("description")
	count := call.GetInt("count")
	if count == 0 {
		count = 5
	}

	resp, err := e.aiProvider.GetRecommendations(ctx, description, count)
	if err != nil {
		return "", err
	}

	// Format recommendations
	var results []map[string]interface{}
	for _, rec := range resp.Recommendations {
		entry := map[string]interface{}{
			"title":      rec.Title,
			"year":       rec.Year,
			"media_type": rec.MediaType,
			"rating":     rec.Rating,
			"genres":     rec.Genres,
			"overview":   rec.Overview,
			"why_watch":  rec.WhyWatch,
		}
		results = append(results, entry)
	}

	result := map[string]interface{}{
		"summary":         resp.Summary,
		"recommendations": results,
	}

	jsonBytes, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonBytes), nil
}

// Helper functions

func formatMediaResults(results []tmdb.Media) string {
	var formatted []map[string]interface{}
	for _, m := range results {
		providers := make([]string, len(m.Providers))
		for i, p := range m.Providers {
			providers[i] = p.Name
		}

		entry := map[string]interface{}{
			"id":         m.ID,
			"title":      m.GetDisplayTitle(),
			"year":       m.GetDisplayYear(),
			"media_type": m.MediaType,
			"rating":     m.VoteAverage,
			"vote_count": m.VoteCount,
			"overview":   truncateStr(m.Overview, 200),
			"providers":  providers,
		}
		formatted = append(formatted, entry)
	}

	jsonBytes, _ := json.MarshalIndent(formatted, "", "  ")
	return string(jsonBytes)
}

func formatProviders(providers []tmdb.Provider) []string {
	names := make([]string, len(providers))
	for i, p := range providers {
		names[i] = p.Name
	}
	return names
}

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Find last space before maxLen
	s = s[:maxLen]
	if idx := strings.LastIndex(s, " "); idx > 0 {
		s = s[:idx]
	}
	return s + "..."
}
