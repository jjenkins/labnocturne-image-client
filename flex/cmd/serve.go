package cmd

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jjenkins/labnocturne/flex/internal/handlers"
	"github.com/spf13/cobra"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Flex web server",
	Long:  `Start the web server for Flex.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use PORT env var if set, otherwise use flag value
		if envPort := os.Getenv("PORT"); envPort != "" && port == "8080" {
			port = envPort
		}

		app := fiber.New(fiber.Config{
			AppName: "Flex",
		})

		app.Use(logger.New())

		// Routes - Landing page
		app.Get("/", handlers.HomeHandler())

		// Onboarding flow
		app.Get("/onboarding/username", handlers.OnboardingUsernameHandler())
		app.Get("/onboarding/user-type", handlers.OnboardingUserTypeHandler())
		app.Get("/onboarding/monetization", handlers.OnboardingMonetizationHandler())
		app.Get("/onboarding/pro", handlers.OnboardingProHandler())
		app.Get("/onboarding/pricing", handlers.OnboardingPricingHandler())
		app.Get("/onboarding/templates", handlers.OnboardingTemplatesHandler())
		app.Get("/onboarding/platforms", handlers.OnboardingPlatformsHandler())
		app.Get("/onboarding/add-links", handlers.OnboardingAddLinksHandler())
		app.Get("/onboarding/profile", handlers.OnboardingProfileHandler())

		// Dashboard
		app.Get("/dashboard", handlers.DashboardHandler())

		// Public profile (/:username)
		app.Get("/:username", handlers.ProfilePublicHandler())

		log.Printf("Starting server on :%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
}
