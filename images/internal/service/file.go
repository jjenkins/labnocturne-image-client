package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/h2non/filetype"
	"github.com/jjenkins/labnocturne/images/internal/model"
	"github.com/jjenkins/labnocturne/images/internal/store"
	"github.com/oklog/ulid/v2"
)

var (
	ErrInvalidAPIKey    = errors.New("invalid API key")
	ErrFileTooLarge     = errors.New("file too large")
	ErrInvalidFileType  = errors.New("invalid file type")
	ErrStorageExceeded  = errors.New("storage quota exceeded")
	ErrFileNotFound     = errors.New("file not found")
	ErrInvalidSortOrder = errors.New("invalid sort order")
)

// FileService handles file operations (upload and retrieval)
type FileService struct {
	userStore       *store.UserStore
	fileStore       *store.FileStore
	s3Client        *s3.Client
	baseURL         string
	s3Bucket        string
	maxFileSizeTest int64
	maxFileSizeLive int64
}

// NewFileService creates a new FileService
func NewFileService(userStore *store.UserStore, fileStore *store.FileStore, s3Client *s3.Client, baseURL string, s3Bucket string) *FileService {
	// Cache max file sizes from env vars at startup
	maxTest, maxLive := loadMaxFileSizes()

	return &FileService{
		userStore:       userStore,
		fileStore:       fileStore,
		s3Client:        s3Client,
		baseURL:         baseURL,
		s3Bucket:        s3Bucket,
		maxFileSizeTest: maxTest,
		maxFileSizeLive: maxLive,
	}
}

// NewUploadService is deprecated. Use NewFileService instead.
// Kept for backwards compatibility with existing code.
func NewUploadService(userStore *store.UserStore, fileStore *store.FileStore, s3Client *s3.Client, baseURL string, s3Bucket string) *FileService {
	return NewFileService(userStore, fileStore, s3Client, baseURL, s3Bucket)
}

// Upload handles the complete file upload process
func (s *FileService) Upload(ctx context.Context, apiKey string, fileHeader *multipart.FileHeader) (*model.File, error) {
	// 1. Authenticate user
	user, err := s.userStore.FindByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, ErrInvalidAPIKey
	}

	// 2. Check file size limits
	maxSize := s.getMaxFileSizeForUser(user.KeyType)
	if fileHeader.Size > maxSize {
		return nil, ErrFileTooLarge
	}

	// 3. Check storage quota before uploading
	currentUsage, _, err := s.userStore.GetStorageUsage(ctx, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to check storage quota: %w", err)
	}

	storageQuota := getStorageQuotaForPlan(user.Plan)
	if currentUsage+fileHeader.Size > storageQuota {
		return nil, ErrStorageExceeded
	}

	// 4. Open file for streaming
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// 5. Read first 8KB for magic byte detection
	header := make([]byte, 8192)
	n, err := io.ReadFull(src, header)
	if err != nil && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file header: %w", err)
	}

	// 6. Validate file type using magic bytes
	extension, mimeType, err := validateFileType(header[:n])
	if err != nil {
		return nil, ErrInvalidFileType
	}

	// 7. Generate ULID and keys
	externalID, s3Key, cdnURL, rawULID, err := generateFileKeys(extension, s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate file keys: %w", err)
	}

	// 8. Create multiReader to combine header + rest of file for streaming
	fullReader := io.MultiReader(bytes.NewReader(header[:n]), src)

	// 9. Upload to S3 with streaming, encryption, and cache headers
	_, err = s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(s.s3Bucket),
		Key:                  aws.String(s3Key),
		Body:                 fullReader,
		ContentLength:        aws.Int64(fileHeader.Size),                // Required for streaming
		ContentType:          aws.String(mimeType),
		ServerSideEncryption: types.ServerSideEncryptionAes256,          // Enable S3 encryption
		CacheControl:         aws.String("max-age=31536000"),            // 1 year cache for CloudFront
		StorageClass:         types.StorageClassIntelligentTiering,      // Cost optimization
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload to S3: %w", err)
	}

	// 10. Save metadata to database
	file := &model.File{
		ID:         rawULID,
		ExternalID: externalID,
		UserID:     user.ID,
		Filename:   fileHeader.Filename,
		Extension:  extension,
		SizeBytes:  fileHeader.Size,
		MimeType:   mimeType,
		S3Key:      s3Key,
		CDNURL:     cdnURL,
	}

	if err := s.fileStore.Create(ctx, file); err != nil {
		// Attempt to clean up S3 object if DB insert fails
		s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.s3Bucket),
			Key:    aws.String(s3Key),
		})
		return nil, fmt.Errorf("failed to save file metadata: %w", err)
	}

	return file, nil
}

