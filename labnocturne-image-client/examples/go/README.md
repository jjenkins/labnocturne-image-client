# Go Examples - Lab Nocturne Images API

Complete Go client examples for integrating with the Lab Nocturne Images API.

## Prerequisites

- Go 1.21+

## Installation

No external dependencies required! Uses standard library.

## Quick Start

### Basic Upload

```go
// main.go
package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "mime/multipart"
    "net/http"
    "os"
)

const APIBase = "https://images.labnocturne.com"

type UploadResponse struct {
    ID        string `json:"id"`
    URL       string `json:"url"`
    Size      int64  `json:"size"`
    MimeType  string `json:"mime_type"`
    CreatedAt string `json:"created_at"`
}

func uploadImage(apiKey, filePath string) (*UploadResponse, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    part, err := writer.CreateFormFile("file", filePath)
    if err != nil {
        return nil, err
    }

    if _, err := io.Copy(part, file); err != nil {
        return nil, err
    }

    writer.Close()

    req, err := http.NewRequest("POST", APIBase+"/upload", body)
    if err != nil {
        return nil, err
    }

    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("Authorization", "Bearer "+apiKey)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("upload failed: %s", resp.Status)
    }

    var result UploadResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

func main() {
    apiKey := "ln_test_01jcd8x9k2..."
    result, err := uploadImage(apiKey, "photo.jpg")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Image URL:", result.URL)
    fmt.Println("Image ID:", result.ID)
}
```

## Complete Client Package

```go
// client.go
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
```

## Complete Example

```go
// example/main.go
package main

import (
    "fmt"
    "log"

    "github.com/jjenkins/labnocturne-image-client/go/labnocturne"
)

func main() {
    // Generate a test API key
    fmt.Println("Generating test API key...")
    apiKey, err := labnocturne.GenerateTestKey()
    if err != nil {
        log.Fatal("Failed to generate API key:", err)
    }
    fmt.Println("API Key:", apiKey)
    fmt.Println()

    // Create client
    client := labnocturne.NewClient(apiKey)

    // Upload an image
    fmt.Println("Uploading image...")
    upload, err := client.Upload("photo.jpg")
    if err != nil {
        log.Fatal("Upload failed:", err)
    }
    fmt.Println("Uploaded:", upload.URL)
    fmt.Println("Image ID:", upload.ID)
    fmt.Printf("Size: %.2f KB\n", float64(upload.Size)/1024)
    fmt.Println()

    // List all files
    fmt.Println("Listing files...")
    files, err := client.ListFiles(1, 10, "created_desc")
    if err != nil {
        log.Fatal("List failed:", err)
    }
    fmt.Printf("Total files: %d\n", files.Pagination.Total)
    for _, file := range files.Files {
        fmt.Printf("  - %s: %.2f KB\n", file.ID, float64(file.Size)/1024)
    }
    fmt.Println()

    // Get usage stats
    fmt.Println("Usage statistics:")
    stats, err := client.GetStats()
    if err != nil {
        log.Fatal("Stats failed:", err)
    }
    fmt.Printf("  Storage: %.2f MB / %.0f MB\n", stats.StorageUsedMB, stats.QuotaMB)
    fmt.Printf("  Files: %d\n", stats.FileCount)
    fmt.Printf("  Usage: %.2f%%\n", stats.UsagePercent)
    fmt.Println()

    // Delete the uploaded file
    fmt.Println("Deleting image...")
    if err := client.DeleteFile(upload.ID); err != nil {
        log.Fatal("Delete failed:", err)
    }
    fmt.Println("Deleted successfully")
}
```

## Batch Upload Example

