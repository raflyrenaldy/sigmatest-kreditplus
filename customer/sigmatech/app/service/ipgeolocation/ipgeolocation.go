// Package ipgeolocation provides functions to interact with the ipgeolocation.io API.
package ipgeolocation

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Client is the main struct for interacting with the ipgeolocation.io API.
type Client struct {
	BaseURL string
	APIKey  string
}

// NewClient creates a new ipgeolocation.io API client with the provided API key.
func NewClient(apiKey string) *Client {
	return &Client{
		BaseURL: BaseURL,
		APIKey:  apiKey,
	}
}

// ErrorResponse represents an error response from the API.
type ErrorResponse struct {
	Message string `json:"message"`
}

// GetLocationInfo retrieves location information for a given IP address.
func (c *Client) GetLocationInfo(ipAddress string) (LocationInfo, error) {
	url := fmt.Sprintf("%s%s?%s=%s&%s=%s&%s=%s", c.BaseURL, GetLocationInfoEndpoint, APIKeyParam, c.APIKey, IPAddressParam, ipAddress, LanguageParam, "en")

	resp, err := http.Get(url)
	if err != nil {
		return LocationInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
			return LocationInfo{}, err
		}
		return LocationInfo{}, fmt.Errorf("error response from API: %s", errorResponse.Message)
	}

	var locationInfo LocationInfo
	if err := json.NewDecoder(resp.Body).Decode(&locationInfo); err != nil {
		return LocationInfo{}, err
	}

	return locationInfo, nil
}
