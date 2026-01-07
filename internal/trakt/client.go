package trakt

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"wtfsiw/internal/config"
)

const baseURL = "https://api.trakt.tv"

// Client handles Trakt API requests
type Client struct {
	clientID    string
	accessToken string
	httpClient  *http.Client
}

// NewClient creates a new Trakt API client
func NewClient() (*Client, error) {
	cfg := config.Get()
	if cfg.Trakt.ClientID == "" {
		return nil, fmt.Errorf("Trakt client ID not configured. Set TRAKT_CLIENT_ID or run: wtfsiw config set trakt.client_id YOUR_CLIENT_ID")
	}
	if cfg.Trakt.AccessToken == "" {
		return nil, fmt.Errorf("Trakt access token not configured. Run: wtfsiw trakt auth")
	}

	return &Client{
		clientID:    cfg.Trakt.ClientID,
		accessToken: cfg.Trakt.AccessToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// get performs an authenticated GET request to the Trakt API
func (c *Client) get(endpoint string) ([]byte, error) {
	fullURL := baseURL + endpoint

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required Trakt headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", c.clientID)
	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Trakt API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
