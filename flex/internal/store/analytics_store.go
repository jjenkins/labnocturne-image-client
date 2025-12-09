package store

import (
	"context"
	"fmt"

	"github.com/jjenkins/labnocturne/flex/db"
	"github.com/jjenkins/labnocturne/flex/internal/model"
)

// AnalyticsStore handles database operations for analytics
type AnalyticsStore struct {
	db *db.DB
}

// NewAnalyticsStore creates a new AnalyticsStore
func NewAnalyticsStore(database *db.DB) *AnalyticsStore {
	return &AnalyticsStore{db: database}
}

// RecordView records a page view event
func (s *AnalyticsStore) RecordView(ctx context.Context, pageID int, userAgent, ipAddress string) error {
	query := `
		INSERT INTO analytics_views (page_id, user_agent, ip_address)
		VALUES ($1, $2, $3)
	`

	_, err := s.db.Exec(ctx, query, pageID, userAgent, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to record view: %w", err)
	}

	return nil
}

// RecordClick records a link click event
func (s *AnalyticsStore) RecordClick(ctx context.Context, linkID, pageID int, userAgent, ipAddress string) error {
	query := `
		INSERT INTO analytics_clicks (link_id, page_id, user_agent, ip_address)
		VALUES ($1, $2, $3, $4)
	`

	_, err := s.db.Exec(ctx, query, linkID, pageID, userAgent, ipAddress)
	if err != nil {
		return fmt.Errorf("failed to record click: %w", err)
	}

	return nil
}

// GetPageViewCount gets total view count for a page
func (s *AnalyticsStore) GetPageViewCount(ctx context.Context, pageID int) (int, error) {
	query := `SELECT COUNT(*) FROM analytics_views WHERE page_id = $1`

	var count int
	err := s.db.QueryRow(ctx, query, pageID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get page view count: %w", err)
	}

	return count, nil
}

// GetPageClickCount gets total click count for a page
func (s *AnalyticsStore) GetPageClickCount(ctx context.Context, pageID int) (int, error) {
	query := `SELECT COUNT(*) FROM analytics_clicks WHERE page_id = $1`

	var count int
	err := s.db.QueryRow(ctx, query, pageID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get page click count: %w", err)
	}

	return count, nil
}

// GetLinkClickCount gets click count for a specific link
func (s *AnalyticsStore) GetLinkClickCount(ctx context.Context, linkID int) (int, error) {
	query := `SELECT COUNT(*) FROM analytics_clicks WHERE link_id = $1`

	var count int
	err := s.db.QueryRow(ctx, query, linkID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get link click count: %w", err)
	}

	return count, nil
}

// GetRecentViews gets recent page views
func (s *AnalyticsStore) GetRecentViews(ctx context.Context, pageID int, limit int) ([]model.AnalyticsView, error) {
	query := `
		SELECT id, page_id, viewed_at, user_agent, ip_address
		FROM analytics_views
		WHERE page_id = $1
		ORDER BY viewed_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, pageID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent views: %w", err)
	}
	defer rows.Close()

	var views []model.AnalyticsView
	for rows.Next() {
		var view model.AnalyticsView
		err := rows.Scan(
			&view.ID,
			&view.PageID,
			&view.ViewedAt,
			&view.UserAgent,
			&view.IPAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan view: %w", err)
		}
		views = append(views, view)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating views: %w", err)
	}

	return views, nil
}

// GetRecentClicks gets recent link clicks
func (s *AnalyticsStore) GetRecentClicks(ctx context.Context, pageID int, limit int) ([]model.AnalyticsClick, error) {
	query := `
		SELECT id, link_id, page_id, clicked_at, user_agent, ip_address
		FROM analytics_clicks
		WHERE page_id = $1
		ORDER BY clicked_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(ctx, query, pageID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent clicks: %w", err)
	}
	defer rows.Close()

	var clicks []model.AnalyticsClick
	for rows.Next() {
		var click model.AnalyticsClick
		err := rows.Scan(
			&click.ID,
			&click.LinkID,
			&click.PageID,
			&click.ClickedAt,
			&click.UserAgent,
			&click.IPAddress,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan click: %w", err)
		}
		clicks = append(clicks, click)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating clicks: %w", err)
	}

	return clicks, nil
}
