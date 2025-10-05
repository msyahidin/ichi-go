package database

import (
	"database/sql"
	entSql "entgo.io/ent/dialect/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	upbun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"ichi-go/config"
	"ichi-go/internal/infra/database/bun"
	"ichi-go/internal/infra/database/ent"
	"ichi-go/internal/infra/database/enthook"
	"ichi-go/pkg/logger"
	"strconv"
	"time"
)

func GetDsn() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.Database().User,
		config.Database().Password,
		config.Database().Host,
		strconv.Itoa(config.Database().Port),
		config.Database().Name)

	return dsn
}

func NewEntClient() *ent.Client {
	dsn := GetDsn()

	db, err := sql.Open(config.Database().Driver, dsn)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}

	db.SetMaxIdleConns(config.Database().MaxIdleConns)
	db.SetMaxOpenConns(config.Database().MaxOPenConns)
	db.SetConnMaxLifetime(time.Duration(config.Database().MaxConnLifeTime) * time.Second)

	drv := entSql.OpenDB("mysql", db)

	var client = &ent.Client{}
	appMode := config.App().Env
	if appMode == "prod" {
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

func NewBunClient() *upbun.DB {
	dsn := GetDsn()

	sqldb, err := sql.Open(config.Database().Driver, dsn)
	if err != nil {
		logger.Fatalf("failed to connect to database: %v", err)
	}

	db := upbun.NewDB(sqldb, mysqldialect.New())

	db.SetMaxIdleConns(config.Database().MaxIdleConns)
	db.SetMaxOpenConns(config.Database().MaxOPenConns)
	db.SetConnMaxLifetime(time.Duration(config.Database().MaxConnLifeTime) * time.Second)

	if err != nil {
		logger.Fatalf("failed opening connection to DB : driver or DB new client is null: %v", err)
	}
	if config.Database().Debug {
		db.AddQueryHook(&bun.DebugHook{})
	}
	return db
}
