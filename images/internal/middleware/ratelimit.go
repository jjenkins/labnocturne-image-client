package middleware

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jjenkins/labnocturne/images/internal/ratelimit"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// RateLimitMiddleware applies rate limiting to all authenticated endpoints
func RateLimitMiddleware(limiter *ratelimit.Limiter, db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip rate limiting if no auth header (public endpoints)
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next()
		}

		// Extract API key from Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid format, but let auth handler deal with it
			return c.Next()
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")

		// Get user to determine plan
		userStore := store.NewUserStore(db)
		user, err := userStore.FindByAPIKey(c.Context(), apiKey)
		if err != nil {
			// Invalid API key, let auth handler deal with it
			return c.Next()
		}

		// Check rate limit
		allowed, limit, remaining, resetAt := limiter.Allow(apiKey, user.Plan)

		// Set rate limit headers (always, even if rate limited)
		c.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		// If rate limited, return 429
		if !allowed {
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			c.Set("Retry-After", strconv.Itoa(retryAfter))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": fiber.Map{
					"message":  fmt.Sprintf("Rate limit exceeded. Limit: %d requests/hour. Resets at %s.", limit, resetAt.Format(time.RFC3339)),
					"type":     "rate_limit_exceeded",
					"code":     "rate_limit_exceeded",
					"limit":    limit,
					"reset_at": resetAt.Format(time.RFC3339),
				},
			})
		}

		return c.Next()
	}
}

// RateLimitByIP creates middleware that rate limits requests by IP address
// This is useful for public endpoints like test key generation
func RateLimitByIP(limiter *ratelimit.Limiter, name string, limit int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get client IP address
		ipAddress := c.IP()

		// Check rate limit for this IP
		allowed, limitVal, remaining, resetAt := limiter.AllowIP(ipAddress, limit, window)

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", strconv.Itoa(limitVal))
		c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		// If rate limited, return 429
		if !allowed {
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			c.Set("Retry-After", strconv.Itoa(retryAfter))

			// Log rate limit violation
			fmt.Printf("Rate limit exceeded for %s - IP: %s, Limit: %d per %v, Reset: %s\n",
				name, ipAddress, limit, window, resetAt.Format(time.RFC3339))

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": fiber.Map{
					"message":  fmt.Sprintf("Rate limit exceeded for %s. Limit: %d per %v. Try again at %s.", name, limit, window, resetAt.Format(time.RFC3339)),
					"type":     "rate_limit_exceeded",
					"code":     "rate_limit_exceeded",
					"limit":    limitVal,
					"window":   window.String(),
					"reset_at": resetAt.Format(time.RFC3339),
				},
			})
		}

		return c.Next()
	}
}
