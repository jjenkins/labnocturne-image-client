/*
Copyright © 2025 Regulation Technology Group
*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jjenkins/neighborhood/db"
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Long: `Run database migrations to set up or update the database schema.
The migrate command applies all pending migrations in the correct order.

Usage:
  go run main.go migrate
  go run main.go migrate up      (default, applies all pending migrations)
  go run main.go migrate down    (rolls back the most recent migration)
  go run main.go migrate version (displays the current migration version)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get migration direction
		direction := "up"
		if len(args) > 0 {
			direction = args[0]
		}

		if direction == "version" {
			// Display current migration version
			version, err := getCurrentMigrationVersion()
			if err != nil {
				log.Fatalf("Error getting migration version: %v", err)
			}
			fmt.Printf("Current migration version: %s\n", version)
			return
		}

		fmt.Printf("Running migrations (%s)...\n", direction)

		// Get the database connection
		database := db.Get()

		// Create migrations table if it doesn't exist
		err := createMigrationsTable(database)
		if err != nil {
			log.Fatalf("Error creating migrations table: %v", err)
		}

		if direction == "up" {
			// Run all pending migrations
			count, err := applyMigrations(database)
			if err != nil {
				log.Fatalf("Error applying migrations: %v", err)
			}
			if count == 0 {
				fmt.Println("No new migrations to apply.")
			} else {
				fmt.Printf("Successfully applied %d migration(s).\n", count)
			}
		} else if direction == "down" {
			// Roll back the most recent migration
			err := rollbackMigration(database)
			if err != nil {
				log.Fatalf("Error rolling back migration: %v", err)
			} else {
				fmt.Println("Successfully rolled back the most recent migration.")
			}
		} else {
			fmt.Printf("Unknown direction: %s. Use 'up', 'down', or 'version'.\n", direction)
		}
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
}

// createMigrationsTable creates the migrations table if it doesn't exist
func createMigrationsTable(db *db.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			version VARCHAR(255) NOT NULL,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.Exec(ctx, query)
	return err
}

// getCurrentMigrationVersion gets the current migration version from the database
func getCurrentMigrationVersion() (string, error) {
	database := db.Get()
	query := `
		SELECT version FROM migrations
		ORDER BY id DESC LIMIT 1;
	`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version string
	err := database.QueryRow(ctx, query).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return "None", nil
		}
		return "", err
	}

	return version, nil
}

// applyMigrations applies all pending migrations
func applyMigrations(db *db.DB) (int, error) {
	// Get all migration files
	migrations, err := getMigrationFiles()
	if err != nil {
		return 0, err
	}

	// Get already applied migrations
	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return 0, err
	}

	// Find pending migrations
	var pendingMigrations []string
	for _, migration := range migrations {
		version := filepath.Base(migration)
		if !contains(appliedMigrations, version) {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	// Sort migrations to ensure they're applied in the correct order
	sort.Strings(pendingMigrations)

	// Apply each pending migration
	count := 0
	for _, migration := range pendingMigrations {
		version := filepath.Base(migration)
		fmt.Printf("Applying migration: %s\n", version)

		// Read migration file
		content, err := os.ReadFile(migration)
		if err != nil {
			return count, err
		}

		// Split the file into up and down migrations
		parts := strings.Split(string(content), "-- DOWN")
		if len(parts) != 2 {
			return count, fmt.Errorf("migration file %s does not contain '-- DOWN' separator", migration)
		}

		upSQL := strings.TrimSpace(parts[0])
		if strings.HasPrefix(upSQL, "-- UP") {
			upSQL = strings.TrimSpace(strings.TrimPrefix(upSQL, "-- UP"))
		}

		// Execute the up migration within a transaction
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = db.Transaction(ctx, func(tx *sql.Tx) error {
			// Apply the migration
			_, err := tx.Exec(upSQL)
			if err != nil {
				return err
			}

			// Record the migration
			_, err = tx.Exec("INSERT INTO migrations (version) VALUES ($1)", version)
			if err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			return count, err
		}

		count++
	}

	return count, nil
}

// rollbackMigration rolls back the most recent migration
func rollbackMigration(db *db.DB) error {
	// Get the most recent migration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var version string
	err := db.QueryRow(ctx, `
		SELECT version FROM migrations
		ORDER BY id DESC LIMIT 1;
	`).Scan(&version)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("no migrations to roll back")
		}
		return err
	}

	// Find the migration file
	migrationPath := filepath.Join("db/migrations", version)
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		return err
	}

	// Split the file into up and down migrations
	parts := strings.Split(string(content), "-- DOWN")
	if len(parts) != 2 {
		return fmt.Errorf("migration file %s does not contain '-- DOWN' separator", migrationPath)
	}

	downSQL := strings.TrimSpace(parts[1])

	// Execute the down migration within a transaction
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = db.Transaction(ctx, func(tx *sql.Tx) error {
		// Apply the down migration
		_, err := tx.Exec(downSQL)
		if err != nil {
			return err
		}

		// Remove the migration record
		_, err = tx.Exec("DELETE FROM migrations WHERE version = $1", version)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

// getMigrationFiles gets all migration files from the migrations directory
func getMigrationFiles() ([]string, error) {
	var migrations []string

	err := filepath.Walk("db/migrations", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".sql") {
			migrations = append(migrations, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return migrations, nil
}

// getAppliedMigrations gets all applied migrations from the database
func getAppliedMigrations(db *db.DB) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, "SELECT version FROM migrations ORDER BY id;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		migrations = append(migrations, version)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return migrations, nil
}

// contains checks if a string slice contains a specific value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
