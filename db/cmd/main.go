package main

import (
	"context"
	"database/sql"
	"flag"
	"github.com/samber/do/v2"
	"ichi-go/config"
	"ichi-go/internal/infra/database"
	"log"
	"os"

	"github.com/pressly/goose/v3"

	// Init DB drivers. -- here I recommend remove unnecessary - but it's up to you
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/godrv"
	// here our migrations will live  -- use your path
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", "./db/migrations", "directory with migration files")
)

func main() {
	ctx := context.Background()
	injector := do.New()

	mainConfig := do.MustInvoke[*config.Config](injector)
	//mainConfig := config.MustLoad()
	flags.Usage = usage
	err := flags.Parse(os.Args[1:])
	if err != nil {
		return
	}

	args := flags.Args()

	if len(args) > 1 && args[0] == "run" {
		log.Printf("PROGRAM RUN\n")
		os.Exit(0)
	}

	if len(args) > 1 && args[0] == "create" {
		if err := goose.Run("create", nil, *dir, args[1:]...); err != nil {
			log.Fatalf("goose run: %v", err)
		}
		return
	}

	if len(args) < 2 {
		flags.Usage()
		return
	}

	if args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	driver, command := args[0], args[1]

	switch driver {
	case "postgres", "mysql", "sqlite3":
		if err := goose.SetDialect(driver); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("%q driver not supported\n", driver)
	}

	dbCfg := mainConfig.Database()
	dbSource := database.GetDsn(&dbCfg)
	db, err := sql.Open(driver, dbSource)
	if err != nil {
		log.Fatalf("-dbstring=%q: %v\n", dbSource, err)
	}

	executeCommand(ctx, args, command, db)
	injector.Shutdown()
}

func executeCommand(ctx context.Context, args []string, command string, db *sql.DB) {
	var arguments []string
	if len(args) > 2 {
		arguments = append(arguments, args[2:]...)
	}

	if err := goose.RunWithOptionsContext(ctx, command, db, *dir, arguments, goose.WithAllowMissing()); err != nil {
		log.Fatalf("goose run: %v", err)
	}
}

func usage() {
	log.Print(usagePrefix)
	flags.PrintDefaults()
	log.Print(usageCommands)
}

var (
	usagePrefix = `Usage: goose [OPTIONS] DRIVER DBSTRING COMMAND
Drivers:
    postgres
    mysql
    sqlite3
    redshift
Examples:
    goose sqlite3 ./foo.db status
    goose sqlite3 ./foo.db create init sql
    goose sqlite3 ./foo.db create add_some_column sql
    goose sqlite3 ./foo.db create fetch_user_data go
    goose sqlite3 ./foo.db up
    goose postgres "user=postgres dbname=postgres sslmode=disable" status
    goose mysql "user:password@/dbname?parseTime=true" status
    goose redshift "postgres://user:password@qwerty.us-east-1.redshift.amazonaws.com:5439/db"
status
Options:
`

	usageCommands = `
Commands:
    up                   Migrate the DB to the most recent version available
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql|go] Creates new migration file with next version
`
)
