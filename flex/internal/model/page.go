package model

import (
	"database/sql"
	"time"
)

// Page represents a user's link-in-bio page
type Page struct {
	ID                   int            `json:"id"`
	UserID               int            `json:"user_id"`
	Username             string         `json:"username"`
	DisplayName          sql.NullString `json:"display_name"`
	Bio                  sql.NullString `json:"bio"`
	ProfileImageURL      sql.NullString `json:"profile_image_url"`
	TemplateID           int            `json:"template_id"`
	IsPublished          bool           `json:"is_published"`
	UserType             string         `json:"user_type"`              // creator, brand, fan
	SelectedMonetization []string       `json:"selected_monetization"`  // array of monetization categories
	SelectedPlan         string         `json:"selected_plan"`          // free, pro, premium
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
}

// PageWithStats represents a page with analytics stats
type PageWithStats struct {
	Page
	TotalViews  int     `json:"total_views"`
	TotalClicks int     `json:"total_clicks"`
	CTR         float64 `json:"ctr"` // Click-through rate
}
