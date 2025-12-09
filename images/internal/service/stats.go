package service

import (
	"context"
	"fmt"
	"time"

	"github.com/jjenkins/labnocturne/images/internal/model"
	"github.com/jjenkins/labnocturne/images/internal/ratelimit"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// StatsService handles usage statistics calculation
type StatsService struct {
	userStore      *store.UserStore
	bandwidthStore *store.BandwidthStore
	rateLimiter    *ratelimit.Limiter
}

// NewStatsService creates a new StatsService
func NewStatsService(userStore *store.UserStore, bandwidthStore *store.BandwidthStore, rateLimiter *ratelimit.Limiter) *StatsService {
	return &StatsService{
		userStore:      userStore,
		bandwidthStore: bandwidthStore,
		rateLimiter:    rateLimiter,
	}
}

// UsageStats represents comprehensive usage statistics for a user
type UsageStats struct {
	Storage     StorageStats    `json:"storage"`
	Files       FileStats       `json:"files"`
	Bandwidth   BandwidthStats  `json:"bandwidth"`
	APIRequests APIRequestStats `json:"api_requests"`
	Account     AccountInfo     `json:"account"`
}

// StorageStats represents storage usage information
type StorageStats struct {
	UsedBytes      int64   `json:"used_bytes"`
	UsedMB         float64 `json:"used_mb"`
	QuotaBytes     int64   `json:"quota_bytes"`
	QuotaMB        float64 `json:"quota_mb"`
	PercentageUsed float64 `json:"percentage_used"`
}

// FileStats represents file count information
type FileStats struct {
	Count int64  `json:"count"`
	Quota *int64 `json:"quota"` // nil = unlimited
}

// BandwidthStats represents bandwidth usage information
type BandwidthStats struct {
	UsedBytes      int64     `json:"used_bytes"`
	UsedMB         float64   `json:"used_mb"`
	QuotaBytes     int64     `json:"quota_bytes"`
	QuotaMB        float64   `json:"quota_mb"`
	PercentageUsed float64   `json:"percentage_used"`
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
}

// APIRequestStats represents API request rate limit information
type APIRequestStats struct {
	Count    int       `json:"count"`
	Quota    int       `json:"quota"`
	Period   string    `json:"period"`
	ResetsAt time.Time `json:"resets_at"`
}

// AccountInfo represents account metadata
type AccountInfo struct {
	KeyType   string    `json:"key_type"`
	CreatedAt time.Time `json:"created_at"`
}

// GetUsageStats retrieves comprehensive usage statistics for a user
func (s *StatsService) GetUsageStats(ctx context.Context, user *model.User) (*UsageStats, error) {
	// Get storage usage from database
	totalBytes, fileCount, err := s.userStore.GetStorageUsage(ctx, user.ID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get storage usage: %w", err)
	}

	// Determine quotas based on plan
	var storageQuota int64
	var bandwidthQuota int64
	var rateLimit int

	switch user.Plan {
	case "test":
		storageQuota = 100 * 1024 * 1024        // 100MB
		bandwidthQuota = 1 * 1024 * 1024 * 1024 // 1GB
		rateLimit = 100                          // per hour
	case "starter":
		storageQuota = 10 * 1024 * 1024 * 1024   // 10GB
		bandwidthQuota = 50 * 1024 * 1024 * 1024 // 50GB
		rateLimit = 1000                          // per hour
	case "pro":
		storageQuota = 100 * 1024 * 1024 * 1024   // 100GB
		bandwidthQuota = 500 * 1024 * 1024 * 1024 // 500GB
		rateLimit = 10000                          // per hour
	default:
		// Fallback to test quotas if plan is unknown
		storageQuota = 100 * 1024 * 1024        // 100MB
		bandwidthQuota = 1 * 1024 * 1024 * 1024 // 1GB
		rateLimit = 100                          // per hour
	}

	// Calculate storage percentage
	var storagePercent float64
	if storageQuota > 0 {
		storagePercent = float64(totalBytes) / float64(storageQuota) * 100
	}

	// Calculate current billing period (monthly)
	now := time.Now().UTC()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	// Calculate rate limit reset (next hour)
	resetsAt := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, time.UTC)

	// Get bandwidth usage from bandwidth_stats table
	bandwidthUsed, err := s.bandwidthStore.GetMonthlyBandwidth(ctx, user.ID.String(), periodStart, periodEnd)
	if err != nil {
		// Log warning but don't fail the entire stats request
		fmt.Printf("warning: failed to get bandwidth for user %s: %v\n", user.ID, err)
		bandwidthUsed = 0
	}

	// Calculate bandwidth percentage
	var bandwidthPercent float64
	if bandwidthQuota > 0 {
		bandwidthPercent = float64(bandwidthUsed) / float64(bandwidthQuota) * 100
	}

	return &UsageStats{
		Storage: StorageStats{
			UsedBytes:      totalBytes,
			UsedMB:         float64(totalBytes) / 1024 / 1024,
			QuotaBytes:     storageQuota,
			QuotaMB:        float64(storageQuota) / 1024 / 1024,
			PercentageUsed: storagePercent,
		},
		Files: FileStats{
			Count: fileCount,
			Quota: nil, // Unlimited files
		},
		Bandwidth: BandwidthStats{
			UsedBytes:      bandwidthUsed,
			UsedMB:         float64(bandwidthUsed) / 1024 / 1024,
			QuotaBytes:     bandwidthQuota,
			QuotaMB:        float64(bandwidthQuota) / 1024 / 1024,
			PercentageUsed: bandwidthPercent,
			PeriodStart:    periodStart,
			PeriodEnd:      periodEnd,
		},
		APIRequests: APIRequestStats{
			Count:    s.rateLimiter.GetRequestCount(user.APIKey),
			Quota:    rateLimit,
			Period:   "hour",
			ResetsAt: resetsAt,
		},
		Account: AccountInfo{
			KeyType:   user.KeyType,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}
