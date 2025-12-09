package handlers

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/jjenkins/labnocturne/flex/internal/templates"
)

func OnboardingUsernameHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingUsername()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingUserTypeHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingUserType()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingMonetizationHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingMonetization()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingProHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingPro()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingPricingHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingPricing()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingTemplatesHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingTemplates()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingPlatformsHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingPlatforms()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingAddLinksHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingAddLinks()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func OnboardingProfileHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.OnboardingProfile()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func DashboardHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.Dashboard()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}

func ProfilePublicHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		page := templates.ProfilePublic()
		handler := adaptor.HTTPHandler(templ.Handler(page))
		return handler(c)
	}
}
