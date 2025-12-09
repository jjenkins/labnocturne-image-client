package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"github.com/jjenkins/labnocturne/images/internal/service"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// GenerateKeyHandler creates a handler for generating test API keys
func GenerateKeyHandler(db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Instantiate stores
		userStore := store.NewUserStore(db)

		// Instantiate services
		keyService := service.NewKeyService(userStore)

		// Generate test key
		user, err := keyService.GenerateTestKey(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Failed to generate API key. Please try again.",
					"type":    "internal_error",
					"code":    "key_generation_failed",
				},
			})
		}

		// Return success response
		return c.JSON(fiber.Map{
			"api_key": user.APIKey,
			"type":    user.KeyType,
			"message": "Test key created! Files are deleted after 7 days. Upgrade at images.labnocturne.com/upgrade for permanent storage.",
			"limits": fiber.Map{
				"max_file_size_mb":       10,
				"storage_mb":             100,
				"bandwidth_gb_per_month": 1,
				"rate_limit_per_hour":    100,
			},
			"docs": "https://images.labnocturne.com/docs",
		})
	}
}
