package database

import (
	"database/sql"
	entSql "entgo.io/ent/dialect/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	dbConfig "ichi-go/config/database"
	"ichi-go/internal/infra/database/ent"
	"ichi-go/internal/infra/database/enthook"
	"ichi-go/pkg/logger"
	"strconv"
	"time"
)

func GetDsn(dbConfig *dbConfig.DatabaseConfig) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		strconv.Itoa(dbConfig.Port),
		dbConfig.Name)

	return dsn
}

func NewEntClient(dbConfig *dbConfig.DatabaseConfig, debug bool) *ent.Client {
	dsn := GetDsn(dbConfig)

	db, err := sql.Open(dbConfig.Driver, dsn)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}

	db.SetMaxIdleConns(dbConfig.MaxIdleConns)
	db.SetMaxOpenConns(dbConfig.MaxOPenConns)
	db.SetConnMaxLifetime(time.Duration(dbConfig.MaxConnLifeTime) * time.Second)

	drv := entSql.OpenDB("mysql", db)

	var client = &ent.Client{}
	if debug {
		client = ent.NewClient(ent.Driver(drv), ent.Debug())
	} else {
		client = ent.NewClient(ent.Driver(drv))
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
