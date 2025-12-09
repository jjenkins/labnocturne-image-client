package handlers

import (
	"context"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"

func HomeHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := context.Background()

		page := templates.Home()
		handler := adaptor.HTTPHandler(templ.Handler(page))

		return handler(c)
	}
}
