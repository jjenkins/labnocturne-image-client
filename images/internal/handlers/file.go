package handlers

import (
	"database/sql"
	"errors"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/jjenkins/labnocturne/images/internal/service"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// 1x1 transparent PNG (68 bytes)
// This is returned for all errors to provide better UX in <img> tags
var transparentPNG = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
	0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
	0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
	0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
	0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
	0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
	0x42, 0x60, 0x82,
}

// GetFileHandler returns a handler for retrieving files by ULID
func GetFileHandler(db *sql.DB, s3Client *s3.Client, baseURL string, s3Bucket string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract ULID from path parameters
		ulid := c.Params("ulid")
		ext := c.Params("ext")

		if ulid == "" || ext == "" {
			return serveErrorImage(c, fiber.StatusNotFound, "File not found")
		}

		// 2. Instantiate stores and services
		fileStore := store.NewFileStore(db)
		fileService := service.NewFileService(nil, fileStore, s3Client, baseURL, s3Bucket)

		// 3. Get file metadata
		file, err := fileService.GetByULID(c.Context(), ulid)
		if err != nil {
			if errors.Is(err, service.ErrFileNotFound) {
				return serveErrorImage(c, fiber.StatusNotFound, "File not found or has been deleted")
			}

			log.Printf("Error retrieving file %s: %v", ulid, err)
			return serveErrorImage(c, fiber.StatusInternalServerError, "Failed to retrieve file")
		}

		// 4. Generate presigned S3 URL
		presignedURL, err := fileService.GetPresignedURL(c.Context(), file.S3Key)
		if err != nil {
			log.Printf("Error generating presigned URL for %s: %v", file.S3Key, err)
			return serveErrorImage(c, fiber.StatusInternalServerError, "Failed to generate download URL")
		}

		// 5. Set cache headers and redirect
		c.Set("Cache-Control", "public, max-age=31536000, immutable")
		c.Set("Content-Type", file.MimeType)

		// 6. 302 redirect to presigned S3 URL
		return c.Redirect(presignedURL, fiber.StatusFound)
	}
}

// serveErrorImage returns a 1x1 transparent PNG for error cases
// This provides better UX when the URL is used in an <img> tag
func serveErrorImage(c *fiber.Ctx, status int, message string) error {
	log.Printf("Serving error image: %s (status: %d)", message, status)

	c.Set("Content-Type", "image/png")
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")

	return c.Status(status).Send(transparentPNG)
}

// DeleteFileHandler returns a handler for soft-deleting files
func DeleteFileHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract and validate Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Authorization header required",
					"type":    "unauthorized",
					"code":    "missing_api_key",
				},
			})
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKey == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Authorization header must use Bearer scheme",
					"type":    "unauthorized",
					"code":    "invalid_auth_format",
				},
			})
		}

		// 2. Extract file ID from path
		fileID := c.Params("id")
		if fileID == "" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "File not found",
					"type":    "not_found",
					"code":    "file_not_found",
				},
			})
		}

		// 3. Instantiate stores and services
		userStore := store.NewUserStore(db)
		fileStore := store.NewFileStore(db)
		fileService := service.NewFileService(userStore, fileStore, nil, "", "")

		// 4. Authenticate user
		user, err := userStore.FindByAPIKey(c.Context(), apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Invalid API key",
					"type":    "unauthorized",
					"code":    "invalid_api_key",
				},
			})
		}

		// 5. Delete file (service handles ownership check)
		err = fileService.DeleteFile(c.Context(), fileID, user.ID.String())
		if err != nil {
			if errors.Is(err, service.ErrFileNotFound) {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "File not found",
						"type":    "not_found",
						"code":    "file_not_found",
					},
				})
			}

			log.Printf("Error deleting file %s: %v", fileID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Failed to delete file",
					"type":    "internal_error",
					"code":    "deletion_failed",
				},
			})
		}

		// 6. Return success
		return c.JSON(fiber.Map{
			"success": true,
			"message": "File deleted successfully",
		})
	}
}

// ListFilesHandler returns a handler for listing user files with pagination
func ListFilesHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract and validate Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Invalid API key",
					"type":    "unauthorized",
					"code":    "invalid_api_key",
				},
			})
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")

		// 2. Parse query parameters
		limit := c.QueryInt("limit", 100)
		offset := c.QueryInt("offset", 0)
		sortOrder := c.Query("sort", "uploaded_at_desc")

		// 3. Instantiate stores and services
		userStore := store.NewUserStore(db)
		fileStore := store.NewFileStore(db)
		fileService := service.NewFileService(userStore, fileStore, nil, "", "")

		// 4. Authenticate user
		user, err := userStore.FindByAPIKey(c.Context(), apiKey)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Invalid API key",
					"type":    "unauthorized",
					"code":    "invalid_api_key",
				},
			})
		}

		// 5. List files
		response, err := fileService.ListFiles(c.Context(), user.ID.String(), service.ListFilesParams{
			Limit:     limit,
			Offset:    offset,
			SortOrder: sortOrder,
		})
		if err != nil {
			if errors.Is(err, service.ErrInvalidSortOrder) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Invalid sort order. Must be one of: uploaded_at_desc, uploaded_at_asc, size_desc, size_asc",
						"type":    "invalid_request",
						"code":    "invalid_sort_order",
					},
				})
			}

			log.Printf("Error listing files for user %s: %v", user.ID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Failed to list files",
					"type":    "internal_error",
					"code":    "list_failed",
				},
			})
		}

		// 6. Build full next URL with scheme and host
		var nextURL *string
		if response.NextURL != nil {
			// Get the base URL from the request
			scheme := "https"
			if c.Protocol() == "http" {
				scheme = "http"
			}
			fullURL := scheme + "://" + c.Hostname() + *response.NextURL
			nextURL = &fullURL
		}

		// 7. Format response
		fileList := make([]fiber.Map, 0, len(response.Files))
		for _, file := range response.Files {
			fileList = append(fileList, fiber.Map{
				"id":          file.ExternalID,
				"url":         file.CDNURL,
				"filename":    file.Filename,
				"size":        file.SizeBytes,
				"mime_type":   file.MimeType,
				"uploaded_at": file.UploadedAt,
			})
		}

		return c.JSON(fiber.Map{
			"files": fileList,
			"pagination": fiber.Map{
				"total":  response.Total,
				"limit":  response.Limit,
				"offset": response.Offset,
				"next":   nextURL,
			},
		})
	}
}
