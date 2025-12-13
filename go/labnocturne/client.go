// Package labnocturne provides a Go client for the Lab Nocturne Images API
package labnocturne

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const DefaultBaseURL = "https://images.labnocturne.com"

// Client represents a Lab Nocturne Images API client
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient creates a new API client
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{},
	}
}

// UploadResponse represents the response from uploading a file
type UploadResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	CreatedAt string `json:"created_at"`
}

// FileInfo represents information about an uploaded file
type FileInfo struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	CreatedAt string `json:"created_at"`
}

// ListFilesResponse represents the response from listing files
type ListFilesResponse struct {
	Files      []FileInfo `json:"files"`
	Pagination struct {
		Page       int `json:"page"`
		Limit      int `json:"limit"`
		Total      int `json:"total"`
		TotalPages int `json:"total_pages"`
	} `json:"pagination"`
}

// StatsResponse represents usage statistics
type StatsResponse struct {
	StorageUsedBytes int64   `json:"storage_used_bytes"`
	StorageUsedMB    float64 `json:"storage_used_mb"`
	FileCount        int     `json:"file_count"`
	QuotaBytes       int64   `json:"quota_bytes"`
	QuotaMB          float64 `json:"quota_mb"`
	UsagePercent     float64 `json:"usage_percent"`
}

// APIError represents an API error response
type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Upload uploads a file to the API
func (c *Client) Upload(filePath string) (*UploadResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/upload", body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil {
			return nil, fmt.Errorf("upload failed: %s", apiErr.Error.Message)
		}
		return nil, fmt.Errorf("upload failed with status: %s", resp.Status)
	}

	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ListFiles lists uploaded files with pagination
func (c *Client) ListFiles(page, limit int, sort string) (*ListFilesResponse, error) {
	url := fmt.Sprintf("%s/files?page=%d&limit=%d&sort=%s", c.BaseURL, page, limit, sort)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list failed with status: %s", resp.Status)
	}

	var result ListFilesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetStats retrieves usage statistics
func (c *Client) GetStats() (*StatsResponse, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/stats", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stats request failed with status: %s", resp.Status)
	}

	var result StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteFile deletes an image (soft delete)
func (c *Client) DeleteFile(imageID string) error {
	req, err := http.NewRequest("DELETE", c.BaseURL+"/i/"+imageID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed with status: %s", resp.Status)
	}

	return nil
}

// GenerateTestKey generates a test API key
func GenerateTestKey() (string, error) {
	resp, err := http.Get(DefaultBaseURL + "/key")
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("key generation failed with status: %s", resp.Status)
	}

	var result struct {
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.APIKey, nil
}
