package handlers

import (
	"database/sql"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jjenkins/labnocturne/images/internal/service"
	"github.com/jjenkins/labnocturne/images/internal/store"
)

// CheckoutRequest represents the request body for POST /checkout
type CheckoutRequest struct {
	Email string `json:"email"`
	Plan  string `json:"plan"`
}

// CheckoutHandler creates a Stripe Checkout session
func CheckoutHandler(db *sql.DB, baseURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body
		var req CheckoutRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Invalid request body. Expected JSON with 'email' and 'plan' fields.",
					"type":    "invalid_request",
					"code":    "invalid_body",
				},
			})
		}

		// Validate required fields
		if req.Plan == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Plan is required. Must be 'starter' or 'pro'.",
					"type":    "validation_error",
					"code":    "missing_plan",
				},
			})
		}

		// Instantiate stores
		userStore := store.NewUserStore(db)

		// Instantiate services
		checkoutService := service.NewCheckoutService(
			userStore,
			os.Getenv("STRIPE_PRICE_ID_STARTER"),
			os.Getenv("STRIPE_PRICE_ID_PRO"),
			os.Getenv("STRIPE_WEBHOOK_SECRET"),
			baseURL,
		)

		// Create checkout session
		response, err := checkoutService.CreateCheckoutSession(c.Context(), req.Email, req.Plan)
		if err != nil {
			// Check for validation errors
			if err.Error() == "invalid plan: must be 'starter' or 'pro'" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fiber.Map{
						"message": err.Error(),
						"type":    "validation_error",
						"code":    "invalid_plan",
					},
				})
			}

			if err.Error()[:20] == "invalid email format" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Invalid email format",
						"type":    "validation_error",
						"code":    "invalid_email",
					},
				})
			}

			// Other errors are server errors
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Failed to create checkout session. Please try again.",
					"type":    "internal_error",
					"code":    "checkout_failed",
				},
			})
		}

		return c.JSON(response)
	}
}

// RetrieveKeyHandler verifies payment and returns API key
func RetrieveKeyHandler(db *sql.DB, baseURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get session_id from query parameter
		sessionID := c.Query("session_id")
		if sessionID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Missing session_id query parameter",
					"type":    "validation_error",
					"code":    "missing_session_id",
				},
			})
		}

		// Instantiate stores
		userStore := store.NewUserStore(db)

		// Instantiate services
		checkoutService := service.NewCheckoutService(
			userStore,
			os.Getenv("STRIPE_PRICE_ID_STARTER"),
			os.Getenv("STRIPE_PRICE_ID_PRO"),
			os.Getenv("STRIPE_WEBHOOK_SECRET"),
			baseURL,
		)

		// Retrieve API key (idempotent)
		response, err := checkoutService.RetrieveAPIKey(c.Context(), sessionID)
		if err != nil {
			// Check for payment not completed error
			if err.Error()[:20] == "payment not completed" {
				return c.Status(fiber.StatusPaymentRequired).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Payment has not been completed yet. Please complete payment to get your API key.",
						"type":    "payment_required",
						"code":    "payment_incomplete",
					},
				})
			}

			// Check for invalid session error
			if err.Error()[:24] == "failed to retrieve session" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Invalid session ID",
						"type":    "validation_error",
						"code":    "invalid_session",
					},
				})
			}

			// Other errors are server errors
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Failed to retrieve API key. Please try again or contact support.",
					"type":    "internal_error",
					"code":    "retrieval_failed",
				},
			})
		}

		return c.JSON(response)
	}
}

// WebhookHandler handles Stripe webhook events
func WebhookHandler(db *sql.DB, baseURL string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get webhook signature from header
		signature := c.Get("Stripe-Signature")
		if signature == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"message": "Missing Stripe-Signature header",
					"type":    "webhook_error",
					"code":    "missing_signature",
				},
			})
		}

		// Get raw request body
		payload := c.Body()

		// Instantiate stores
		userStore := store.NewUserStore(db)

		// Instantiate services
		checkoutService := service.NewCheckoutService(
			userStore,
			os.Getenv("STRIPE_PRICE_ID_STARTER"),
			os.Getenv("STRIPE_PRICE_ID_PRO"),
			os.Getenv("STRIPE_WEBHOOK_SECRET"),
			baseURL,
		)

		// Handle webhook (verifies signature internally)
		if err := checkoutService.HandleWebhook(c.Context(), payload, signature); err != nil {
			// Check for signature verification error
			if err.Error()[:33] == "failed to verify webhook signature" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fiber.Map{
						"message": "Invalid webhook signature",
						"type":    "webhook_error",
						"code":    "invalid_signature",
					},
				})
			}

			// Other errors are server errors but we still return 200 to Stripe
			// to avoid retries for unrecoverable errors
			return c.SendStatus(fiber.StatusOK)
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