```go
// batch_upload.go
package main

import (
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/jjenkins/labnocturne-image-client/go/labnocturne"
)

func main() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: go run batch_upload.go <api_key> <directory>")
        os.Exit(1)
    }

    apiKey := os.Args[1]
    directory := os.Args[2]

    client := labnocturne.NewClient(apiKey)

    imageExtensions := map[string]bool{
        ".jpg":  true,
        ".jpeg": true,
        ".png":  true,
        ".gif":  true,
        ".webp": true,
    }

    var uploaded []string

    err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if info.IsDir() {
            return nil
        }

        ext := strings.ToLower(filepath.Ext(path))
        if !imageExtensions[ext] {
            return nil
        }

        fmt.Printf("Uploading %s...\n", filepath.Base(path))
        result, err := client.Upload(path)
        if err != nil {
            fmt.Printf("  ✗ Failed: %v\n", err)
            return nil
        }

        fmt.Printf("  ✓ %s\n", result.URL)
        uploaded = append(uploaded, result.ID)

        return nil
    })

    if err != nil {
        log.Fatal("Walk failed:", err)
    }

    fmt.Printf("\nUploaded %d files\n", len(uploaded))

    stats, err := client.GetStats()
    if err != nil {
        log.Fatal("Stats failed:", err)
    }
    fmt.Printf("Total storage: %.2f MB\n", stats.StorageUsedMB)
}
```

## CLI Tool

```go
// cmd/ln-images/main.go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/jjenkins/labnocturne-image-client/go/labnocturne"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    command := os.Args[1]

    switch command {
    case "key":
        cmdGenerateKey()
    case "upload":
        cmdUpload()
    case "list":
        cmdList()
    case "stats":
        cmdStats()
    case "delete":
        cmdDelete()
    default:
        printUsage()
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("Lab Nocturne Images CLI")
    fmt.Println()
    fmt.Println("Usage:")
    fmt.Println("  ln-images key                    Generate a test API key")
    fmt.Println("  ln-images upload <file>          Upload an image")
    fmt.Println("  ln-images list                   List files")
    fmt.Println("  ln-images stats                  Show usage statistics")
    fmt.Println("  ln-images delete <image_id>      Delete an image")
    fmt.Println()
    fmt.Println("Environment Variables:")
    fmt.Println("  LABNOCTURNE_API_KEY    API key for authentication")
}

func cmdGenerateKey() {
    apiKey, err := labnocturne.GenerateTestKey()
    if err != nil {
        log.Fatal("Failed to generate key:", err)
    }

    fmt.Println("Generated API key:", apiKey)
    fmt.Println()
    fmt.Println("Save this key:")
    fmt.Printf("  export LABNOCTURNE_API_KEY='%s'\n", apiKey)
}

func cmdUpload() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: ln-images upload <file>")
        os.Exit(1)
    }

    apiKey := getAPIKey()
    filePath := os.Args[2]

    client := labnocturne.NewClient(apiKey)
    result, err := client.Upload(filePath)
    if err != nil {
        log.Fatal("Upload failed:", err)
    }

    fmt.Println("Uploaded:", result.URL)
    fmt.Println("ID:", result.ID)
    fmt.Printf("Size: %.2f KB\n", float64(result.Size)/1024)
}

func cmdList() {
    apiKey := getAPIKey()
    client := labnocturne.NewClient(apiKey)

    files, err := client.ListFiles(1, 50, "created_desc")
    if err != nil {
        log.Fatal("List failed:", err)
    }

    fmt.Printf("Files (page %d of %d):\n", files.Pagination.Page, files.Pagination.TotalPages)
    for _, file := range files.Files {
        fmt.Printf("\n%s\n", file.ID)
        fmt.Printf("  URL: %s\n", file.URL)
        fmt.Printf("  Size: %.2f KB\n", float64(file.Size)/1024)
        fmt.Printf("  Created: %s\n", file.CreatedAt)
    }
}

func cmdStats() {
    apiKey := getAPIKey()
    client := labnocturne.NewClient(apiKey)

    stats, err := client.GetStats()
    if err != nil {
        log.Fatal("Stats failed:", err)
    }

    fmt.Println("Usage Statistics:")
    fmt.Printf("  Storage: %.2f MB / %.0f MB\n", stats.StorageUsedMB, stats.QuotaMB)
    fmt.Printf("  Files: %d\n", stats.FileCount)
    fmt.Printf("  Usage: %.2f%%\n", stats.UsagePercent)
}

func cmdDelete() {
    if len(os.Args) < 3 {
        fmt.Println("Usage: ln-images delete <image_id>")
        os.Exit(1)
    }

    apiKey := getAPIKey()
    imageID := os.Args[2]

    client := labnocturne.NewClient(apiKey)
    if err := client.DeleteFile(imageID); err != nil {
        log.Fatal("Delete failed:", err)
    }

    fmt.Println("Deleted:", imageID)
}

func getAPIKey() string {
    apiKey := os.Getenv("LABNOCTURNE_API_KEY")
    if apiKey == "" {
        log.Fatal("LABNOCTURNE_API_KEY environment variable not set")
    }
    return apiKey
}
```

