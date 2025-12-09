package store

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *sql.DB {
	dsn := "postgresql://admin:admin@localhost:5432/labnocturne_images?sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Ping to ensure connection is valid
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// cleanupBandwidthStats removes all test data from bandwidth_stats table
func cleanupBandwidthStats(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM bandwidth_stats")
	if err != nil {
		t.Fatalf("Failed to cleanup bandwidth_stats: %v", err)
	}
}

func TestBandwidthStore_UpdateBandwidth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupBandwidthStats(t, db)

	store := NewBandwidthStore(db)
	ctx := context.Background()

	// Create a test user (assuming users table exists)
	var userID string
	err := db.QueryRow(`
		INSERT INTO users (api_key, key_type, plan, created_at, updated_at)
		VALUES ('test_key_123', 'test', 'test', now(), now())
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer db.Exec("DELETE FROM users WHERE id = $1", userID)

	date := time.Date(2025, 12, 3, 0, 0, 0, 0, time.UTC)

	t.Run("Insert new bandwidth record", func(t *testing.T) {
		err := store.UpdateBandwidth(ctx, userID, date, 12345, 10)
		if err != nil {
			t.Fatalf("UpdateBandwidth failed: %v", err)
		}

		// Verify the record was inserted
		var bytesServed int64
		var requestCount int
		err = db.QueryRow(`
			SELECT bytes_served, request_count
			FROM bandwidth_stats
			WHERE user_id = $1 AND date = $2
		`, userID, date).Scan(&bytesServed, &requestCount)
		if err != nil {
			t.Fatalf("Failed to query bandwidth_stats: %v", err)
		}

		if bytesServed != 12345 {
			t.Errorf("Expected bytes_served = 12345, got %d", bytesServed)
		}
		if requestCount != 10 {
			t.Errorf("Expected request_count = 10, got %d", requestCount)
		}
	})

	t.Run("Update existing bandwidth record (idempotency)", func(t *testing.T) {
		// Update with new values
		err := store.UpdateBandwidth(ctx, userID, date, 54321, 20)
		if err != nil {
			t.Fatalf("UpdateBandwidth failed: %v", err)
		}

		// Verify the record was updated (not accumulated)
		var bytesServed int64
		var requestCount int
		err = db.QueryRow(`
			SELECT bytes_served, request_count
			FROM bandwidth_stats
			WHERE user_id = $1 AND date = $2
		`, userID, date).Scan(&bytesServed, &requestCount)
		if err != nil {
			t.Fatalf("Failed to query bandwidth_stats: %v", err)
		}

		if bytesServed != 54321 {
			t.Errorf("Expected bytes_served = 54321 (last write wins), got %d", bytesServed)
		}
		if requestCount != 20 {
			t.Errorf("Expected request_count = 20 (last write wins), got %d", requestCount)
		}
	})
}

func TestBandwidthStore_GetMonthlyBandwidth(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	defer cleanupBandwidthStats(t, db)

	store := NewBandwidthStore(db)
	ctx := context.Background()

	// Create a test user
	var userID string
	err := db.QueryRow(`
		INSERT INTO users (api_key, key_type, plan, created_at, updated_at)
		VALUES ('test_key_456', 'test', 'test', now(), now())
		RETURNING id
	`).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	defer db.Exec("DELETE FROM users WHERE id = $1", userID)

	// Insert test data for December 2025
	testData := []struct {
		date         time.Time
		bytesServed  int64
		requestCount int
	}{
		{time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC), 1000, 5},
		{time.Date(2025, 12, 2, 0, 0, 0, 0, time.UTC), 2000, 10},
		{time.Date(2025, 12, 3, 0, 0, 0, 0, time.UTC), 3000, 15},
		{time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC), 5000, 25},
		{time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), 10000, 50},
	}

	for _, td := range testData {
		err := store.UpdateBandwidth(ctx, userID, td.date, td.bytesServed, td.requestCount)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	t.Run("Get total bandwidth for December", func(t *testing.T) {
		periodStart := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

		totalBytes, err := store.GetMonthlyBandwidth(ctx, userID, periodStart, periodEnd)
		if err != nil {
			t.Fatalf("GetMonthlyBandwidth failed: %v", err)
		}

		expectedTotal := int64(1000 + 2000 + 3000 + 5000 + 10000)
		if totalBytes != expectedTotal {
			t.Errorf("Expected total bytes = %d, got %d", expectedTotal, totalBytes)
		}
	})

	t.Run("Get bandwidth for partial month", func(t *testing.T) {
		periodStart := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC)

		totalBytes, err := store.GetMonthlyBandwidth(ctx, userID, periodStart, periodEnd)
		if err != nil {
			t.Fatalf("GetMonthlyBandwidth failed: %v", err)
		}

		// Should only include Dec 1-3 (Dec 10 is exclusive)
		expectedTotal := int64(1000 + 2000 + 3000)
		if totalBytes != expectedTotal {
			t.Errorf("Expected total bytes = %d, got %d", expectedTotal, totalBytes)
		}
	})

	t.Run("Get bandwidth for user with no data", func(t *testing.T) {
		// Create another user with no bandwidth data
		var emptyUserID string
		err := db.QueryRow(`
			INSERT INTO users (api_key, key_type, plan, created_at, updated_at)
			VALUES ('test_key_empty', 'test', 'test', now(), now())
			RETURNING id
		`).Scan(&emptyUserID)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		defer db.Exec("DELETE FROM users WHERE id = $1", emptyUserID)

		periodStart := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

		totalBytes, err := store.GetMonthlyBandwidth(ctx, emptyUserID, periodStart, periodEnd)
		if err != nil {
			t.Fatalf("GetMonthlyBandwidth failed: %v", err)
		}

		// Should return 0 due to COALESCE
		if totalBytes != 0 {
			t.Errorf("Expected total bytes = 0 for user with no data, got %d", totalBytes)
		}
	})

	t.Run("Get bandwidth for future period", func(t *testing.T) {
		periodStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		periodEnd := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

		totalBytes, err := store.GetMonthlyBandwidth(ctx, userID, periodStart, periodEnd)
		if err != nil {
			t.Fatalf("GetMonthlyBandwidth failed: %v", err)
		}

		// Should return 0 for future period
		if totalBytes != 0 {
			t.Errorf("Expected total bytes = 0 for future period, got %d", totalBytes)
		}
	})
}
