package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// BandwidthStore handles bandwidth statistics database operations
type BandwidthStore struct {
	db *sql.DB
}

// NewBandwidthStore creates a new BandwidthStore
func NewBandwidthStore(db *sql.DB) *BandwidthStore {
	return &BandwidthStore{db: db}
}

// UpdateBandwidth upserts bandwidth data for a user on a specific date.
// If a record already exists for the user and date, it replaces the values (last write wins).
// This ensures idempotency when the worker is run multiple times for the same date.
func (s *BandwidthStore) UpdateBandwidth(ctx context.Context, userID string, date time.Time, bytesServed int64, requestCount int) error {
	query := `
		INSERT INTO bandwidth_stats (user_id, date, bytes_served, request_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, date)
		DO UPDATE SET
			bytes_served = EXCLUDED.bytes_served,
			request_count = EXCLUDED.request_count,
			updated_at = now()
	`
	_, err := s.db.ExecContext(ctx, query, userID, date, bytesServed, requestCount)
	if err != nil {
		return fmt.Errorf("failed to update bandwidth stats: %w", err)
	}
	return nil
}

// GetMonthlyBandwidth calculates total bandwidth for a user in the specified billing period.
// Returns 0 if no bandwidth data exists for the period.
func (s *BandwidthStore) GetMonthlyBandwidth(ctx context.Context, userID string, periodStart, periodEnd time.Time) (int64, error) {
	query := `
		SELECT COALESCE(SUM(bytes_served), 0)
		FROM bandwidth_stats
		WHERE user_id = $1 AND date >= $2 AND date < $3
	`
	var totalBytes int64
	err := s.db.QueryRowContext(ctx, query, userID, periodStart, periodEnd).Scan(&totalBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to get monthly bandwidth: %w", err)
	}
	return totalBytes, nil
}
