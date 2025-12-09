package cmd

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jjenkins/usds/internal/handlers"
	"github.com/jjenkins/usds/internal/store"
	"github.com/spf13/cobra"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the eCFR Analyzer web server",
	Long:  `Start the web server to analyze Federal Regulations from the eCFR.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use PORT env var if set, otherwise use flag value
		if envPort := os.Getenv("PORT"); envPort != "" && port == "8080" {
			port = envPort
		}

		// Database connection
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			dsn = "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"
		}

		db, err := store.NewDB(dsn)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		app := fiber.New(fiber.Config{
			AppName: "Lab Nocturne Framework",
		})

		app.Use(logger.New())

		// Routes
		app.Get("/", handlers.HomeHandler())

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
