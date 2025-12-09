package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jjenkins/labnocturne/flex/db"
	"github.com/jjenkins/labnocturne/flex/internal/model"
)

// TemplateStore handles database operations for templates
type TemplateStore struct {
	db *db.DB
}

// NewTemplateStore creates a new TemplateStore
func NewTemplateStore(database *db.DB) *TemplateStore {
	return &TemplateStore{db: database}
}

// GetAll retrieves all templates
func (s *TemplateStore) GetAll(ctx context.Context) ([]model.Template, error) {
	query := `
		SELECT id, name, display_name, description, background_color,
		       background_gradient, button_style, accent_color, border_radius,
		       is_pro, created_at
		FROM templates
		ORDER BY id ASC
	`

	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}
	defer rows.Close()

	var templates []model.Template
	for rows.Next() {
		var template model.Template
		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.DisplayName,
			&template.Description,
			&template.BackgroundColor,
			&template.BackgroundGradient,
			&template.ButtonStyle,
			&template.AccentColor,
			&template.BorderRadius,
			&template.IsPro,
			&template.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}
		templates = append(templates, template)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating templates: %w", err)
	}

	return templates, nil
}

// GetByID retrieves a template by ID
func (s *TemplateStore) GetByID(ctx context.Context, id int) (*model.Template, error) {
	query := `
		SELECT id, name, display_name, description, background_color,
		       background_gradient, button_style, accent_color, border_radius,
		       is_pro, created_at
		FROM templates
		WHERE id = $1
	`

	var template model.Template
	err := s.db.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.DisplayName,
		&template.Description,
		&template.BackgroundColor,
		&template.BackgroundGradient,
		&template.ButtonStyle,
		&template.AccentColor,
		&template.BorderRadius,
		&template.IsPro,
		&template.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	return &template, nil
}

// GetByName retrieves a template by name
func (s *TemplateStore) GetByName(ctx context.Context, name string) (*model.Template, error) {
	query := `
		SELECT id, name, display_name, description, background_color,
		       background_gradient, button_style, accent_color, border_radius,
		       is_pro, created_at
		FROM templates
		WHERE name = $1
	`

	var template model.Template
	err := s.db.QueryRow(ctx, query, name).Scan(
		&template.ID,
		&template.Name,
		&template.DisplayName,
		&template.Description,
		&template.BackgroundColor,
		&template.BackgroundGradient,
		&template.ButtonStyle,
		&template.AccentColor,
		&template.BorderRadius,
		&template.IsPro,
		&template.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get template by name: %w", err)
	}

	return &template, nil
}
