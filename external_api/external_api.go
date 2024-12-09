package externalapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"online-library/internal/logger"
	"online-library/internal/models"
	"strings"
)

type ExternalAPI interface {
	GetSongDetails(group string, song string) (*models.SongDetail, error)
}

type ExternalAPIClient struct {
	fullURL string
	Method  string
}

func NewExternalAPIClient(URL, Method string) *ExternalAPIClient {
	Method = strings.ToUpper(Method)
	switch Method {
	case http.MethodGet, http.MethodPatch, http.MethodPost:
	default:
		Method = http.MethodGet
	}
	return &ExternalAPIClient{
		fullURL: URL,
		Method:  Method,
	}
}

func (c *ExternalAPIClient) GetSongDetails(group string, song string) (*models.SongDetail, error) {

	req, err := http.NewRequest(c.Method, c.fullURL, nil)
	if err != nil {
		logger.Log.Errorf("Failed to create request: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Log.Errorf("Failed tto send request: %v", err)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var songDetail models.SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&songDetail); err != nil {
		logger.Log.Errorf("Failed to parse response: %v", err)
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &songDetail, nil
}
