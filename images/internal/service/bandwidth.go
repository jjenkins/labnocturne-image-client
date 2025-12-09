package service

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// BandwidthService processes CloudFront logs and aggregates bandwidth usage
type BandwidthService struct {
	s3Client       *s3.Client
	fileStore      *store.FileStore
	bandwidthStore *store.BandwidthStore
	logBucket      string
	logPrefix      string
}

// NewBandwidthService creates a new BandwidthService
func NewBandwidthService(s3Client *s3.Client, fileStore *store.FileStore, bandwidthStore *store.BandwidthStore, logBucket, logPrefix string) *BandwidthService {
	return &BandwidthService{
		s3Client:       s3Client,
		fileStore:      fileStore,
		bandwidthStore: bandwidthStore,
		logBucket:      logBucket,
		logPrefix:      logPrefix,
	}
}

// ProcessLogFiles processes CloudFront logs for a given date
func (s *BandwidthService) ProcessLogFiles(ctx context.Context, date time.Time) (int, error) {
	// 1. List log files in S3 for the given date
	logFiles, err := s.listLogFiles(ctx, date)
	if err != nil {
		return 0, fmt.Errorf("failed to list log files: %w", err)
	}

	if len(logFiles) == 0 {
		log.Printf("No CloudFront logs found for date %s", date.Format("2006-01-02"))
		return 0, nil
	}

	log.Printf("Found %d CloudFront log files for date %s", len(logFiles), date.Format("2006-01-02"))

	// 2. Process each log file and aggregate by user
	userBandwidth := make(map[string]*bandwidthAggregation)

	for _, logFile := range logFiles {
		if err := s.processLogFile(ctx, logFile, userBandwidth); err != nil {
			log.Printf("warning: failed to process log file %s: %v", logFile, err)
			continue
		}
	}

	// 3. Save aggregated data to database
	processedCount := 0
	for userID, agg := range userBandwidth {
		if err := s.bandwidthStore.UpdateBandwidth(ctx, userID, date, agg.bytesServed, agg.requestCount); err != nil {
			log.Printf("error: failed to update bandwidth for user %s: %v", userID, err)
		} else {
			processedCount++
		}
	}

	log.Printf("Successfully processed bandwidth for %d users on %s", processedCount, date.Format("2006-01-02"))
	return processedCount, nil
}

// bandwidthAggregation holds aggregated bandwidth data for a user
type bandwidthAggregation struct {
	bytesServed  int64
	requestCount int
}

// listLogFiles lists CloudFront log files in S3 for a specific date
func (s *BandwidthService) listLogFiles(ctx context.Context, date time.Time) ([]string, error) {
	// CloudFront log file naming pattern: E<DISTRIBUTION_ID>.YYYY-MM-DD-HH.xxxx.gz
	dateStr := date.Format("2006-01-02")

	// List objects with the log prefix
	result, err := s.s3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.logBucket),
		Prefix: aws.String(s.logPrefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	// Filter log files by date
	var logFiles []string
	for _, obj := range result.Contents {
		key := aws.ToString(obj.Key)
		// Check if the filename contains the date string
		if strings.Contains(key, dateStr) {
			logFiles = append(logFiles, key)
		}
	}

	return logFiles, nil
}

// processLogFile parses a single CloudFront log file (gzipped TSV)
func (s *BandwidthService) processLogFile(ctx context.Context, key string, userBandwidth map[string]*bandwidthAggregation) error {
	// Download log file from S3
	result, err := s.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.logBucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get S3 object: %w", err)
	}
	defer result.Body.Close()

	// Decompress gzipped log
	gzReader, err := gzip.NewReader(result.Body)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Parse TSV format
	lineCount := 0
	processedCount := 0
	scanner := bufio.NewScanner(gzReader)
	for scanner.Scan() {
		lineCount++
		line := scanner.Text()

		// Skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Parse the log line
		if err := s.processLogLine(ctx, line, userBandwidth); err != nil {
			// Log parsing errors but don't fail the entire file
			if lineCount <= 5 {
				log.Printf("debug: skipping malformed log line in %s (line %d): %v", key, lineCount, err)
			}
			continue
		}
		processedCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	log.Printf("Processed %d/%d lines from %s", processedCount, lineCount, key)
	return nil
}

// processLogLine parses a single log line and updates the bandwidth aggregation
func (s *BandwidthService) processLogLine(ctx context.Context, line string, userBandwidth map[string]*bandwidthAggregation) error {
	// CloudFront log format is TSV with specific field positions (0-indexed)
	// Field 3: sc-bytes (bytes served)
	// Field 7: cs-uri-stem (request URI)
	fields := strings.Split(line, "\t")
	if len(fields) < 15 {
		return fmt.Errorf("insufficient fields: expected at least 15, got %d", len(fields))
	}

	// Extract bytes served (field 3)
	scBytesStr := fields[3]
	bytesServed, err := strconv.ParseInt(scBytesStr, 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse sc-bytes '%s': %w", scBytesStr, err)
	}

	// Extract URI (field 7)
	uriStem := fields[7]

	// Extract ULID from URI
	ulid, err := extractULIDFromURI(uriStem)
	if err != nil {
		return fmt.Errorf("failed to extract ULID: %w", err)
	}

	// Look up file to get user_id
	file, err := s.fileStore.FindByID(ctx, ulid)
	if err != nil {
		// File not found or deleted - skip silently
		return fmt.Errorf("file not found for ULID %s: %w", ulid, err)
	}

	// Aggregate bytes per user
	userID := file.UserID.String()
	if _, exists := userBandwidth[userID]; !exists {
		userBandwidth[userID] = &bandwidthAggregation{
			bytesServed:  0,
			requestCount: 0,
		}
	}
	userBandwidth[userID].bytesServed += bytesServed
	userBandwidth[userID].requestCount++

	return nil
}

// extractULIDFromURI extracts ULID from CloudFront URI
// Example: "/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg" -> "01ARZ3NDEKTSV4RRFFQ69G5FAV"
func extractULIDFromURI(uri string) (string, error) {
	// Trim leading slash
	uri = strings.TrimPrefix(uri, "/")

	// Split by slash
	parts := strings.Split(uri, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid URI format: %s", uri)
	}

	// Check if first part is "i"
	if parts[0] != "i" {
		return "", fmt.Errorf("URI does not match /i/ pattern: %s", uri)
	}

	// Extract filename (second part)
	filename := parts[1]

	// Remove extension
	dotIndex := strings.LastIndex(filename, ".")
	if dotIndex == -1 {
		return "", fmt.Errorf("no extension found in filename: %s", filename)
	}

	ulid := filename[:dotIndex]

	// Validate ULID length (26 characters)
	if len(ulid) != 26 {
		return "", fmt.Errorf("invalid ULID length: expected 26, got %d", len(ulid))
	}

	// Convert to uppercase (canonical form)
	ulid = strings.ToUpper(ulid)

	return ulid, nil
}

// ExtractULIDFromURI is exported for testing
func ExtractULIDFromURI(uri string) (string, error) {
	return extractULIDFromURI(uri)
}

// ProcessLogReader processes a log file from an io.Reader (exported for testing)
func (s *BandwidthService) ProcessLogReader(ctx context.Context, reader io.Reader, userBandwidth map[string]*bandwidthAggregation) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	scanner := bufio.NewScanner(gzReader)
	for scanner.Scan() {
		line := scanner.Text()

		// Skip comment lines
		if strings.HasPrefix(line, "#") {
			continue
		}

		// Process log line (ignore errors for individual lines)
		_ = s.processLogLine(ctx, line, userBandwidth)
	}

	return scanner.Err()
}
