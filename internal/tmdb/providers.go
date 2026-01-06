package tmdb

import (
	"encoding/json"
	"fmt"
)

// WatchProvidersResponse represents the watch providers API response
type WatchProvidersResponse struct {
	ID      int                        `json:"id"`
	Results map[string]CountryProvider `json:"results"`
}

// CountryProvider represents providers available in a specific country
type CountryProvider struct {
	Link     string     `json:"link"`
	Flatrate []Provider `json:"flatrate"` // Subscription streaming
	Rent     []Provider `json:"rent"`     // Rent
	Buy      []Provider `json:"buy"`      // Buy
	Free     []Provider `json:"free"`     // Free with ads
}

// GetWatchProviders fetches streaming providers for a movie or TV show
func (c *Client) GetWatchProviders(mediaType string, id int) ([]Provider, string, error) {
	endpoint := fmt.Sprintf("/%s/%d/watch/providers", mediaType, id)

	data, err := c.get(endpoint, nil)
	if err != nil {
		return nil, "", err
	}

	var resp WatchProvidersResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("failed to parse providers response: %w", err)
	}

	// Get providers for the configured region
	region := c.region
	if region == "" {
		region = "US"
	}

	countryProviders, ok := resp.Results[region]
	if !ok {
		return nil, "", nil // No providers in this region
	}

	// Combine all provider types, prioritizing flatrate (streaming)
	var providers []Provider
	seen := make(map[int]bool)

	addProviders := func(list []Provider) {
		for _, p := range list {
			if !seen[p.ID] {
				seen[p.ID] = true
				providers = append(providers, p)
			}
		}
	}

	addProviders(countryProviders.Flatrate)
	addProviders(countryProviders.Free)
	addProviders(countryProviders.Rent)
	addProviders(countryProviders.Buy)

	return providers, countryProviders.Link, nil
}

// EnrichWithProviders adds streaming provider info to media items
func (c *Client) EnrichWithProviders(results []Media) {
	for i := range results {
		mediaType := results[i].MediaType
		if mediaType == "" {
			// Try to determine from available data
			if results[i].Title != "" {
				mediaType = "movie"
			} else {
				mediaType = "tv"
			}
		}

		providers, _, err := c.GetWatchProviders(mediaType, results[i].ID)
		if err == nil {
			results[i].Providers = providers
		}
	}
}

// ProviderEmoji returns an emoji for common streaming providers
func ProviderEmoji(name string) string {
	switch name {
	case "Netflix":
		return "N"
	case "Amazon Prime Video", "Prime Video":
		return "P"
	case "Disney Plus":
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
