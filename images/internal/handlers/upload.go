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

// UploadHandler creates a handler for file uploads
func UploadHandler(db *sql.DB, s3Client *s3.Client, baseURL string, s3Bucket string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Extract Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Authorization header required. Get your API key at http://localhost:8080/key",
					"type":    "unauthorized",
					"code":    "missing_api_key",
				},
			})
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKey == authHeader {
			// No "Bearer " prefix found
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Authorization header must use Bearer scheme. Example: 'Authorization: Bearer ln_test_...'",
					"type":    "unauthorized",
					"code":    "invalid_auth_format",
				},
			})
		}

		// 2. Extract file from multipart form
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "File field required in multipart/form-data request",
					"type":    "bad_request",
					"code":    "missing_file",
				},
			})
		}

		// 3. Instantiate stores and services
		userStore := store.NewUserStore(db)
		fileStore := store.NewFileStore(db)
		uploadService := service.NewUploadService(userStore, fileStore, s3Client, baseURL, s3Bucket)

		// 4. Upload file
		uploadedFile, err := uploadService.Upload(c.Context(), apiKey, file)
		if err != nil {
			// Log the actual error for debugging
			log.Printf("Upload error: %v", err)

			// Handle specific error types
			if errors.Is(err, service.ErrInvalidAPIKey) {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Invalid API key. Generate a new key at http://localhost:8080/key",
						"type":    "unauthorized",
						"code":    "invalid_api_key",
					},
				})
			}
			if errors.Is(err, service.ErrFileTooLarge) {
				return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "File size exceeds limit for test keys (10MB). Upgrade to increase limits.",
						"type":    "file_too_large",
						"code":    "file_size_exceeded",
						"docs":    "https://images.labnocturne.com/docs#limits",
					},
				})
			}
			if errors.Is(err, service.ErrStorageExceeded) {
				return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Storage quota exceeded. Delete files or upgrade your plan.",
						"type":    "storage_quota_exceeded",
						"code":    "quota_exceeded",
						"docs":    "https://images.labnocturne.com/docs#quotas",
					},
				})
			}
			if errors.Is(err, service.ErrInvalidFileType) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "File type not supported. Allowed: jpg, png, gif, webp, svg",
						"type":    "invalid_file_type",
						"code":    "unsupported_file_type",
					},
				})
			}

			// Generic error
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Failed to upload file. Please try again.",
					"type":    "internal_error",
					"code":    "upload_failed",
				},
			})
		}

		// 5. Return success response
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":          uploadedFile.ExternalID,
			"url":         uploadedFile.CDNURL,
			"size":        uploadedFile.SizeBytes,
			"uploaded_at": uploadedFile.UploadedAt,
		})
	}
}