// loadMaxFileSizes reads max file size limits from environment variables at startup
func loadMaxFileSizes() (testMax int64, liveMax int64) {
	testMax = 10 * 1024 * 1024  // Default 10MB for test keys
	liveMax = 100 * 1024 * 1024 // Default 100MB for live keys

	if maxStr := os.Getenv("MAX_FILE_SIZE_TEST"); maxStr != "" {
		if max, err := strconv.ParseInt(maxStr, 10, 64); err == nil {
			testMax = max
		}
	}

	if maxStr := os.Getenv("MAX_FILE_SIZE_LIVE"); maxStr != "" {
		if max, err := strconv.ParseInt(maxStr, 10, 64); err == nil {
			liveMax = max
		}
	}

	return testMax, liveMax
}

// getMaxFileSizeForUser returns the cached maximum file size for a key type
func (s *FileService) getMaxFileSizeForUser(keyType string) int64 {
	switch keyType {
	case "live":
		return s.maxFileSizeLive
	default:
		return s.maxFileSizeTest
	}
}

// getStorageQuotaForPlan returns the storage quota for a user's plan
func getStorageQuotaForPlan(plan string) int64 {
	switch plan {
	case "starter":
		return 10 * 1024 * 1024 * 1024 // 10GB
	case "pro":
		return 100 * 1024 * 1024 * 1024 // 100GB
	default: // "test"
		return 100 * 1024 * 1024 // 100MB
	}
}

// validateFileType checks file magic bytes and returns extension and MIME type
func validateFileType(fileData []byte) (extension string, mimeType string, err error) {
	kind, err := filetype.Match(fileData)
	if err != nil {
		return "", "", err
	}

	allowedTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
		"svg":  "image/svg+xml",
	}

	mimeType, ok := allowedTypes[kind.Extension]
	if !ok {
		return "", "", fmt.Errorf("file type not allowed: %s", kind.Extension)
	}

	return kind.Extension, mimeType, nil
}

// generateFileKeys creates ULID and all related keys
func generateFileKeys(extension string, baseURL string) (externalID, s3Key, cdnURL, rawULID string, err error) {
	// Generate ULID with monotonic entropy
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	rawULID = id.String() // e.g., "01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Create external ID (what API returns)
	externalID = "img_" + strings.ToLower(rawULID)

	// Create S3 partition path using lowercase first 3 chars for partitioning
	lower := strings.ToLower(rawULID)
	s3Key = fmt.Sprintf("%c/%c/%c/%s.%s", lower[0], lower[1], lower[2], rawULID, extension)
	// Result: "0/1/a/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg"

	// Create clean CDN URL (what users see)
	cdnURL = fmt.Sprintf("%s/i/%s.%s", baseURL, rawULID, extension)
	// Result: "http://localhost:8080/i/01ARZ3NDEKTSV4RRFFQ69G5FAV.jpg"

	return externalID, s3Key, cdnURL, rawULID, nil
}

// GetByULID retrieves a file by its ULID (case-insensitive)
func (s *FileService) GetByULID(ctx context.Context, ulid string) (*model.File, error) {
	// Normalize ULID to uppercase for database lookup
	ulid = strings.ToUpper(ulid)

	file, err := s.fileStore.FindByID(ctx, ulid)
	if err != nil {
		return nil, ErrFileNotFound
	}

	return file, nil
}