Build and install:
```bash
go build -o ln-images ./cmd/ln-images
sudo mv ln-images /usr/local/bin/

# Usage
ln-images key
export LABNOCTURNE_API_KEY='ln_test_...'
ln-images upload photo.jpg
ln-images list
ln-images stats
ln-images delete img_01jcd...
```

## Go Module Setup

```go
// go.mod
module github.com/jjenkins/labnocturne-image-client/go

go 1.21
```

## Project Structure

```
go/
├── labnocturne/
│   └── client.go       # Main client package
├── example/
│   └── main.go         # Complete usage example
├── cmd/
│   └── ln-images/
│       └── main.go     # CLI tool
├── go.mod
└── README.md
```

## Error Handling Example

```go
package main

import (
    "errors"
    "fmt"
    "net/http"

    "github.com/jjenkins/labnocturne-image-client/go/labnocturne"
)

func safeUpload(client *labnocturne.Client, filePath string) {
    result, err := client.Upload(filePath)
    if err != nil {
        // Check for specific error types
        if errors.Is(err, os.ErrNotExist) {
            fmt.Println("Error: File not found")
        } else if errors.Is(err, http.ErrBodyNotAllowed) {
            fmt.Println("Error: Invalid request")
        } else {
            fmt.Printf("Error: %v\n", err)
        }
        return
    }

    fmt.Println("Success:", result.URL)
}
```

## Testing

```go
// client_test.go
package labnocturne_test

import (
    "testing"

    "github.com/jjenkins/labnocturne-image-client/go/labnocturne"
)

func TestGenerateKey(t *testing.T) {
    apiKey, err := labnocturne.GenerateTestKey()
    if err != nil {
        t.Fatalf("Failed to generate key: %v", err)
    }

    if apiKey == "" {
        t.Fatal("API key is empty")
    }

    if !strings.HasPrefix(apiKey, "ln_test_") {
        t.Errorf("Expected test key prefix, got: %s", apiKey)
    }
}

func TestUpload(t *testing.T) {
    apiKey, err := labnocturne.GenerateTestKey()
    if err != nil {
        t.Fatal(err)
    }

    client := labnocturne.NewClient(apiKey)

    // Create a test image file
    testFile := "test_image.jpg"
    defer os.Remove(testFile)

    result, err := client.Upload(testFile)
    if err != nil {
        t.Fatalf("Upload failed: %v", err)
    }

    if result.ID == "" {
        t.Error("Expected non-empty ID")
    }

    if result.URL == "" {
        t.Error("Expected non-empty URL")
    }
}
```

Run tests:
```bash
go test ./...
```

## Next Steps

- Try the [Ruby examples](../ruby/)
- Try the [PHP examples](../php/)
- Check out [Python examples](../python/) or [JavaScript examples](../javascript/)
- Read the [API documentation](https://images.labnocturne.com/docs)
