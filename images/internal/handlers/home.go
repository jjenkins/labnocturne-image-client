package handlers

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/jjenkins/labnocturne/images/internal/templates"
)

func HomeHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.Homepage()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func APIInfoHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Lab Nocturne Images API",
			"version": "0.1.0",
			"docs":    "https://images.labnocturne.com/docs",
		})
	}
}

func HealthHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	}
}

func SuccessHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.Success()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}
