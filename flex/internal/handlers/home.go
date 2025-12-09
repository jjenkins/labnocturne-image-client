package handlers

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/jjenkins/labnocturne/flex/internal/templates"
)

func HomeHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.Home()
		handler := adaptor.HTTPHandler(templ.Handler(page))

		return handler(c)
	}
}
