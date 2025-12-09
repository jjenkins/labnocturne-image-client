package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// AgeTestFile sets a file's uploaded_at timestamp to simulate age
// This helper is used for testing time-based file expiration
func AgeTestFile(db *sql.DB, fileID string, days int) error {
	query := `UPDATE files SET uploaded_at = NOW() - INTERVAL '1 day' * $1 WHERE id = $2`
	result, err := db.ExecContext(context.Background(), query, days, fileID)
	if err != nil {
		return fmt.Errorf("failed to age test file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found: %s", fileID)
	}

	return nil
}

// AgeSoftDeletedFile sets a file's deleted_at timestamp to simulate age
// This helper is used for testing time-based permanent deletion
func AgeSoftDeletedFile(db *sql.DB, fileID string, days int) error {
	query := `UPDATE files SET deleted_at = NOW() - INTERVAL '1 day' * $1 WHERE id = $2`
	result, err := db.ExecContext(context.Background(), query, days, fileID)
	if err != nil {
		return fmt.Errorf("failed to age soft-deleted file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("file not found: %s", fileID)
	}

	return nil
}

// GetFileUploadedAt retrieves a file's uploaded_at timestamp for verification
func GetFileUploadedAt(db *sql.DB, fileID string) (time.Time, error) {
	var uploadedAt time.Time
	query := `SELECT uploaded_at FROM files WHERE id = $1`
	err := db.QueryRowContext(context.Background(), query, fileID).Scan(&uploadedAt)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get uploaded_at: %w", err)
	}
	return uploadedAt, nil
}

// GetFileDeletedAt retrieves a file's deleted_at timestamp for verification
func GetFileDeletedAt(db *sql.DB, fileID string) (*time.Time, error) {
	var deletedAt *time.Time
	query := `SELECT deleted_at FROM files WHERE id = $1`
	err := db.QueryRowContext(context.Background(), query, fileID).Scan(&deletedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get deleted_at: %w", err)
	}
	return deletedAt, nil
}

// FileExists checks if a file exists in the database
func FileExists(db *sql.DB, fileID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM files WHERE id = $1)`
	err := db.QueryRowContext(context.Background(), query, fileID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	return exists, nil
}
