package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// CleanupService handles cleanup operations for expired files
type CleanupService struct {
	fileStore *store.FileStore
	s3Client  *s3.Client
	s3Bucket  string
	dryRun    bool
}

// NewCleanupService creates a new CleanupService
func NewCleanupService(db *sql.DB, s3Client *s3.Client, s3Bucket string, dryRun bool) *CleanupService {
	return &CleanupService{
		fileStore: store.NewFileStore(db),
		s3Client:  s3Client,
		s3Bucket:  s3Bucket,
		dryRun:    dryRun,
	}
}

// CleanupExpiredTestFiles deletes test key files older than 7 days
// Returns the count of successfully deleted files and any error encountered
func (s *CleanupService) CleanupExpiredTestFiles(ctx context.Context) (int, error) {
	// 1. Find expired test files
	files, err := s.fileStore.FindExpiredTestFiles(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to find expired test files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("No expired test files to delete")
		return 0, nil
	}

	log.Printf("Found %d expired test files to delete", len(files))

	// 2. Delete each file from S3 and database
	successCount := 0
	var lastError error

	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping cleanup after processing %d files", successCount)
			return successCount, ctx.Err()
		default:
		}

		// Format file size for logging
		sizeMB := float64(file.SizeBytes) / (1024 * 1024)

		if s.dryRun {
			log.Printf("[DRY RUN] Would delete test file: %s (%s, %.2f MB)", file.ID, file.Filename, sizeMB)
			successCount++
			continue
		}

		// Delete from S3
		_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.s3Bucket),
			Key:    aws.String(file.S3Key),
		})
		if err != nil {
			log.Printf("Failed to delete test file %s (%s) from S3: %v", file.ID, file.Filename, err)
			lastError = err
			continue
		}

		// Delete from database (only if S3 delete succeeded)
		err = s.fileStore.PermanentlyDelete(ctx, file.ID)
		if err != nil {
			log.Printf("Failed to delete test file %s (%s) from database: %v", file.ID, file.Filename, err)
			lastError = err
			continue
		}

		log.Printf("Deleted test file: %s (%s, %.2f MB)", file.ID, file.Filename, sizeMB)
		successCount++
	}

	if lastError != nil {
		return successCount, fmt.Errorf("completed with errors (deleted %d/%d files): %w", successCount, len(files), lastError)
	}

	return successCount, nil
}

// CleanupExpiredSoftDeleted permanently deletes soft-deleted files older than 30 days
// Returns the count of successfully deleted files and any error encountered
func (s *CleanupService) CleanupExpiredSoftDeleted(ctx context.Context) (int, error) {
	// 1. Find expired soft-deleted files
	files, err := s.fileStore.FindExpiredSoftDeleted(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to find expired soft-deleted files: %w", err)
	}

	if len(files) == 0 {
		log.Printf("No expired soft-deleted files to delete")
		return 0, nil
	}

	log.Printf("Found %d expired soft-deleted files to permanently delete", len(files))

	// 2. Delete each file from S3 and database
	successCount := 0
	var lastError error

	for _, file := range files {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			log.Printf("Context cancelled, stopping cleanup after processing %d files", successCount)
			return successCount, ctx.Err()
		default:
		}

		// Format file size for logging
		sizeMB := float64(file.SizeBytes) / (1024 * 1024)

		if s.dryRun {
			log.Printf("[DRY RUN] Would permanently delete soft-deleted file: %s (%s, %.2f MB)", file.ID, file.Filename, sizeMB)
			successCount++
			continue
		}

		// Delete from S3
		_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: aws.String(s.s3Bucket),
			Key:    aws.String(file.S3Key),
		})
		if err != nil {
			log.Printf("Failed to delete soft-deleted file %s (%s) from S3: %v", file.ID, file.Filename, err)
			lastError = err
			continue
		}

		// Delete from database (only if S3 delete succeeded)
		err = s.fileStore.PermanentlyDelete(ctx, file.ID)
		if err != nil {
			log.Printf("Failed to delete soft-deleted file %s (%s) from database: %v", file.ID, file.Filename, err)
			lastError = err
			continue
		}

		log.Printf("Permanently deleted soft-deleted file: %s (%s, %.2f MB)", file.ID, file.Filename, sizeMB)
		successCount++
	}

	if lastError != nil {
		return successCount, fmt.Errorf("completed with errors (deleted %d/%d files): %w", successCount, len(files), lastError)
	}

	return successCount, nil
}
