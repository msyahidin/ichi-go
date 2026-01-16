// Package testutil provides test container utilities for integration testing
package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Container Configuration
// ============================================================================

// MySQLContainer represents a MySQL test container
type MySQLContainer struct {
	Host     string
	Port     string
	Database string
	Username string
	Password string
	DSN      string
	cleanup  func()
}

// ============================================================================
// MySQL Test Container Setup
// ============================================================================

// SetupMySQLContainer creates a MySQL container for integration testing
// Note: This is a basic implementation. For production use, consider using testcontainers-go
func SetupMySQLContainer(t *testing.T) *MySQLContainer {
	t.Helper()

	// For now, use local MySQL instance for testing
	// You can enhance this to use Docker containers with testcontainers-go
	container := &MySQLContainer{
		Host:     "localhost",
		Port:     "3306",
		Database: "ichi_go_test",
		Username: "root",
		Password: "root",
	}

	container.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=UTC",
		container.Username,
		container.Password,
		container.Host,
		container.Port,
		container.Database,
	)

	// Test the connection
	db, err := sql.Open("mysql", container.DSN)
	require.NoError(t, err, "Failed to connect to MySQL")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	require.NoError(t, err, "Failed to ping MySQL")

	// Setup cleanup
	container.cleanup = func() {
		if db != nil {
			db.Close()
		}
	}

	t.Cleanup(container.cleanup)

	return container
}

// GetDB returns a database connection for the container
func (c *MySQLContainer) GetDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("mysql", c.DSN)
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

// ============================================================================
// Redis Test Container
// ============================================================================

// RedisContainer represents a Redis test container
type RedisContainer struct {
	Host    string
	Port    string
	cleanup func()
}

// SetupRedisContainer creates a Redis container for integration testing
func SetupRedisContainer(t *testing.T) *RedisContainer {
	t.Helper()

	// For now, use local Redis instance for testing
	container := &RedisContainer{
		Host: "localhost",
		Port: "6379",
	}

	// Setup cleanup if needed
	container.cleanup = func() {
		// Cleanup logic
	}

	t.Cleanup(container.cleanup)

	return container
}

// GetAddress returns the Redis address
func (c *RedisContainer) GetAddress() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// ============================================================================
// RabbitMQ Test Container
// ============================================================================

// RabbitMQContainer represents a RabbitMQ test container
type RabbitMQContainer struct {
	Host     string
	Port     string
	Username string
	Password string
	URL      string
	cleanup  func()
}

// SetupRabbitMQContainer creates a RabbitMQ container for integration testing
func SetupRabbitMQContainer(t *testing.T) *RabbitMQContainer {
	t.Helper()

	// For now, use local RabbitMQ instance for testing
	container := &RabbitMQContainer{
		Host:     "localhost",
		Port:     "5672",
		Username: "guest",
		Password: "guest",
	}

	container.URL = fmt.Sprintf("amqp://%s:%s@%s:%s/",
		container.Username,
		container.Password,
		container.Host,
		container.Port,
	)

	// Setup cleanup
	container.cleanup = func() {
		// Cleanup logic
	}

	t.Cleanup(container.cleanup)

	return container
}

// ============================================================================
// Database Transaction Helpers
// ============================================================================

// WithTransaction executes a function within a database transaction and rolls back
// This is useful for tests that need to write to database but not persist changes
func WithTransaction(t *testing.T, db *sql.DB, fn func(*sql.Tx) error) {
	t.Helper()

	tx, err := db.Begin()
	require.NoError(t, err, "Failed to begin transaction")

	defer func() {
		// Always rollback in tests
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			t.Logf("Failed to rollback transaction: %v", err)
		}
	}()

	err = fn(tx)
	require.NoError(t, err, "Transaction function failed")
}

// ============================================================================
// Database Seed Helpers
// ============================================================================

// SeedDatabase seeds the database with test data
func SeedDatabase(t *testing.T, db *sql.DB, seedFunc func(*sql.DB) error) {
	t.Helper()
	err := seedFunc(db)
	require.NoError(t, err, "Failed to seed database")

	// Cleanup: truncate tables after test
	t.Cleanup(func() {
		TruncateTables(t, db)
	})
}

// TruncateTables truncates all test tables
func TruncateTables(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"orders",
		"users",
		// Add more tables as needed
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s", table))
		if err != nil {
			t.Logf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// ============================================================================
// Wait for Dependencies
// ============================================================================

// WaitForMySQL waits for MySQL to be ready
func WaitForMySQL(t *testing.T, dsn string, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Timeout waiting for MySQL to be ready")
		case <-ticker.C:
			db, err := sql.Open("mysql", dsn)
			if err != nil {
				continue
			}
			defer db.Close()

			if err := db.PingContext(ctx); err == nil {
				return
			}
		}
	}
}

// WaitForRedis waits for Redis to be ready
func WaitForRedis(t *testing.T, address string, timeout time.Duration) {
	t.Helper()
	// Implementation depends on your Redis client
	// For now, this is a placeholder
	time.Sleep(1 * time.Second)
}

// WaitForRabbitMQ waits for RabbitMQ to be ready
func WaitForRabbitMQ(t *testing.T, url string, timeout time.Duration) {
	t.Helper()
	// Implementation depends on your RabbitMQ client
	// For now, this is a placeholder
	time.Sleep(1 * time.Second)
}
