package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jjenkins/labnocturne/flex/db"
	"github.com/jjenkins/labnocturne/flex/internal/model"
)

// UserStore handles database operations for users
type UserStore struct {
	db *db.DB
}

// NewUserStore creates a new UserStore
func NewUserStore(database *db.DB) *UserStore {
	return &UserStore{db: database}
}

// GetByGoogleID retrieves a user by their Google ID
func (s *UserStore) GetByGoogleID(ctx context.Context, googleID string) (*model.User, error) {
	query := `
		SELECT id, google_id, email, name, profile_image_url, created_at, updated_at
		FROM users
		WHERE google_id = $1
	`

	var user model.User
	err := s.db.QueryRow(ctx, query, googleID).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.ProfileImageURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by google_id: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by their ID
func (s *UserStore) GetByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, google_id, email, name, profile_image_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := s.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.ProfileImageURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// Create creates a new user
func (s *UserStore) Create(ctx context.Context, googleID, email, name, profileImageURL string) (*model.User, error) {
	query := `
		INSERT INTO users (google_id, email, name, profile_image_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id, google_id, email, name, profile_image_url, created_at, updated_at
	`

	var user model.User
	err := s.db.QueryRow(ctx, query, googleID, email, name, profileImageURL).Scan(
		&user.ID,
		&user.GoogleID,
		&user.Email,
		&user.Name,
		&user.ProfileImageURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

// Update updates a user's information
func (s *UserStore) Update(ctx context.Context, id int, name, profileImageURL string) error {
	query := `
		UPDATE users
		SET name = $2, profile_image_url = $3
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query, id, name, profileImageURL)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
