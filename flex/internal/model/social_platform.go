package model

import (
	"database/sql"
	"time"
)

// SocialPlatform represents a social platform connection for a page
type SocialPlatform struct {
	ID           int            `json:"id"`
	PageID       int            `json:"page_id"`
	Platform     string         `json:"platform"` // instagram, youtube, tiktok, etc.
	Username     sql.NullString `json:"username"`
	URL          sql.NullString `json:"url"`
	DisplayOrder int            `json:"display_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}
