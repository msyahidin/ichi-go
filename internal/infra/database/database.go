package database

import (
	"database/sql"
	"fmt"
	"ichi-go/internal/infra/database/bun"
	"ichi-go/pkg/logger"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	upbun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

func GetDsn(dbConfig *Config) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		strconv.Itoa(dbConfig.Port),
		dbConfig.Name)

	return dsn
}

func NewBunClient(cfg *Config) (*upbun.DB, error) {
	dsn := GetDsn(cfg)

	// Open connection
	sqldb, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Test connection
	if err := sqldb.Ping(); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create Bun DB
	db := upbun.NewDB(sqldb, mysqldialect.New())

	// Set connection pool settings
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns) // fixed typo
	db.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifeTime) * time.Second)

	// Enable debug mode if configured
	if cfg.Debug {
		db.WithQueryHook(&bun.DebugHook{})
	}

	logger.Debugf("Database connection established: driver=%s, maxIdle=%d, maxOpen=%d",
		cfg.Driver, cfg.MaxIdleConns, cfg.MaxOpenConns)

	return db, nil
}
