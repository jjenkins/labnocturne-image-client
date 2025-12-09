package service

import (
	"bytes"
	"compress/gzip"
	"context"
	"testing"
)

func TestExtractULIDFromURI(t *testing.T) {
	tests := []struct {
		name      string
		uri       string
		wantULID  string
		wantError bool
	}{
		{
			name:      "Valid URI with jpg extension",
			uri:       "/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg",
			wantULID:  "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantError: false,
		},
		{
			name:      "Valid URI with png extension",
			uri:       "/i/01BX5ZZKBKACTAV9WEVGEMMVRZ.png",
			wantULID:  "01BX5ZZKBKACTAV9WEVGEMMVRZ",
			wantError: false,
		},
		{
			name:      "Valid URI without leading slash",
			uri:       "i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg",
			wantULID:  "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantError: false,
		},
		{
			name:      "Lowercase ULID (should be converted to uppercase)",
			uri:       "/i/01arz3ndektsv4rrffq69g5fav.jpg",
			wantULID:  "01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantError: false,
		},
		{
			name:      "Invalid URI - missing /i/ prefix",
			uri:       "/images/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg",
			wantULID:  "",
			wantError: true,
		},
		{
			name:      "Invalid URI - no extension",
			uri:       "/i/01ARZ3NDEKTSV4RRFFQ69G5FAV",
			wantULID:  "",
			wantError: true,
		},
		{
			name:      "Invalid URI - wrong ULID length",
			uri:       "/i/TOOSHORT.jpg",
			wantULID:  "",
			wantError: true,
		},
		{
			name:      "Invalid URI - empty",
			uri:       "",
			wantULID:  "",
			wantError: true,
		},
		{
			name:      "Invalid URI - only slash",
			uri:       "/",
			wantULID:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotULID, err := ExtractULIDFromURI(tt.uri)

			if tt.wantError {
				if err == nil {
					t.Errorf("ExtractULIDFromURI() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ExtractULIDFromURI() unexpected error: %v", err)
				return
			}

			if gotULID != tt.wantULID {
				t.Errorf("ExtractULIDFromURI() = %v, want %v", gotULID, tt.wantULID)
			}
		})
	}
}

func TestProcessLogReader(t *testing.T) {
	// Create mock CloudFront log data with only comment lines
	// This tests that the gzip decompression and comment line skipping works
	// without requiring database lookups
	mockLogData := `#Version: 1.0
#Fields: date time x-edge-location sc-bytes c-ip cs-method cs(Host) cs-uri-stem sc-status cs(Referer) cs(User-Agent) cs-uri-query cs(Cookie) x-edge-result-type x-edge-request-id x-host-header cs-protocol cs-bytes time-taken x-forwarded-for ssl-protocol ssl-cipher x-edge-response-result-type cs-protocol-version
`

	// Compress the mock log data
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	if _, err := gzWriter.Write([]byte(mockLogData)); err != nil {
		t.Fatalf("Failed to compress mock log data: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %v", err)
	}

	// Create mock service (without database dependencies)
	service := &BandwidthService{
		fileStore: nil, // No database needed for header-only log
	}

	// Process the log reader
	userBandwidth := make(map[string]*bandwidthAggregation)
	ctx := context.Background()

	err := service.ProcessLogReader(ctx, &buf, userBandwidth)
	if err != nil {
		t.Fatalf("ProcessLogReader() failed: %v", err)
	}

	// userBandwidth should be empty since we only had comment lines
	if len(userBandwidth) != 0 {
		t.Errorf("Expected empty userBandwidth, got %d entries", len(userBandwidth))
	}

	t.Logf("Successfully processed log reader with comment lines only")
}

func TestProcessLogLine_Parsing(t *testing.T) {
	// Test that log line parsing extracts the correct fields
	// CloudFront log format (tab-separated):
	// 0:date 1:time 2:x-edge-location 3:sc-bytes 4:c-ip 5:cs-method 6:cs(Host) 7:cs-uri-stem ...
	logLine := "2025-12-03	14:23:45	IAD89-C1	12345	1.2.3.4	GET	d12345.cloudfront.net	/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg	200	-	Mozilla/5.0	-	-	Hit	abc123	example.com	https	500	0.001	-	TLSv1.3	TLS_AES_128_GCM_SHA256	Hit	HTTP/2.0"

	fields := bytes.Split([]byte(logLine), []byte("\t"))

	if len(fields) < 15 {
		t.Fatalf("Expected at least 15 fields, got %d", len(fields))
	}

	// Note: sc-bytes is at index 3 (not 4)
	// Verify field 3 (sc-bytes)
	scBytes := string(fields[3])
	if scBytes != "12345" {
		t.Errorf("Expected sc-bytes = '12345', got '%s'", scBytes)
	}

	// Verify field 7 (cs-uri-stem)
	uriStem := string(fields[7])
	if uriStem != "/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg" {
		t.Errorf("Expected uri-stem = '/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg', got '%s'", uriStem)
	}

	// Verify ULID extraction
	ulid, err := ExtractULIDFromURI(uriStem)
	if err != nil {
		t.Fatalf("Failed to extract ULID: %v", err)
	}
	if ulid != "01ARZ3NDEKTSV4RRFFQ69G5FAV" {
		t.Errorf("Expected ULID = '01ARZ3NDEKTSV4RRFFQ69G5FAV', got '%s'", ulid)
	}
}

func TestBandwidthAggregation(t *testing.T) {
	// Test that bandwidth aggregation accumulates correctly
	userBandwidth := make(map[string]*bandwidthAggregation)
	userID := "test-user-123"

	// Simulate multiple requests from the same user
	for i := 0; i < 5; i++ {
		if _, exists := userBandwidth[userID]; !exists {
			userBandwidth[userID] = &bandwidthAggregation{
				bytesServed:  0,
				requestCount: 0,
			}
		}
		userBandwidth[userID].bytesServed += 1000
		userBandwidth[userID].requestCount++
	}

	// Verify aggregation
	if userBandwidth[userID].bytesServed != 5000 {
		t.Errorf("Expected bytes_served = 5000, got %d", userBandwidth[userID].bytesServed)
	}
	if userBandwidth[userID].requestCount != 5 {
		t.Errorf("Expected request_count = 5, got %d", userBandwidth[userID].requestCount)
	}
}

func TestExtractULIDFromURI_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		uri         string
		wantError   bool
		description string
	}{
		{
			name:        "URI with query parameters",
			uri:         "/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg?size=large",
			wantError:   false,
			description: "LastIndex finds the last dot, so query params after extension are ignored",
		},
		{
			name:        "URI with multiple dots in filename",
			uri:         "/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.thumb.jpg",
			wantError:   true,
			description: "Multiple dots mean ULID is extracted up to last dot, resulting in wrong length",
		},
		{
			name:        "URI with unusual extension",
			uri:         "/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.webp",
			wantError:   false,
			description: "Should work fine with any extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ulid, err := ExtractULIDFromURI(tt.uri)

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error for URI %s (%s), got ULID: %s", tt.uri, tt.description, ulid)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for URI %s: %v", tt.uri, err)
				}
				if len(ulid) != 26 {
					t.Errorf("Expected ULID length 26, got %d for URI %s", len(ulid), tt.uri)
				}
			}
		})
	}
}
