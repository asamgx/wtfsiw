package tmdb

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Search performs a multi-search for movies and TV shows
func (c *Client) Search(query string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("include_adult", "false")

	data, err := c.get("/search/multi", params)
	if err != nil {
		return nil, err
	}

	resp, err := c.parseSearchResponse(data)
	if err != nil {
		return nil, err
	}

	// Filter to only movies and TV shows
	filtered := make([]Media, 0)
	for _, m := range resp.Results {
		if m.MediaType == "movie" || m.MediaType == "tv" {
			filtered = append(filtered, m)
		}
	}
	resp.Results = filtered

	return resp, nil
}

// Discover finds movies/TV shows based on structured parameters
func (c *Client) Discover(searchParams *SearchParams) (*SearchResponse, error) {
	var allResults []Media

	// Determine which endpoints to query
	endpoints := []string{}
	switch searchParams.MediaType {
	case "movie":
		endpoints = []string{"/discover/movie"}
	case "tv":
		endpoints = []string{"/discover/tv"}
	default:
		endpoints = []string{"/discover/movie", "/discover/tv"}
	}

	for _, endpoint := range endpoints {
		params := c.buildDiscoverParams(searchParams, endpoint)
		data, err := c.get(endpoint, params)
		if err != nil {
			continue // Try other endpoints on error
		}

		resp, err := c.parseSearchResponse(data)
		if err != nil {
			continue
		}

		// Set media type based on endpoint
		mediaType := "movie"
		if strings.Contains(endpoint, "/tv") {
			mediaType = "tv"
		}
		for i := range resp.Results {
			resp.Results[i].MediaType = mediaType
		}

		allResults = append(allResults, resp.Results...)
	}

	// If we have similar_to references, also search for those
	if len(searchParams.SimilarTo) > 0 {
		similarResults := c.findSimilar(searchParams.SimilarTo, searchParams.MediaType)
		allResults = append(allResults, similarResults...)
	}

	// If we have keywords, also do a keyword search
	if len(searchParams.Keywords) > 0 {
		keywordQuery := strings.Join(searchParams.Keywords, " ")
		searchResp, err := c.Search(keywordQuery)
		if err == nil {
			allResults = append(allResults, searchResp.Results...)
		}
	}

	// Deduplicate and sort by relevance (vote_average * log(vote_count))
	allResults = deduplicateAndSort(allResults, searchParams.MinRating)

	// Limit results
	maxResults := 10
	if len(allResults) > maxResults {
		allResults = allResults[:maxResults]
	}

	return &SearchResponse{
		Page:         1,
		Results:      allResults,
		TotalResults: len(allResults),
		TotalPages:   1,
	}, nil
}

