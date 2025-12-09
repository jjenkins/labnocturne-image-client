package handlers

import (
	"database/sql"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jjenkins/labnocturne/images/internal/ratelimit"
	"github.com/jjenkins/labnocturne/images/internal/service"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// StatsHandler returns usage statistics for the authenticated user
func StatsHandler(db *sql.DB, rateLimiter *ratelimit.Limiter) fiber.Handler {
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

		// 2. Instantiate stores and services
		userStore := store.NewUserStore(db)
		bandwidthStore := store.NewBandwidthStore(db)
		statsService := service.NewStatsService(userStore, bandwidthStore, rateLimiter)

		// 3. Authenticate user
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

		// 4. Get usage stats
		stats, err := statsService.GetUsageStats(c.Context(), user)
		if err != nil {
			log.Printf("Error getting stats for user %s: %v", user.ID, err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Failed to retrieve usage statistics",
					"type":    "internal_error",
					"code":    "stats_failed",
				},
			})
		}

		// 5. Return stats
		return c.JSON(stats)
	}
}
