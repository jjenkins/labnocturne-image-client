package model

import (
	"time"

	"github.com/google/uuid"
)

// File represents an uploaded file in the system
type File struct {
	ID         string     // ULID (uppercase canonical)
	ExternalID string     // "img_" + lowercase ULID
	UserID     uuid.UUID  // User who uploaded the file
	Filename   string     // Original filename
	Extension  string     // File extension (jpg, png, etc.)
	SizeBytes  int64      // File size in bytes
	MimeType   string     // MIME type (image/jpeg, etc.)
	S3Key      string     // Internal S3 key: "0/1/a/01ARZ3...FAV.jpg"
	CDNURL     string     // External CDN URL: "https://.../i/01ARZ3...FAV.jpg"
	UploadedAt time.Time  // Timestamp when file was uploaded
	DeletedAt  *time.Time // For soft delete (NULL if not deleted)
}
