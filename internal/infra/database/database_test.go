package database_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ichi-go/internal/infra/database"
)

func TestGetMySQLDSN(t *testing.T) {
	cfg := &database.Config{
		User:     "root",
		Password: "secret",
		Host:     "localhost",
		Port:     3306,
		Name:     "testdb",
	}
	dsn := database.GetMySQLDSN(cfg)
	assert.Equal(t, "root:secret@tcp(localhost:3306)/testdb?parseTime=true&multiStatements=true", dsn)
}

func TestGetPostgresDSN(t *testing.T) {
	cfg := &database.Config{
		User:     "postgres",
		Password: "secret",
		Host:     "localhost",
		Port:     5432,
		Name:     "testdb",
	}
	dsn := database.GetPostgresDSN(cfg)
	assert.Equal(t, "postgres://postgres:secret@localhost:5432/testdb?sslmode=disable", dsn)
}
