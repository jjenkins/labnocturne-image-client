package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jjenkins/labnocturne/flex/db"
	"github.com/jjenkins/labnocturne/flex/internal/model"
)

// LinkStore handles database operations for links
type LinkStore struct {
	db *db.DB
}

// NewLinkStore creates a new LinkStore
func NewLinkStore(database *db.DB) *LinkStore {
	return &LinkStore{db: database}
}

// GetByPageID retrieves all links for a page, ordered by display_order
func (s *LinkStore) GetByPageID(ctx context.Context, pageID int) ([]model.Link, error) {
	query := `
		SELECT id, page_id, title, url, link_type, platform, is_visible,
		       display_order, embed_data, created_at, updated_at
		FROM links
		WHERE page_id = $1
		ORDER BY display_order ASC
	`

	rows, err := s.db.Query(ctx, query, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get links: %w", err)
	}
	defer rows.Close()

	var links []model.Link
	for rows.Next() {
		var link model.Link
		err := rows.Scan(
			&link.ID,
			&link.PageID,
			&link.Title,
			&link.URL,
			&link.LinkType,
			&link.Platform,
			&link.IsVisible,
			&link.DisplayOrder,
			&link.EmbedData,
			&link.CreatedAt,
			&link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating links: %w", err)
	}

	return links, nil
}

// GetVisibleByPageID retrieves all visible links for a page
func (s *LinkStore) GetVisibleByPageID(ctx context.Context, pageID int) ([]model.Link, error) {
	query := `
		SELECT id, page_id, title, url, link_type, platform, is_visible,
		       display_order, embed_data, created_at, updated_at
		FROM links
		WHERE page_id = $1 AND is_visible = true
		ORDER BY display_order ASC
	`

	rows, err := s.db.Query(ctx, query, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get visible links: %w", err)
	}
	defer rows.Close()

	var links []model.Link
	for rows.Next() {
		var link model.Link
		err := rows.Scan(
			&link.ID,
			&link.PageID,
			&link.Title,
			&link.URL,
			&link.LinkType,
			&link.Platform,
			&link.IsVisible,
			&link.DisplayOrder,
			&link.EmbedData,
			&link.CreatedAt,
			&link.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating links: %w", err)
	}

	return links, nil
}

// GetWithClicks retrieves all links with click counts for a page
func (s *LinkStore) GetWithClicks(ctx context.Context, pageID int) ([]model.LinkWithClicks, error) {
	query := `
		SELECT
			l.id, l.page_id, l.title, l.url, l.link_type, l.platform, l.is_visible,
			l.display_order, l.embed_data, l.created_at, l.updated_at,
			COALESCE((SELECT COUNT(*) FROM analytics_clicks WHERE link_id = l.id), 0) as click_count
		FROM links l
		WHERE l.page_id = $1
		ORDER BY l.display_order ASC
	`

	rows, err := s.db.Query(ctx, query, pageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get links with clicks: %w", err)
	}
	defer rows.Close()

	var links []model.LinkWithClicks
	for rows.Next() {
		var link model.LinkWithClicks
		err := rows.Scan(
			&link.ID,
			&link.PageID,
			&link.Title,
			&link.URL,
			&link.LinkType,
			&link.Platform,
			&link.IsVisible,
			&link.DisplayOrder,
			&link.EmbedData,
			&link.CreatedAt,
			&link.UpdatedAt,
			&link.ClickCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan link with clicks: %w", err)
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating links: %w", err)
	}

	return links, nil
}

// GetByID retrieves a link by ID
func (s *LinkStore) GetByID(ctx context.Context, id int) (*model.Link, error) {
	query := `
		SELECT id, page_id, title, url, link_type, platform, is_visible,
		       display_order, embed_data, created_at, updated_at
		FROM links
		WHERE id = $1
	`

	var link model.Link
	err := s.db.QueryRow(ctx, query, id).Scan(
		&link.ID,
		&link.PageID,
		&link.Title,
		&link.URL,
		&link.LinkType,
		&link.Platform,
		&link.IsVisible,
		&link.DisplayOrder,
		&link.EmbedData,
		&link.CreatedAt,
		&link.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get link: %w", err)
	}

	return &link, nil
}

// Create creates a new link
func (s *LinkStore) Create(ctx context.Context, link *model.Link) (*model.Link, error) {
	query := `
		INSERT INTO links (page_id, title, url, link_type, platform, is_visible, display_order, embed_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, page_id, title, url, link_type, platform, is_visible,
		          display_order, embed_data, created_at, updated_at
	`

	var newLink model.Link
	err := s.db.QueryRow(ctx, query,
		link.PageID,
		link.Title,
		link.URL,
		link.LinkType,
		link.Platform,
		link.IsVisible,
		link.DisplayOrder,
		link.EmbedData,
	).Scan(
		&newLink.ID,
		&newLink.PageID,
		&newLink.Title,
		&newLink.URL,
		&newLink.LinkType,
		&newLink.Platform,
		&newLink.IsVisible,
		&newLink.DisplayOrder,
		&newLink.EmbedData,
		&newLink.CreatedAt,
		&newLink.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create link: %w", err)
	}

	return &newLink, nil
}

// Update updates a link
func (s *LinkStore) Update(ctx context.Context, link *model.Link) error {
	query := `
		UPDATE links
		SET title = $2, url = $3, link_type = $4, platform = $5,
		    is_visible = $6, display_order = $7, embed_data = $8
		WHERE id = $1
	`

	_, err := s.db.Exec(ctx, query,
		link.ID,
		link.Title,
		link.URL,
		link.LinkType,
		link.Platform,
		link.IsVisible,
		link.DisplayOrder,
		link.EmbedData,
	)
	if err != nil {
		return fmt.Errorf("failed to update link: %w", err)
	}

	return nil
}

// Delete deletes a link
func (s *LinkStore) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM links WHERE id = $1`

	_, err := s.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete link: %w", err)
	}

	return nil
}

// Reorder updates the display order of multiple links
func (s *LinkStore) Reorder(ctx context.Context, linkOrders map[int]int) error {
	// Use a transaction to update all links atomically
	return s.db.Transaction(ctx, func(tx *sql.Tx) error {
		query := `UPDATE links SET display_order = $2 WHERE id = $1`

		for linkID, order := range linkOrders {
			_, err := tx.Exec(query, linkID, order)
			if err != nil {
				return fmt.Errorf("failed to reorder link %d: %w", linkID, err)
			}
		}

		return nil
	})
}
