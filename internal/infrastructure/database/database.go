package database

import (
	"database/sql"
	entSql "entgo.io/ent/dialect/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"rathalos-kit/config"
	"rathalos-kit/internal/infrastructure/database/ent"
	"rathalos-kit/internal/infrastructure/database/ent/hook"
	"strconv"
	"time"
)

func NewEntClient() *ent.Client {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		config.Cfg.Database.User,
		config.Cfg.Database.Password,
		config.Cfg.Database.Host,
		strconv.Itoa(config.Cfg.Database.Port),
		config.Cfg.Database.Name)

	//log.Debug("DSN=", dsn) //for debug only

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed opening connection to DB: %v", err)
	}

	db.SetMaxIdleConns(config.Cfg.Database.MaxIdleConns)
	db.SetMaxOpenConns(config.Cfg.Database.MaxOPenConns)
	db.SetConnMaxLifetime(time.Duration(config.Cfg.Database.MaxConnLifeTime) * time.Second)

	drv := entSql.OpenDB("mysql", db)

	var client = &ent.Client{}
	appMode := config.Cfg.App.Env
	if appMode == "prod" {
		//log.Info("initialized database x sqlDb x orm ent : DEV")
		client = ent.NewClient(ent.Driver(drv))
	} else {
		//log.Info("initialized database x sqlDb x orm ent : PROD")
		client = ent.NewClient(ent.Driver(drv), ent.Debug())
	}

	if drv == nil || client == nil {
		log.Fatalf("failed opening connection to DB : driver or DB new client is null")
	}

	// Run the auto migration tool.
	//if err := client.Schema.Create(context.Background()); err != nil {
	//	log.Fatalf("failed creating schema resources: %v", err)
	//}

	if err != nil {
		log.Printf("err : %s\n", err)
	}

	//log.Info("initialized database x sqlDb x orm ent : success")

	//setup hooks
	//SetupHooks(client)
	client.Use(hook.VersionHook())
	//client.Intercept(intercept.NewRelicSegmentDb())
	//log.Info("initialized SetupHooks configuration=")

	return client
}
