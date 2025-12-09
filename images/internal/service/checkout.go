package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/mail"

	"github.com/jjenkins/labnocturne/images/internal/model"
	"github.com/jjenkins/labnocturne/images/internal/store"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/webhook"
)

// CheckoutService handles Stripe checkout and payment verification
type CheckoutService struct {
	userStore             *store.UserStore
	stripePriceIDStarter  string
	stripePriceIDPro      string
	stripeWebhookSecret   string
	appURL                string
}

// NewCheckoutService creates a new CheckoutService
func NewCheckoutService(
	userStore *store.UserStore,
	stripePriceIDStarter string,
	stripePriceIDPro string,
	stripeWebhookSecret string,
	appURL string,
) *CheckoutService {
	return &CheckoutService{
		userStore:            userStore,
		stripePriceIDStarter: stripePriceIDStarter,
		stripePriceIDPro:     stripePriceIDPro,
		stripeWebhookSecret:  stripeWebhookSecret,
		appURL:               appURL,
	}
}

// CheckoutSessionResponse represents the response from creating a checkout session
type CheckoutSessionResponse struct {
	CheckoutURL string `json:"checkout_url"`
	SessionID   string `json:"session_id"`
	Message     string `json:"message"`
}

// APIKeyResponse represents the response containing an API key
type APIKeyResponse struct {
	APIKey  string `json:"api_key"`
	Plan    string `json:"plan"`
	Message string `json:"message"`
}

// CreateCheckoutSession creates a Stripe Checkout session for purchasing a plan
func (s *CheckoutService) CreateCheckoutSession(ctx context.Context, email, plan string) (*CheckoutSessionResponse, error) {
	// Validate email format if provided
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return nil, fmt.Errorf("invalid email format: %w", err)
		}
	}

	// Validate plan and get price ID
	var priceID string
	switch plan {
	case "starter":
		priceID = s.stripePriceIDStarter
	case "pro":
		priceID = s.stripePriceIDPro
	default:
		return nil, fmt.Errorf("invalid plan: must be 'starter' or 'pro'")
	}

	// Create Stripe Checkout session
	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(s.appURL + "/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(s.appURL + "/#pricing"),
		Metadata: map[string]string{
			"plan": plan,
		},
	}

	// Set customer email if provided, otherwise Stripe will collect it
	if email != "" {
		params.CustomerEmail = stripe.String(email)
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	return &CheckoutSessionResponse{
		CheckoutURL: sess.URL,
		SessionID:   sess.ID,
		Message:     "Complete payment at the URL above to get your production API key",
	}, nil
}

// RetrieveAPIKey verifies a checkout session payment and returns the API key (idempotent)
func (s *CheckoutService) RetrieveAPIKey(ctx context.Context, sessionID string) (*APIKeyResponse, error) {
	// Retrieve session from Stripe to verify payment
	sess, err := session.Get(sessionID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	// Verify payment was successful
	if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return nil, fmt.Errorf("payment not completed: status is %s", sess.PaymentStatus)
	}

	// Extract email and plan from session
	email := sess.CustomerEmail
	if email == "" {
		email = sess.CustomerDetails.Email
	}
	if email == "" {
		return nil, fmt.Errorf("no email found in session")
	}

	plan, ok := sess.Metadata["plan"]
	if !ok || (plan != "starter" && plan != "pro") {
		return nil, fmt.Errorf("invalid or missing plan in session metadata")
	}

	// Check if user already exists (idempotency)
	existingUser, err := s.userStore.FindByEmail(ctx, email)
	if err == nil {
		// User exists, return existing key
		return &APIKeyResponse{
			APIKey:  existingUser.APIKey,
			Plan:    existingUser.Plan,
			Message: "API key retrieved successfully",
		}, nil
	}

	// Only continue if error was "no rows" (user doesn't exist)
	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Create new user with API key
	apiKey, err := GenerateAPIKey("ln_live_")
	if err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	user := &model.User{
		APIKey:           apiKey,
		Email:            &email,
		KeyType:          "live",
		Plan:             plan,
		StripeCustomerID: &sess.Customer.ID,
	}

	if err := s.userStore.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &APIKeyResponse{
		APIKey:  user.APIKey,
		Plan:    user.Plan,
		Message: fmt.Sprintf("Production API key created! Plan: %s", plan),
	}, nil
}

// HandleWebhook processes Stripe webhook events (idempotent)
func (s *CheckoutService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, signature, s.stripeWebhookSecret)
	if err != nil {
		return fmt.Errorf("failed to verify webhook signature: %w", err)
	}

	// Handle checkout.session.completed event
	if event.Type == "checkout.session.completed" {
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return fmt.Errorf("failed to unmarshal session: %w", err)
		}

		// Use the same logic as RetrieveAPIKey for idempotency
		email := sess.CustomerEmail
		if email == "" {
			email = sess.CustomerDetails.Email
		}
		if email == "" {
			return fmt.Errorf("no email found in session")
		}

		plan, ok := sess.Metadata["plan"]
		if !ok || (plan != "starter" && plan != "pro") {
			return fmt.Errorf("invalid or missing plan in session metadata")
		}

		// Check if user already exists
		_, err := s.userStore.FindByEmail(ctx, email)
		if err == nil {
			// User already exists, webhook is a duplicate (success)
			return nil
		}

		// Only continue if error was "no rows"
		if err != sql.ErrNoRows {
			return fmt.Errorf("failed to check existing user: %w", err)
		}

		// Create new user
		apiKey, err := GenerateAPIKey("ln_live_")
		if err != nil {
			return fmt.Errorf("failed to generate API key: %w", err)
		}

		user := &model.User{
			APIKey:           apiKey,
			Email:            &email,
			KeyType:          "live",
			Plan:             plan,
			StripeCustomerID: &sess.Customer.ID,
		}

		if err := s.userStore.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	return nil
}

