package database

import (
	"database/sql"
	entSql "entgo.io/ent/dialect/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	upbun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"ichi-go/internal/infra/database/bun"
	"ichi-go/internal/infra/database/ent"
	"ichi-go/internal/infra/database/enthook"
	"ichi-go/pkg/logger"
	"strconv"
	"time"
)

func GetDsn(dbConfig *Config) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		strconv.Itoa(dbConfig.Port),
		dbConfig.Name)

	return dsn
}

func NewEntClient(dbConfig *Config) *ent.Client {
	dsn := GetDsn(dbConfig)

	db, err := sql.Open(dbConfig.Driver, dsn)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}

	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetMaxOpenConns(dbConfig.MaxOpenConns)
	db.SetConnMaxLifetime(time.Duration(dbConfig.MaxConnLifeTime) * time.Second)

	drv := entSql.OpenDB("mysql", db)

	var client = &ent.Client{}
	if !dbConfig.Debug {
		client = ent.NewClient(ent.Driver(drv))
	} else {
		client = ent.NewClient(ent.Driver(drv), ent.Debug())
	}

	if drv == nil || client == nil {
		logger.Fatalf("failed opening connection to DB : driver or DB new client is null: %v", err)
	}

	// Run the auto migration tool.
	//if err := client.Schema.Create(context.Background()); err != nil {
	//	log.Fatalf("failed creating schema resources: %v", err)
	//}

	//if err != nil {
	//	log.Printf("err : %s\n", err)
	//}

	//setup hooks
	//SetupHooks(client)
	client.Use(enthook.VersionHook())
	//client.Intercept(intercept.NewRelicSegmentDb())
	//log.Info("initialized SetupHooks configuration=")

	return client
}

func NewBunClient(cfg Config) (*upbun.DB, error) {
	dsn := GetDsn(&cfg)

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
		db.AddQueryHook(&bun.DebugHook{})
	}

	logger.Debugf("Database connection established: driver=%s, maxIdle=%d, maxOpen=%d",
		cfg.Driver, cfg.MaxIdleConns, cfg.MaxOpenConns)

	return db, nil
}
