package model

import (
	"database/sql"
	"encoding/json"
	"time"
)

// Link represents an individual link on a page
type Link struct {
	ID           int             `json:"id"`
	PageID       int             `json:"page_id"`
	Title        string          `json:"title"`
	URL          string          `json:"url"`
	LinkType     string          `json:"link_type"` // social, custom, embed
	Platform     sql.NullString  `json:"platform"`  // instagram, youtube, tiktok, etc.
	IsVisible    bool            `json:"is_visible"`
	DisplayOrder int             `json:"display_order"`
	EmbedData    json.RawMessage `json:"embed_data"` // JSON data for embeds
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// LinkWithClicks represents a link with click count
type LinkWithClicks struct {
	Link
	ClickCount int `json:"click_count"`
}

// EmbedData represents the structure for YouTube and Spotify embeds
type EmbedData struct {
	Type     string `json:"type"`      // youtube, spotify
	VideoID  string `json:"video_id"`  // for YouTube
	TrackID  string `json:"track_id"`  // for Spotify
	PlaylistID string `json:"playlist_id"` // for Spotify playlists
}
