package handlers

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jjenkins/labnocturne/images/internal/templates"
)

// DocsHandler returns the API documentation page
func DocsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")

		// Get base URL from environment or construct from request
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			// Fallback to constructing from request
			scheme := "https"
			if c.Protocol() == "http" {
				scheme = "http"
			}
			baseURL = scheme + "://" + c.Hostname()
		}

		return templates.Docs(baseURL).Render(c.Context(), c.Response().BodyWriter())
	}
}
