/*
Package db provides database connectivity and operations for the FDA Compliance Verification Framework.
It manages PostgreSQL connections, connection pooling, and transaction handling.
*/
package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/lib/pq" // Import PostgreSQL driver

	"github.com/jjenkins/neighborhood/config"
)

// DB encapsulates the database connection
type DB struct {
	*sql.DB
}

var (
	db   *DB
	once sync.Once
)

// Get returns a singleton instance of the database connection
// This ensures we only have one database connection pool throughout the application
func Get() *DB {
	once.Do(func() {
		var err error
		db, err = connect()
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
	})

	return db
}

// connect establishes a connection to the database
func connect() (*DB, error) {
	cfg := config.Get()

	// Construct PostgreSQL connection string from individual config fields
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Dbname,
	)

	// Open connection to PostgreSQL
	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdle)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Verify connection with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		// Close the connection if ping fails
		sqlDB.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("Successfully connected to database")

	return &DB{sqlDB}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

// Transaction executes the given function within a transaction
// It handles commit/rollback automatically based on the function's result
func (db *DB) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			// A panic occurred, rollback and re-panic
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		// Error occurred, rollback
		if rbErr := tx.Rollback(); rbErr != nil {
			// Return a combined error message
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	// All good, commit
	return tx.Commit()
}

// Exec executes a query without returning any rows
// Wraps sql.DB.ExecContext with context for timeout/cancellation support
func (db *DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
// Wraps sql.DB.QueryContext with context for timeout/cancellation support
func (db *DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row
// Wraps sql.DB.QueryRowContext with context for timeout/cancellation support
func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return db.QueryRowContext(ctx, query, args...)
}
