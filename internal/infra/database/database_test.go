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
	assert.Equal(t, "root:secret@tcp(localhost:3306)/testdb?multiStatements=true&parseTime=true", dsn)
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

func TestGetPostgresDSN_SSLModeAndEncoding(t *testing.T) {
	cfg := &database.Config{
		User:     "post:gres",
		Password: "p@ss/word",
		Host:     "localhost",
		Port:     5432,
		Name:     "test db",
		SSLMode:  "require",
	}
	dsn := database.GetPostgresDSN(cfg)
	// url.URL encodes the userinfo and path components; verify each part is present and escaped.
	assert.Contains(t, dsn, "sslmode=require")
	assert.Contains(t, dsn, "post%3Agres")  // "post:gres" → colon encoded
	assert.Contains(t, dsn, "p%40ss%2Fword") // "p@ss/word" → @ and / encoded
	assert.Contains(t, dsn, "test%20db")    // "test db" → space encoded
	assert.Contains(t, dsn, "localhost:5432")
}
