package model

import (
	"database/sql"
	"time"
)

// AnalyticsView represents a page view event
type AnalyticsView struct {
	ID        int            `json:"id"`
	PageID    int            `json:"page_id"`
	ViewedAt  time.Time      `json:"viewed_at"`
	UserAgent sql.NullString `json:"user_agent"`
	IPAddress sql.NullString `json:"ip_address"`
}

// AnalyticsClick represents a link click event
type AnalyticsClick struct {
	ID        int            `json:"id"`
	LinkID    int            `json:"link_id"`
	PageID    int            `json:"page_id"`
	ClickedAt time.Time      `json:"clicked_at"`
	UserAgent sql.NullString `json:"user_agent"`
	IPAddress sql.NullString `json:"ip_address"`
}
