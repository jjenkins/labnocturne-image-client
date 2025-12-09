package model

import "time"

// User represents an authenticated user account
type User struct {
	ID              int       `json:"id"`
	GoogleID        string    `json:"google_id"`
	Email           string    `json:"email"`
	Name            string    `json:"name"`
	ProfileImageURL string    `json:"profile_image_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
