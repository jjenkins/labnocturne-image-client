package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jjenkins/labnocturne/flex/db"
	"github.com/jjenkins/labnocturne/flex/internal/model"
	"github.com/lib/pq"
)

// PageStore handles database operations for pages
type PageStore struct {
	db *db.DB
}

// NewPageStore creates a new PageStore
func NewPageStore(database *db.DB) *PageStore {
	return &PageStore{db: database}
}

// GetByUsername retrieves a page by username
func (s *PageStore) GetByUsername(ctx context.Context, username string) (*model.Page, error) {
	query := `
		SELECT id, user_id, username, display_name, bio, profile_image_url,
		       template_id, is_published, user_type, selected_monetization,
		       selected_plan, created_at, updated_at
		FROM pages
		WHERE username = $1
	`

	var page model.Page
	err := s.db.QueryRow(ctx, query, username).Scan(
		&page.ID,
		&page.UserID,
		&page.Username,
		&page.DisplayName,
		&page.Bio,
		&page.ProfileImageURL,
		&page.TemplateID,
		&page.IsPublished,
		&page.UserType,
		pq.Array(&page.SelectedMonetization),
		&page.SelectedPlan,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Page not found
		}
		return nil, fmt.Errorf("failed to get page by username: %w", err)
	}

	return &page, nil
}

// GetByUserID retrieves a page by user ID
func (s *PageStore) GetByUserID(ctx context.Context, userID int) (*model.Page, error) {
	query := `
		SELECT id, user_id, username, display_name, bio, profile_image_url,
		       template_id, is_published, user_type, selected_monetization,
		       selected_plan, created_at, updated_at
		FROM pages
		WHERE user_id = $1
	`

	var page model.Page
	err := s.db.QueryRow(ctx, query, userID).Scan(
		&page.ID,
		&page.UserID,
		&page.Username,
		&page.DisplayName,
		&page.Bio,
		&page.ProfileImageURL,
		&page.TemplateID,
		&page.IsPublished,
		&page.UserType,
		pq.Array(&page.SelectedMonetization),
		&page.SelectedPlan,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Page not found
		}
		return nil, fmt.Errorf("failed to get page by user_id: %w", err)
	}

	return &page, nil
}

// GetWithStats retrieves a page with analytics stats
func (s *PageStore) GetWithStats(ctx context.Context, username string) (*model.PageWithStats, error) {
	query := `
		SELECT
			p.id, p.user_id, p.username, p.display_name, p.bio, p.profile_image_url,
			p.template_id, p.is_published, p.user_type, p.selected_monetization,
			p.selected_plan, p.created_at, p.updated_at,
			COALESCE((SELECT COUNT(*) FROM analytics_views WHERE page_id = p.id), 0) as total_views,
			COALESCE((SELECT COUNT(*) FROM analytics_clicks WHERE page_id = p.id), 0) as total_clicks
		FROM pages p
		WHERE p.username = $1
	`

	var pageStats model.PageWithStats
	var totalViews, totalClicks int

	err := s.db.QueryRow(ctx, query, username).Scan(
		&pageStats.ID,
		&pageStats.UserID,
		&pageStats.Username,
		&pageStats.DisplayName,
		&pageStats.Bio,
		&pageStats.ProfileImageURL,
		&pageStats.TemplateID,
		&pageStats.IsPublished,
		&pageStats.UserType,
		pq.Array(&pageStats.SelectedMonetization),
		&pageStats.SelectedPlan,
		&pageStats.CreatedAt,
		&pageStats.UpdatedAt,
		&totalViews,
		&totalClicks,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get page with stats: %w", err)
	}

	pageStats.TotalViews = totalViews
	pageStats.TotalClicks = totalClicks

	// Calculate CTR
	if totalViews > 0 {
		pageStats.CTR = float64(totalClicks) / float64(totalViews) * 100
	}

	return &pageStats, nil
}

// Create creates a new page
func (s *PageStore) Create(ctx context.Context, userID int, username string) (*model.Page, error) {
	query := `
		INSERT INTO pages (user_id, username)
		VALUES ($1, $2)
		RETURNING id, user_id, username, display_name, bio, profile_image_url,
		          template_id, is_published, user_type, selected_monetization,
		          selected_plan, created_at, updated_at
	`

	var page model.Page
	err := s.db.QueryRow(ctx, query, userID, username).Scan(
		&page.ID,
		&page.UserID,
		&page.Username,
		&page.DisplayName,
		&page.Bio,
		&page.ProfileImageURL,
		&page.TemplateID,
		&page.IsPublished,
		&page.UserType,
		pq.Array(&page.SelectedMonetization),
		&page.SelectedPlan,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create page: %w", err)
	}

	return &page, nil
}

// Update updates a page's information
func (s *PageStore) Update(ctx context.Context, page *model.Page) error {
	query := `
		UPDATE pages
		SET display_name = $2, bio = $3, profile_image_url = $4,
		    template_id = $5, is_published = $6, user_type = $7,
		    selected_monetization = $8, selected_plan = $9
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query,
		page.ID,
		page.DisplayName,
		page.Bio,
		page.ProfileImageURL,
		page.TemplateID,
		page.IsPublished,
		page.UserType,
		pq.Array(page.SelectedMonetization),
		page.SelectedPlan,
	)
	if err != nil {
		return fmt.Errorf("failed to update page: %w", err)
	}

	return nil
}

// UpdatePublishStatus updates the published status of a page
func (s *PageStore) UpdatePublishStatus(ctx context.Context, pageID int, isPublished bool) error {
	query := `UPDATE pages SET is_published = $2 WHERE id = $1`

	_, err := s.db.Exec(ctx, query, pageID, isPublished)
	if err != nil {
		return fmt.Errorf("failed to update publish status: %w", err)
	}

	return nil
}

// IsUsernameAvailable checks if a username is available
func (s *PageStore) IsUsernameAvailable(ctx context.Context, username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pages WHERE username = $1)`

	var exists bool
	err := s.db.QueryRow(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username availability: %w", err)
	}

	return !exists, nil
}