func (c *Client) buildDiscoverParams(sp *SearchParams, endpoint string) url.Values {
	params := url.Values{}
	isMovie := strings.Contains(endpoint, "/movie")

	// Sorting
	sortBy := "vote_average.desc" // default
	if sp.SortBy != "" {
		if mapped, ok := SortByMap[strings.ToLower(sp.SortBy)]; ok {
			sortBy = mapped
		}
	}
	params.Set("sort_by", sortBy)

	// Vote count filtering (quality control)
	minVotes := 100 // default minimum
	if sp.MinVoteCount > 0 {
		minVotes = sp.MinVoteCount
	}
	params.Set("vote_count.gte", strconv.Itoa(minVotes))

	// Genre filtering
	if len(sp.Genres) > 0 {
		genreIDs := []string{}
		for _, genre := range sp.Genres {
			if id, ok := GenreMap[strings.ToLower(genre)]; ok {
				genreIDs = append(genreIDs, strconv.Itoa(id))
			}
		}
		if len(genreIDs) > 0 {
			params.Set("with_genres", strings.Join(genreIDs, ","))
		}
	}

	// Year filtering
	if sp.YearFrom > 0 {
		if isMovie {
			params.Set("primary_release_date.gte", fmt.Sprintf("%d-01-01", sp.YearFrom))
		} else {
			params.Set("first_air_date.gte", fmt.Sprintf("%d-01-01", sp.YearFrom))
		}
	}
	if sp.YearTo > 0 {
		if isMovie {
			params.Set("primary_release_date.lte", fmt.Sprintf("%d-12-31", sp.YearTo))
		} else {
			params.Set("first_air_date.lte", fmt.Sprintf("%d-12-31", sp.YearTo))
		}
	}

	// Rating filtering
	if sp.MinRating > 0 {
		params.Set("vote_average.gte", fmt.Sprintf("%.1f", sp.MinRating))
	}

	// Runtime filtering
	if sp.MaxRuntime > 0 {
		params.Set("with_runtime.lte", strconv.Itoa(sp.MaxRuntime))
	}

	// Language filtering
	if sp.OriginalLang != "" {
		params.Set("with_original_language", sp.OriginalLang)
	}

	// Studio/Company filtering
	if len(sp.Studios) > 0 {
		companyIDs := []string{}
		for _, studio := range sp.Studios {
			if id, ok := StudioMap[strings.ToLower(studio)]; ok {
				companyIDs = append(companyIDs, strconv.Itoa(id))
			}
		}
		if len(companyIDs) > 0 {
			params.Set("with_companies", strings.Join(companyIDs, "|")) // OR logic
		}
	}

	// Actor/People filtering
	if len(sp.Actors) > 0 || len(sp.Directors) > 0 {
		peopleIDs := []string{}
		allPeople := append(sp.Actors, sp.Directors...)
		for _, person := range allPeople {
			if id := c.searchPersonID(person); id > 0 {
				peopleIDs = append(peopleIDs, strconv.Itoa(id))
			}
		}
		if len(peopleIDs) > 0 {
			if isMovie {
				// For movies, use with_people (cast or crew)
				params.Set("with_people", strings.Join(peopleIDs, ",")) // AND logic
			}
			// Note: TV discover doesn't support with_people directly
		}
	}

	// Watch provider filtering
	if len(sp.WatchProviders) > 0 {
		providerIDs := []string{}
		for _, provider := range sp.WatchProviders {
			if id, ok := WatchProviderMap[strings.ToLower(provider)]; ok {
				providerIDs = append(providerIDs, strconv.Itoa(id))
			}
		}
		if len(providerIDs) > 0 {
			params.Set("with_watch_providers", strings.Join(providerIDs, "|")) // OR logic
		}
	}

	// Monetization type
	if sp.MonetizationType != "" {
		if mapped, ok := MonetizationTypeMap[strings.ToLower(sp.MonetizationType)]; ok {
			params.Set("with_watch_monetization_types", mapped)
		}
	}

	// Certification filtering
	if sp.Certification != "" {
		cert := strings.ToUpper(sp.Certification)
		if mapped, ok := CertificationMap[strings.ToLower(sp.Certification)]; ok {
			cert = mapped
		}
		params.Set("certification_country", "US")
		params.Set("certification", cert)
	}

	// TV Status filtering
	if sp.TVStatus != "" && !isMovie {
		if status, ok := TVStatusMap[strings.ToLower(sp.TVStatus)]; ok {
			params.Set("with_status", strconv.Itoa(status))
		}
	}

	// Region for watch providers
	region := c.region
	if sp.AvailableInRegion != "" {
		region = sp.AvailableInRegion
	}
	if region != "" {
		params.Set("watch_region", region)
	}

	return params
}

// searchPersonID searches for a person by name and returns their TMDb ID
func (c *Client) searchPersonID(name string) int {
	params := url.Values{}
	params.Set("query", name)

	data, err := c.get("/search/person", params)
	if err != nil {
		return 0
	}

	var resp struct {
		Results []struct {
			ID int `json:"id"`
		} `json:"results"`
	}

	if err := json.Unmarshal(data, &resp); err != nil || len(resp.Results) == 0 {
		return 0
	}

	return resp.Results[0].ID
}

func (c *Client) findSimilar(titles []string, mediaType string) []Media {
	var results []Media

	for _, title := range titles {
		// First search for the title to get its ID
		searchResp, err := c.Search(title)
		if err != nil || len(searchResp.Results) == 0 {
			continue
		}

		// Get the first result's ID
		first := searchResp.Results[0]

		// Fetch similar titles
		var endpoint string
		if first.MediaType == "movie" {
			endpoint = fmt.Sprintf("/movie/%d/similar", first.ID)
		} else if first.MediaType == "tv" {
			endpoint = fmt.Sprintf("/tv/%d/similar", first.ID)
		} else {
			continue
		}

		data, err := c.get(endpoint, nil)
		if err != nil {
			continue
		}

		resp, err := c.parseSearchResponse(data)
		if err != nil {
			continue
		}

		// Set media type
		for i := range resp.Results {
			resp.Results[i].MediaType = first.MediaType
		}

		results = append(results, resp.Results...)
	}

	return results
}

func deduplicateAndSort(results []Media, minRating float64) []Media {
	seen := make(map[string]bool)
	unique := make([]Media, 0)

	for _, r := range results {
		key := fmt.Sprintf("%s-%d", r.MediaType, r.ID)
		if seen[key] {
			continue
		}
		if minRating > 0 && r.VoteAverage < minRating {
			continue
		}
		seen[key] = true
		unique = append(unique, r)
	}

	// Sort by score (vote_average weighted by popularity)
	for i := 0; i < len(unique)-1; i++ {
		for j := i + 1; j < len(unique); j++ {
			scoreI := unique[i].VoteAverage * (1 + unique[i].Popularity/100)
			scoreJ := unique[j].VoteAverage * (1 + unique[j].Popularity/100)
			if scoreJ > scoreI {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}

	return unique
}