// GetPresignedURL generates a temporary presigned URL for accessing a file
// If CloudFront domain is configured, returns CloudFront URL, otherwise falls back to S3 presigned URL
func (s *FileService) GetPresignedURL(ctx context.Context, s3Key string) (string, error) {
	// Check if CloudFront is configured
	cloudFrontDomain := os.Getenv("CLOUDFRONT_DOMAIN_NAME")
	if cloudFrontDomain != "" {
		// Use CloudFront URL (no presigning needed if distribution is public)
		// CloudFront will fetch from S3 origin using its IAM role
		return fmt.Sprintf("https://%s/%s", cloudFrontDomain, s3Key), nil
	}

	// Fallback to S3 presigned URL
	presignClient := s3.NewPresignClient(s.s3Client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.s3Bucket),
		Key:    aws.String(s3Key),
	}, s3.WithPresignExpires(1*time.Hour))

	if err != nil {
		return "", fmt.Errorf("failed to presign request: %w", err)
	}

	return request.URL, nil
}

// DeleteFile soft-deletes a file (sets deleted_at timestamp)
// Only the file owner can delete their files
func (s *FileService) DeleteFile(ctx context.Context, fileID string, userID string) error {
	// Normalize ID: handle both external ID (img_...) and raw ULID
	ulid := fileID
	if strings.HasPrefix(fileID, "img_") {
		ulid = strings.ToUpper(strings.TrimPrefix(fileID, "img_"))
	} else {
		ulid = strings.ToUpper(ulid)
	}

	// Get file to verify ownership
	file, err := s.fileStore.FindByID(ctx, ulid)
	if err != nil {
		return ErrFileNotFound
	}

	// Check ownership (security: don't reveal if file exists but belongs to someone else)
	if file.UserID.String() != userID {
		return ErrFileNotFound
	}

	// Soft delete (idempotent: if already deleted, FindByID would have returned error)
	err = s.fileStore.SoftDelete(ctx, ulid)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ListFilesParams contains parameters for listing files
type ListFilesParams struct {
	Limit     int
	Offset    int
	SortOrder string
}

// ListFilesResponse contains the paginated file list response
type ListFilesResponse struct {
	Files   []*model.File
	Total   int64
	Limit   int
	Offset  int
	NextURL *string // Pointer to string, nil if no next page
}

// ListFiles retrieves a paginated list of files for a user
func (s *FileService) ListFiles(ctx context.Context, userID string, params ListFilesParams) (*ListFilesResponse, error) {
	// Validate and apply defaults
	if params.Limit <= 0 {
		params.Limit = 100
	}
	if params.Limit > 1000 {
		params.Limit = 1000
	}
	if params.Offset < 0 {
		params.Offset = 0
	}
	if params.SortOrder == "" {
		params.SortOrder = "uploaded_at_desc"
	}

	// Validate sort order
	validSortOrders := map[string]bool{
		"uploaded_at_desc": true,
		"uploaded_at_asc":  true,
		"size_desc":        true,
		"size_asc":         true,
	}
	if !validSortOrders[params.SortOrder] {
		return nil, ErrInvalidSortOrder
	}

	// Get files from store
	files, total, err := s.fileStore.FindByUserID(ctx, userID, params.Limit, params.Offset, params.SortOrder)
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Calculate next URL (nil if no more pages)
	var nextURL *string
	if int64(params.Offset+params.Limit) < total {
		nextOffset := params.Offset + params.Limit
		url := fmt.Sprintf("/files?limit=%d&offset=%d&sort=%s", params.Limit, nextOffset, params.SortOrder)
		nextURL = &url
	}

	return &ListFilesResponse{
		Files:   files,
		Total:   total,
		Limit:   params.Limit,
		Offset:  params.Offset,
		NextURL: nextURL,
	}, nil
}
