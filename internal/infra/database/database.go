package database

import (
	"database/sql"
	"fmt"
	"ichi-go/pkg/db/hook"
	"ichi-go/pkg/logger"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	upbun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func GetMySQLDSN(cfg *Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func GetPostgresDSN(cfg *Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func NewMySQLClient(cfg *Config) (*upbun.DB, error) {
	sqldb, err := sql.Open("mysql", GetMySQLDSN(cfg))
	if err != nil {
		return nil, fmt.Errorf("failed to open mysql connection: %w", err)
	}
	if err := sqldb.Ping(); err != nil {
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
	if err := sqldb.Ping(); err != nil {
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
