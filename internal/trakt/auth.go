package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DeviceCodeResponse represents the response from /oauth/device/code
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse represents the response from /oauth/device/token
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	CreatedAt    int    `json:"created_at"`
}

// GetDeviceCode requests a device code for OAuth authentication
func GetDeviceCode(clientID string) (*DeviceCodeResponse, error) {
	payload := map[string]string{
		"client_id": clientID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", baseURL+"/oauth/device/code", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Trakt API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result DeviceCodeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// PollForToken polls the token endpoint until the user authorizes or timeout
func PollForToken(clientID, clientSecret, deviceCode string, interval int) (*TokenResponse, error) {
	payload := map[string]string{
		"code":          deviceCode,
		"client_id":     clientID,
		"client_secret": clientSecret,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	for {
		req, err := http.NewRequest("POST", baseURL+"/oauth/device/token", bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("HTTP request failed: %w", err)
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		switch resp.StatusCode {
		case http.StatusOK:
			// Success - user authorized
			var result TokenResponse
			if err := json.Unmarshal(respBody, &result); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}
			return &result, nil

		case http.StatusBadRequest:
			// 400 - Pending authorization, keep polling
			time.Sleep(time.Duration(interval) * time.Second)
			continue

		case http.StatusNotFound:
			// 404 - Invalid device code
			return nil, fmt.Errorf("invalid device code")

		case http.StatusConflict:
			// 409 - Code already used
			return nil, fmt.Errorf("device code already used")

		case http.StatusGone:
			// 410 - Code expired
			return nil, fmt.Errorf("device code expired")

		case http.StatusTeapot:
			// 418 - User denied authorization
			return nil, fmt.Errorf("user denied authorization")

		case http.StatusTooManyRequests:
			// 429 - Polling too fast
			time.Sleep(time.Duration(interval*2) * time.Second)
			continue

		default:
			return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
		}
	}
}
