package cmd

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jjenkins/usds/internal/handlers"
	"github.com/spf13/cobra"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the RG Golf Design web server",
	Long:  `Start the web server for RG Golf Design premium golf simulator site.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use PORT env var if set, otherwise use flag value
		if envPort := os.Getenv("PORT"); envPort != "" && port == "8080" {
			port = envPort
		}

		app := fiber.New(fiber.Config{
			AppName: "RG Golf Design",
		})

		app.Use(logger.New())

		// Routes
		app.Get("/", handlers.HomeHandler())
		app.Get("/option1", handlers.Option1Handler())
		app.Get("/option2", handlers.Option2Handler())
		app.Post("/api/consultation", handlers.ConsultationHandler())

		log.Printf("Starting RG Golf Design server on http://localhost:%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
}
