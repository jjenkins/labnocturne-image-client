package model

import (
	"database/sql"
	"time"
)

// Template represents a preset design template
type Template struct {
	ID                 int            `json:"id"`
	Name               string         `json:"name"`
	DisplayName        string         `json:"display_name"`
	Description        sql.NullString `json:"description"`
	BackgroundColor    sql.NullString `json:"background_color"`
	BackgroundGradient sql.NullString `json:"background_gradient"`
	ButtonStyle        sql.NullString `json:"button_style"`
	AccentColor        string         `json:"accent_color"`
	BorderRadius       int            `json:"border_radius"`
	IsPro              bool           `json:"is_pro"`
	CreatedAt          time.Time      `json:"created_at"`
}
