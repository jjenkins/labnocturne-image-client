# Lab Nocturne Images - Go Client

Go client library for the [Lab Nocturne Images API](https://images.labnocturne.com).

## Installation

```bash
go get github.com/jjenkins/labnocturne-image-client/go/labnocturne
```

## Requirements

- Go 1.21+
- No external dependencies (uses standard library only)

## Quick Start

```go
package main

import (
	"fmt"
	"log"

	"github.com/jjenkins/labnocturne-image-client/go/labnocturne"
)

func main() {
	// Generate a test API key
	apiKey, err := labnocturne.GenerateTestKey()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("API Key:", apiKey)

	// Create client
	client := labnocturne.NewClient(apiKey)

	// Upload an image
	result, err := client.Upload("photo.jpg")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Image URL:", result.URL)
	fmt.Println("Image ID:", result.ID)

	// List files
	files, err := client.ListFiles(1, 10, "created_desc")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total files: %d\n", files.Pagination.Total)

	// Get usage stats
	stats, err := client.GetStats()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Storage: %.2f MB / %.0f MB\n", stats.StorageUsedMB, stats.QuotaMB)

	// Delete a file
	err = client.DeleteFile(result.ID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("File deleted")
}
```

## API Reference

### Types

#### `Client`

```go
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}
```

The main client for interacting with the API.

#### `UploadResponse`

```go
type UploadResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	Size      int64  `json:"size"`
	MimeType  string `json:"mime_type"`
	CreatedAt string `json:"created_at"`
}
```

Response from uploading a file.

#### `ListFilesResponse`

```go
type ListFilesResponse struct {
	Files      []FileInfo
	Pagination struct {
		Page       int
		Limit      int
		Total      int
		TotalPages int
	}
}
```

Response from listing files.

#### `StatsResponse`

```go
type StatsResponse struct {
	StorageUsedBytes int64
	StorageUsedMB    float64
	FileCount        int
	QuotaBytes       int64
	QuotaMB          float64
	UsagePercent     float64
}
```

Usage statistics for your account.

### Functions

#### `NewClient(apiKey string) *Client`

Create a new client instance.

**Parameters:**
- `apiKey`: Your API key

**Returns:** `*Client`

**Example:**
```go
client := labnocturne.NewClient("ln_test_abc123...")
```

#### `GenerateTestKey() (string, error)`

Generate a test API key for development.

**Returns:** API key string and error

**Example:**
```go
apiKey, err := labnocturne.GenerateTestKey()
if err != nil {
	log.Fatal(err)
}
```

### Methods

#### `(*Client) Upload(filePath string) (*UploadResponse, error)`

Upload an image file.

**Parameters:**
- `filePath`: Path to the image file

**Returns:** `*UploadResponse` and error

**Example:**
```go
result, err := client.Upload("photo.jpg")
if err != nil {
	log.Fatal(err)
}
fmt.Println("Uploaded:", result.URL)
```

#### `(*Client) ListFiles(page, limit int, sort string) (*ListFilesResponse, error)`

List uploaded files with pagination.

**Parameters:**
- `page`: Page number
- `limit`: Files per page
- `sort`: Sort order (`created_desc`, `created_asc`, `size_desc`, `size_asc`, `name_asc`, `name_desc`)

**Returns:** `*ListFilesResponse` and error

**Example:**
```go
files, err := client.ListFiles(1, 50, "created_desc")
if err != nil {
	log.Fatal(err)
}
for _, file := range files.Files {
	fmt.Printf("%s: %d bytes\n", file.ID, file.Size)
}
```

#### `(*Client) GetStats() (*StatsResponse, error)`

Get usage statistics for your account.

**Returns:** `*StatsResponse` and error

**Example:**
```go
stats, err := client.GetStats()
if err != nil {
	log.Fatal(err)
}
fmt.Printf("Using %.1f%% of quota\n", stats.UsagePercent)
```

#### `(*Client) DeleteFile(imageID string) error`

Delete an image (soft delete).

**Parameters:**
- `imageID`: The image ID

**Returns:** error

**Example:**
```go
err := client.DeleteFile("img_01jcd8x9k2n...")
if err != nil {
	log.Fatal(err)
}
```

## Complete Example

```go
package main

import (
	"fmt"
	"log"

	"github.com/jjenkins/labnocturne-image-client/go/labnocturne"
)

func main() {
	// Generate test API key
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

## Error Handling

All methods return standard Go errors. Use error checking:

```go
result, err := client.Upload("photo.jpg")
if err != nil {
	// Handle error
	if errors.Is(err, os.ErrNotExist) {
		log.Println("File not found")
	} else {
		log.Printf("Upload failed: %v\n", err)
	}
	return
}
```

## Custom HTTP Client

You can provide a custom HTTP client:

```go
import (
	"net/http"
	"time"
)

client := labnocturne.NewClient(apiKey)
client.HTTPClient = &http.Client{
	Timeout: 30 * time.Second,
}
```

## Custom Base URL

For testing or custom deployments:

```go
client := labnocturne.NewClient(apiKey)
client.BaseURL = "https://custom.example.com"
```

## License

MIT License - See [LICENSE](../LICENSE) for details.

## Links

- [Main Repository](https://github.com/jjenkins/labnocturne-image-client)
- [API Documentation](https://images.labnocturne.com/docs)
- [Other Language Clients](https://github.com/jjenkins/labnocturne-image-client#readme)
