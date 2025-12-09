package handlers

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/jjenkins/usds/internal/service"
)

func ConsultationHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse form data
		name := c.FormValue("name")
		email := c.FormValue("email")
		phone := c.FormValue("phone")
		serviceType := c.FormValue("service")
		message := c.FormValue("message")

		// Validate required fields
		if name == "" || email == "" || phone == "" {
			return c.Status(fiber.StatusBadRequest).SendString(
				`<div class="form-error" style="color: #d32f2f; padding: 1rem; background: #ffebee; border-radius: 4px; margin-top: 1rem; width: 100%; max-width: 100%; text-align: center;">
					Please fill in all required fields.
				</div>`,
			)
		}

		// Get spreadsheet configuration from environment
		spreadsheetID := os.Getenv("GOOGLE_SPREADSHEET_ID")
		if spreadsheetID == "" {
			log.Println("Error: GOOGLE_SPREADSHEET_ID not set")
			return c.Status(fiber.StatusInternalServerError).SendString(
				`<div class="form-error" style="color: #d32f2f; padding: 1rem; background: #ffebee; border-radius: 4px; margin-top: 1rem; width: 100%; max-width: 100%; text-align: center;">
					Configuration error. Please contact support.
				</div>`,
			)
		}

		sheetName := os.Getenv("GOOGLE_SHEET_NAME")
		if sheetName == "" {
			sheetName = "Consultations" // Default sheet name
		}

		// Create sheets service and append request asynchronously
		sheetsService := service.NewSheetsService(spreadsheetID, sheetName)

		req := service.ConsultationRequest{
			Name:    name,
			Email:   email,
			Phone:   phone,
			Service: serviceType,
			Message: message,
		}

		// Send to Google Sheets in background
		go func() {
			ctx := context.Background()
			err := sheetsService.AppendConsultationRequest(ctx, req)
			if err != nil {
				log.Printf("Error appending to sheet: %v (Name: %s, Email: %s)", err, name, email)
			} else {
				log.Printf("Successfully logged consultation request: %s (%s)", name, email)
			}
		}()

		// Return success message immediately
		return c.SendString(
			`<div class="form-success" style="color: #2e7d32; padding: 1.5rem; background: #e8f5e9; border-radius: 4px; margin-top: 1rem; text-align: center; width: 100%; max-width: 100%;">
				<h3 style="margin: 0 0 0.5rem 0; font-family: 'Cormorant Garamond', serif; font-size: 1.5rem; color: #1b5e20;">Thank You!</h3>
				<p style="margin: 0; font-size: 1rem;">Your consultation request has been received. We'll be in touch shortly!</p>
			</div>`,
		)
	}
}
