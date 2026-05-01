package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	upbun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"

	"ichi-go/pkg/db/hook"
	"ichi-go/pkg/logger"
)

// GetMySQLDSN builds a DSN using the mysql driver's config builder so
// special characters in credentials are handled correctly.
func GetMySQLDSN(cfg *Config) string {
	mc := gomysql.NewConfig()
	mc.User = cfg.User
	mc.Passwd = cfg.Password
	mc.Net = "tcp"
	mc.Addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mc.DBName = cfg.Name
	mc.ParseTime = true
	mc.MultiStatements = true
	return mc.FormatDSN()
}

// GetPostgresDSN builds a postgres:// URL with properly percent-encoded credentials.
// SSLMode defaults to "disable" when the Config field is empty.
func GetPostgresDSN(cfg *Config) string {
	sslMode := cfg.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	u := &url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Password),
		Host:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Path:     "/" + cfg.Name,
		RawQuery: "sslmode=" + sslMode,
	}
	return u.String()
}

func NewMySQLClient(cfg *Config) (*upbun.DB, error) {
	sqldb, err := sql.Open("mysql", GetMySQLDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqldb.PingContext(ctx); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("failed to ping mysql: %w", err)
	}
	db := upbun.NewDB(sqldb, mysqldialect.New())
	applyPoolSettings(db, cfg)
	return db, nil
}

func NewPostgresClient(cfg *Config) (*upbun.DB, error) {
	sqldb, err := sql.Open("pgx", GetPostgresDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqldb.PingContext(ctx); err != nil {
		sqldb.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}
	db := upbun.NewDB(sqldb, pgdialect.New())
	applyPoolSettings(db, cfg)
	return db, nil
}

func applyPoolSettings(db *upbun.DB, cfg *Config) {
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifeTime) * time.Second)
	if cfg.Debug {
		db.WithQueryHook(&hook.DebugHook{})
	}
	logger.Debugf("db connection ready: driver=%s maxIdle=%d maxOpen=%d",
		cfg.Driver, cfg.MaxIdleConns, cfg.MaxOpenConns)
}
