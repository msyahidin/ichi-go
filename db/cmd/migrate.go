package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
)

const (
	defaultMigrationDir = "./db/migrations/schema"
	defaultSeedDir      = "./db/migrations/seeds"
)

var (
	flags      = flag.NewFlagSet("migrate", flag.ExitOnError)
	dir        = flags.String("dir", defaultMigrationDir, "directory with migration files")
	seedDir    = flags.String("seed-dir", defaultSeedDir, "directory with seed files")
	tableSpace = flags.String("table", "", "migration type: schema (default) or data")
	env        = flags.String("env", "local", "environment (local, dev, staging, prod)")
)

func main() {
	ctx := context.Background()
	flags.Usage = usage

	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("flag parsing error: %v", err)
	}

	args := flags.Args()

	// Handle special commands
	if len(args) > 0 {
		switch args[0] {
		case "help", "-h", "--help":
			flags.Usage()
			return
		case "seed":
			handleSeedCommand(ctx, args[1:])
			return
		case "create":
			handleCreateCommand(args[1:])
			return
		}
	}

	if len(args) < 1 {
		flags.Usage()
		return
	}

	// Load config and execute migration command
	migrationDir := determineMigrationDir(*dir, *tableSpace)
	db := connectDatabase(*env)
	defer db.Close()

	executeCommand(ctx, args[0], db, migrationDir, args[1:])
}

func determineMigrationDir(baseDir, tableSpace string) string {
	if tableSpace == "data" {
		return filepath.Join("./db/migrations", "data")
	}
	if baseDir == defaultMigrationDir || tableSpace == "schema" {
		return filepath.Join("./db/migrations", "schema")
	}
	return baseDir
}

func connectDatabase(environment string) *sql.DB {
	// Load config based on environment
	// For now, using default connection
	// TODO: Integrate with your config system
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "root:password@tcp(localhost:3306)/ichigo_db?parseTime=true"
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if err := goose.SetDialect("mysql"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	return db
}

func handleCreateCommand(args []string) {
	if len(args) < 1 {
		log.Fatal("create command requires a name argument")
	}

	name := args[0]
	fileType := "sql"
	if len(args) > 1 {
		fileType = args[1]
	}

	targetDir := determineMigrationDir(*dir, *tableSpace)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		log.Fatalf("failed to create directory %s: %v", targetDir, err)
	}

	log.Printf("Creating %s migration in %s", fileType, targetDir)
	if err := goose.Run("create", nil, targetDir, name, fileType); err != nil {
		log.Fatalf("goose create failed: %v", err)
	}

	log.Printf("✅ Migration created successfully in %s", targetDir)
}

func handleSeedCommand(ctx context.Context, args []string) {
	if len(args) < 1 {
		log.Println("Available seed commands:")
		log.Println("  run [file]     - Run specific seed file or all seeds")
		log.Println("  list           - List all available seed files")
		return
	}

	subCmd := args[0]
	db := connectDatabase(*env)
	defer db.Close()

	switch subCmd {
	case "run":
		if len(args) > 1 {
			runSpecificSeed(ctx, db, args[1])
		} else {
			runAllSeeds(ctx, db)
		}
	case "list":
		listSeeds()
	default:
		log.Fatalf("unknown seed command: %s", subCmd)
	}
}

func runAllSeeds(ctx context.Context, db *sql.DB) {
	files, err := filepath.Glob(filepath.Join(*seedDir, "*.sql"))
	if err != nil {
		log.Fatalf("failed to read seed directory: %v", err)
	}

	log.Printf("🌱 Running %d seed files...", len(files))
	for _, file := range files {
		if err := executeSeedFile(ctx, db, file); err != nil {
			log.Fatalf("❌ seed failed %s: %v", filepath.Base(file), err)
		}
		log.Printf("✅ %s", filepath.Base(file))
	}
	log.Println("🎉 All seeds completed successfully")
}

func runSpecificSeed(ctx context.Context, db *sql.DB, filename string) {
	file := filepath.Join(*seedDir, filename)
	if !strings.HasSuffix(file, ".sql") {
		file = file + ".sql"
	}

	log.Printf("🌱 Running seed: %s", filename)
	if err := executeSeedFile(ctx, db, file); err != nil {
		log.Fatalf("❌ seed failed: %v", err)
	}
	log.Println("✅ Seed completed successfully")
}

func executeSeedFile(ctx context.Context, db *sql.DB, filepath string) error {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		return fmt.Errorf("failed to execute seed: %w", err)
	}

	return tx.Commit()
}

func listSeeds() {
	files, err := filepath.Glob(filepath.Join(*seedDir, "*.sql"))
	if err != nil {
		log.Fatalf("failed to read seed directory: %v", err)
	}

	log.Println("Available seed files:")
	for _, file := range files {
		log.Printf("  - %s", filepath.Base(file))
	}
}

func executeCommand(ctx context.Context, command string, db *sql.DB, dir string, arguments []string) {
	opts := []goose.OptionsFunc{
		goose.WithAllowMissing(),
	}

	log.Printf("📁 Using migration directory: %s", dir)
	log.Printf("🔧 Executing command: %s", command)

	if err := goose.RunWithOptionsContext(ctx, command, db, dir, arguments, opts...); err != nil {
		log.Fatalf("❌ goose %s failed: %v", command, err)
	}

	log.Printf("✅ Command completed successfully")
}

func usage() {
	log.Print(usagePrefix)
	flags.PrintDefaults()
	log.Print(usageCommands)
}

var (
	usagePrefix = `
🗄️  Database Migration Manager (Goose v3)

Usage: 
  go run db/cmd/migrate.go [OPTIONS] COMMAND [ARGS]

Common Examples:
  # Schema migrations
  go run db/cmd/migrate.go create create_products_table sql
  go run db/cmd/migrate.go up
  go run db/cmd/migrate.go status
  
  # Data migrations  
  go run db/cmd/migrate.go --table=data create fix_legacy_emails sql
  go run db/cmd/migrate.go --table=data up
  
  # Seeders
  go run db/cmd/migrate.go seed run
  go run db/cmd/migrate.go seed run 00_base_roles.sql
  go run db/cmd/migrate.go seed list

Options:
`

	usageCommands = `
Migration Commands:
    create NAME [sql|go]  Create new migration file
    up                    Migrate to the most recent version
    up-to VERSION         Migrate to a specific VERSION
    down                  Roll back one version
    down-to VERSION       Roll back to specific VERSION
    redo                  Re-run the latest migration
    reset                 Roll back all migrations
    status                Show migration status
    version               Print current database version
    fix                   Apply sequential ordering to migrations

Seed Commands:
    seed run [file]       Run all seeds or specific seed file
    seed list             List available seed files

Flags:
    --dir string          Migration directory (default: ./db/migrations/schema)
    --table string        Migration type: 'schema' or 'data'
    --seed-dir string     Seed directory (default: ./db/migrations/seeds)
    --env string          Environment (local, dev, staging, prod)
`
)
