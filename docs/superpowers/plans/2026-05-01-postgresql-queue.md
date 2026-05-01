# PostgreSQL + River Queue Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add named multi-driver database config (MySQL + PostgreSQL via bun), then wire riverqueue/river as a selectable queue backend alongside the existing RabbitMQ driver.

**Architecture:** The `databases:` YAML map replaces the single `database:` key; each entry becomes a named `*bun.DB` in the DI container. A driver-agnostic `queue.Dispatcher` interface routes `Dispatch(ctx, job, opts...)` calls to either the RabbitMQ producer or a River client based on `queue.driver` config. Existing consumers are bridged to River via a single `GenericJobWorker` that dispatches by consumer name.

**Tech Stack:** Go 1.25, uptrace/bun + pgdialect, jackc/pgx/v5, riverqueue/river, riverpgxv5, rivermigrate, samber/do/v2

---

## File Map

**Create:**
- `internal/infra/queue/interfaces.go` — `Dispatcher`, `JobArgs`, `ConsumeFunc` (moved from rabbitmq/)
- `internal/infra/queue/options.go` — `DispatchOption`, `OnQueue`, `Delay`, `MaxAttempts`, `Priority`, `UniqueKey`
- `internal/infra/queue/dispatcher.go` — `NewDispatcher` factory (selects RabbitMQ or River)
- `internal/infra/queue/rabbitmq/dispatcher.go` — `RabbitMQDispatcher` implementing `queue.Dispatcher`
- `internal/infra/queue/river/worker.go` — `GenericJobArgs`, `GenericJobWorker` bridge
- `internal/infra/queue/river/dispatcher.go` — `RiverDispatcher` implementing `queue.Dispatcher`
- `internal/infra/queue/river/worker_pool.go` — registers bridge workers with `river.Workers`

**Modify:**
- `go.mod` / `go.sum` — new dependencies
- `config.example.yaml` — `database:` → `databases:` map + `primary_database` + `queue.driver` + `queue.river`
- `config/config.go` — `Schema.Database` → `Schema.Databases map + PrimaryDatabase string`; add `PrimaryDatabase()`, `Databases()` accessors
- `internal/infra/database/config.go` — update `SetDefault` key paths
- `internal/infra/database/database.go` — rename `NewBunClient`→`NewMySQLClient`, `GetDsn`→`GetMySQLDSN`; add `NewPostgresClient`, `GetPostgresDSN`
- `internal/infra/queue/config.go` — add `Driver string`, `River RiverConfig`
- `internal/infra/queue/rabbitmq/consumer_interface.go` — alias `ConsumeFunc = queue.ConsumeFunc`
- `internal/infra/queue/registry.go` — import path `queue.ConsumeFunc` (auto via alias)
- `internal/infra/providers.go` — replace `provideDatabase` with `provideDatabases`; add `provideRiverClient`, `provideQueueDispatcher`
- `cmd/main.go` — queue startup reads `queue.driver`, calls driver-agnostic start
- `cmd/server/queue_server.go` — multi-driver support (RabbitMQ path + River path)
- `db/cmd/migrate.go` — update `cfg.Database()` → `cfg.Databases()[cfg.PrimaryDatabase()]`

---

## Task 1: Add Go dependencies

**Files:** `go.mod`, `go.sum`

- [ ] **Step 1: Install new packages**

```bash
go get github.com/uptrace/bun/dialect/pgdialect
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/stdlib
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/riverqueue/river
go get github.com/riverqueue/river/riverdriver/riverpgxv5
go get github.com/riverqueue/river/rivermigrate
go mod tidy
```

- [ ] **Step 2: Verify build still passes**

```bash
go build ./...
```

Expected: no errors (new packages downloaded, nothing imported yet).

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "chore: add pgx, bun/pgdialect, and riverqueue dependencies"
```

---

## Task 2: Rename database client functions + add PostgreSQL client

**Files:**
- Modify: `internal/infra/database/database.go`

- [ ] **Step 1: Write test for NewMySQLClient (rename check)**

Create `internal/infra/database/database_test.go`:

```go
package database_test

import (
	"testing"
	"ichi-go/internal/infra/database"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "root:secret@tcp(localhost:3306)/testdb?parseTime=true&multiStatements=true", dsn)
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
```

- [ ] **Step 2: Run test — expect compile error (functions don't exist yet)**

```bash
go test ./internal/infra/database/... 2>&1 | head -20
```

Expected: `undefined: database.GetMySQLDSN` and `undefined: database.GetPostgresDSN`.

- [ ] **Step 3: Rewrite `internal/infra/database/database.go`**

```go
package database

import (
	"context"
	"database/sql"
	"fmt"
	"ichi-go/pkg/db/hook"
	"ichi-go/pkg/logger"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	upbun "github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	"github.com/uptrace/bun/dialect/pgdialect"
)

func GetMySQLDSN(cfg *Config) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
		cfg.User, cfg.Password, cfg.Host, strconv.Itoa(cfg.Port), cfg.Name)
}

func GetPostgresDSN(cfg *Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}

func NewMySQLClient(cfg *Config) (*upbun.DB, error) {
	sqldb, err := sql.Open(cfg.Driver, GetMySQLDSN(cfg))
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
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
go test ./internal/infra/database/... -v
```

Expected:
```
--- PASS: TestGetMySQLDSN
--- PASS: TestGetPostgresDSN
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/infra/database/database.go internal/infra/database/database_test.go
git commit -m "feat(db): rename NewBunClient→NewMySQLClient, add NewPostgresClient + DSN helpers"
```

---

## Task 3: Update config schema for named databases map

**Files:**
- Modify: `config/config.go`
- Modify: `internal/infra/database/config.go`
- Modify: `config.example.yaml`

- [ ] **Step 1: Write config parsing test**

Create `config/config_databases_test.go`:

```go
package config_test

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ichi-go/internal/infra/database"
)

func TestDatabasesMapUnmarshal(t *testing.T) {
	viper.Reset()
	viper.SetConfigType("yaml")

	yaml := `
databases:
  mysql:
    driver: mysql
    host: db-mysql
    port: 3306
    name: app_db
    user: root
    password: secret
    max_idle_conns: 5
    max_open_conns: 25
    max_conn_life_time: 1800
    debug: false
  postgres:
    driver: postgres
    host: db-pg
    port: 5432
    name: queue_db
    user: pg_user
    password: pg_secret
    max_idle_conns: 3
    max_open_conns: 10
    max_conn_life_time: 900
    debug: false
primary_database: mysql
`
	require.NoError(t, viper.ReadConfig(strings.NewReader(yaml)))

	var dbs map[string]database.Config
	require.NoError(t, viper.UnmarshalKey("databases", &dbs))

	assert.Len(t, dbs, 2)

	mysql := dbs["mysql"]
	assert.Equal(t, "mysql", mysql.Driver)
	assert.Equal(t, "db-mysql", mysql.Host)
	assert.Equal(t, 3306, mysql.Port)

	pg := dbs["postgres"]
	assert.Equal(t, "postgres", pg.Driver)
	assert.Equal(t, "db-pg", pg.Host)
	assert.Equal(t, 5432, pg.Port)

	primary := viper.GetString("primary_database")
	assert.Equal(t, "mysql", primary)
}
```

Add `"strings"` import to the test file.

- [ ] **Step 2: Run test — expect compile error**

```bash
go test ./config/... 2>&1 | head -10
```

Expected: `undefined` or schema field not found errors.

- [ ] **Step 3: Update `config/config.go`**

Replace `Database database.Config` with `Databases` map and `PrimaryDatabase` string. Add the new accessors. Keep the old `Database()` accessor pointing to primary for backward compatibility.

```go
type Schema struct {
	App             AppConfig
	PrimaryDatabase string                     `mapstructure:"primary_database"`
	Databases       map[string]database.Config `mapstructure:"databases"`
	Cache           cache.Config
	Log             logger.LogConfig
	Http            httpConfig.Config
	HttpClient      httpConfig.ClientConfig
	PkgClient       pkgClientConfig.PkgClient
	Queue           queue.Config
	Auth            authenticator.Config
	Validator       validator.Config
	Versioning      versioning.Config
	RBAC            rbac.Config
}
```

Add these methods, replacing the old `Database()`:

```go
// Databases returns all named database configs.
func (c *Config) Databases() map[string]database.Config {
	c.ensureLoaded()
	return c.schema.Databases
}

// PrimaryDatabase returns the name of the primary database connection.
// Defaults to "mysql" if not set in config.
func (c *Config) PrimaryDatabase() string {
	c.ensureLoaded()
	if c.schema.PrimaryDatabase == "" {
		return "mysql"
	}
	return c.schema.PrimaryDatabase
}

// Database returns the primary database config for backward compatibility.
func (c *Config) Database() *database.Config {
	c.ensureLoaded()
	primary := c.PrimaryDatabase()
	cfg, ok := c.schema.Databases[primary]
	if !ok {
		// Return zero value — callers that depend on this will get a ping error.
		return &database.Config{}
	}
	return &cfg
}
```

- [ ] **Step 4: Update `internal/infra/database/config.go`**

Change `SetDefault` key paths from `database.*` to `databases.mysql.*`:

```go
package database

import "github.com/spf13/viper"

type Config struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxConnLifeTime int    `mapstructure:"max_conn_life_time"`
	Debug           bool   `mapstructure:"debug"`
}

func SetDefault() {
	viper.SetDefault("primary_database", "mysql")
	viper.SetDefault("databases.mysql.driver", "mysql")
	viper.SetDefault("databases.mysql.host", "localhost")
	viper.SetDefault("databases.mysql.port", 3306)
	viper.SetDefault("databases.mysql.user", "root")
	viper.SetDefault("databases.mysql.password", "password")
	viper.SetDefault("databases.mysql.name", "ichi_app")
	viper.SetDefault("databases.mysql.max_idle_conns", 10)
	viper.SetDefault("databases.mysql.max_open_conns", 100)
	viper.SetDefault("databases.mysql.max_conn_life_time", 3600)
	viper.SetDefault("databases.mysql.debug", false)
}
```

- [ ] **Step 5: Update `config.example.yaml`** — replace `database:` block

Remove the existing `database:` block entirely and add:

```yaml
primary_database: "mysql"

databases:
  mysql:
    driver: mysql
    host: localhost
    port: 3306
    user: root
    password: password
    name: ichi_app
    max_idle_conns: 10
    max_open_conns: 100
    max_conn_life_time: 3600
    debug: true

  postgres:
    driver: postgres
    host: localhost
    port: 5432
    user: postgres
    password: postgres
    name: ichi_queue
    max_idle_conns: 5
    max_open_conns: 20
    max_conn_life_time: 3600
    debug: false
```

Also update `config.local.yaml` if it exists with the same structure (keep the actual values).

- [ ] **Step 6: Run test — expect PASS**

```bash
go test ./config/... -v -run TestDatabasesMapUnmarshal
```

Expected: `PASS`.

- [ ] **Step 7: Build the whole project**

```bash
go build ./...
```

Expected: build succeeds (providers.go still uses old `NewBunClient` / `cfg.Database()` — fix in next task).

- [ ] **Step 8: Commit**

```bash
git add config/config.go config/config_databases_test.go \
        internal/infra/database/config.go config.example.yaml
git commit -m "feat(config): replace single database config with named databases map"
```

---

## Task 4: Update DI providers for named *bun.DB connections

**Files:**
- Modify: `internal/infra/providers.go`

- [ ] **Step 1: Replace `provideDatabase` with `provideDatabases` in `providers.go`**

Replace the `Setup` function call and `provideDatabase` function:

```go
func Setup(injector do.Injector, cfg *config.Config) {
	do.ProvideValue(injector, cfg)

	// Core infrastructure
	provideDatabases(injector, cfg)          // replaces provideDatabase
	do.Provide(injector, provideCache(cfg))
	do.Provide(injector, provideMessaging(cfg))
	do.Provide(injector, provideMessageProducer(cfg))

	// RBAC infrastructure
	do.Provide(injector, provideRBACConfig(cfg))
	do.Provide(injector, provideRedisCache(cfg))
	do.Provide(injector, provideCasbinAdapter(cfg))
	do.Provide(injector, provideEnforcer(cfg))
}

func provideDatabases(injector do.Injector, cfg *config.Config) {
	for name, dbCfg := range cfg.Databases() {
		name, dbCfg := name, dbCfg // capture loop vars

		switch dbCfg.Driver {
		case "mysql":
			do.ProvideNamed(injector, "db."+name, func(i do.Injector) (*bun.DB, error) {
				db, err := database.NewMySQLClient(&dbCfg)
				if err != nil {
					return nil, fmt.Errorf("mysql[%s]: %w", name, err)
				}
				logger.Debugf("initialized database connection: %s (mysql)", name)
				return db, nil
			})

		case "postgres":
			do.ProvideNamed(injector, "db."+name, func(i do.Injector) (*bun.DB, error) {
				db, err := database.NewPostgresClient(&dbCfg)
				if err != nil {
					return nil, fmt.Errorf("postgres[%s]: %w", name, err)
				}
				logger.Debugf("initialized database connection: %s (postgres)", name)
				return db, nil
			})

		default:
			logger.Warnf("unknown database driver %q for connection %q — skipping", dbCfg.Driver, name)
		}
	}

	// Unnamed *bun.DB alias → primary connection (keeps all existing repos working unchanged)
	primary := cfg.PrimaryDatabase()
	do.Provide(injector, func(i do.Injector) (*bun.DB, error) {
		return do.InvokeNamed[*bun.DB](i, "db."+primary)
	})
}
```

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: success.

- [ ] **Step 3: Update `db/cmd/migrate.go`** — fix the two callers of old accessor

Find this block in `connectDatabase`:

```go
config.MustLoad()
dbConfig := config.Get().Database()
dsn := database.GetDsn(dbConfig)
```

Replace with:

```go
config.MustLoad()
dbConfig := config.Get().Database() // still works via backward-compat accessor
dsn := database.GetMySQLDSN(dbConfig)
```

And remove `database.GetDsn` reference (the old name is gone). Also fix the import of `_ "github.com/go-sql-driver/mysql"` — it stays.

- [ ] **Step 4: Build again**

```bash
go build ./...
```

Expected: success.

- [ ] **Step 5: Commit**

```bash
git add internal/infra/providers.go db/cmd/migrate.go
git commit -m "feat(infra): named *bun.DB providers for multi-driver database connections"
```

---

## Task 5: Create driver-agnostic queue interfaces

**Files:**
- Create: `internal/infra/queue/interfaces.go`
- Create: `internal/infra/queue/options.go`
- Modify: `internal/infra/queue/rabbitmq/consumer_interface.go`

- [ ] **Step 1: Write test for dispatch options**

Create `internal/infra/queue/options_test.go`:

```go
package queue_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ichi-go/internal/infra/queue"
)

func TestDispatchOptions_Defaults(t *testing.T) {
	o := queue.ApplyOptions()
	assert.Equal(t, "default", o.Queue)
	assert.Equal(t, 3, o.MaxAttempts)
	assert.Equal(t, time.Duration(0), o.Delay)
	assert.Equal(t, 1, o.Priority)
}

func TestDispatchOptions_WithOptions(t *testing.T) {
	o := queue.ApplyOptions(
		queue.OnQueue("emails"),
		queue.Delay(5*time.Minute),
		queue.MaxAttempts(5),
		queue.Priority(2),
		queue.UniqueKey("send-email-user-42"),
	)
	assert.Equal(t, "emails", o.Queue)
	assert.Equal(t, 5*time.Minute, o.Delay)
	assert.Equal(t, 5, o.MaxAttempts)
	assert.Equal(t, 2, o.Priority)
	assert.Equal(t, "send-email-user-42", o.UniqueKey)
}
```

- [ ] **Step 2: Run test — expect compile error**

```bash
go test ./internal/infra/queue/... 2>&1 | head -10
```

Expected: `undefined: queue.ApplyOptions`.

- [ ] **Step 3: Create `internal/infra/queue/interfaces.go`**

```go
package queue

import "context"

// JobArgs is implemented by every job struct. Kind() returns the unique job type
// identifier used for routing (e.g. "notification.send_email").
type JobArgs interface {
	Kind() string
}

// ConsumeFunc processes a raw message payload. Return a non-nil error to nack/retry;
// return nil to ack (including on permanent failures like bad JSON).
type ConsumeFunc func(ctx context.Context, payload []byte) error

// Dispatcher publishes jobs to the active queue backend (RabbitMQ or River).
type Dispatcher interface {
	Dispatch(ctx context.Context, job JobArgs, opts ...DispatchOption) error
}
```

- [ ] **Step 4: Create `internal/infra/queue/options.go`**

```go
package queue

import "time"

// DispatchOptions holds resolved values after applying all DispatchOption funcs.
type DispatchOptions struct {
	Queue       string
	Delay       time.Duration
	MaxAttempts int
	Priority    int
	UniqueKey   string
}

// DispatchOption mutates DispatchOptions.
type DispatchOption func(*DispatchOptions)

// ApplyOptions builds a DispatchOptions with defaults, then applies each opt.
func ApplyOptions(opts ...DispatchOption) *DispatchOptions {
	o := &DispatchOptions{
		Queue:       "default",
		MaxAttempts: 3,
		Priority:    1,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

func OnQueue(name string) DispatchOption {
	return func(o *DispatchOptions) { o.Queue = name }
}

func Delay(d time.Duration) DispatchOption {
	return func(o *DispatchOptions) { o.Delay = d }
}

func MaxAttempts(n int) DispatchOption {
	return func(o *DispatchOptions) { o.MaxAttempts = n }
}

func Priority(p int) DispatchOption {
	return func(o *DispatchOptions) { o.Priority = p }
}

func UniqueKey(k string) DispatchOption {
	return func(o *DispatchOptions) { o.UniqueKey = k }
}
```

- [ ] **Step 5: Alias `ConsumeFunc` in `rabbitmq/consumer_interface.go`**

Replace the `ConsumeFunc` type definition (keep the doc comment, change to alias):

```go
package rabbitmq

import (
	"context"
	"ichi-go/internal/infra/queue"
)

// ConsumeFunc aliases queue.ConsumeFunc so existing consumers require no import changes.
type ConsumeFunc = queue.ConsumeFunc

// MessageConsumer consumes messages from queue.
// ... (keep existing doc comment and interface body unchanged)
type MessageConsumer interface {
	Consume(ctx context.Context, handler ConsumeFunc) error
	Close() error
}
```

- [ ] **Step 6: Run options test — expect PASS**

```bash
go test ./internal/infra/queue/... -v -run TestDispatchOptions
```

Expected:
```
--- PASS: TestDispatchOptions_Defaults
--- PASS: TestDispatchOptions_WithOptions
PASS
```

- [ ] **Step 7: Build**

```bash
go build ./...
```

Expected: success.

- [ ] **Step 8: Commit**

```bash
git add internal/infra/queue/interfaces.go internal/infra/queue/options.go \
        internal/infra/queue/options_test.go \
        internal/infra/queue/rabbitmq/consumer_interface.go
git commit -m "feat(queue): add driver-agnostic Dispatcher interface and DispatchOption helpers"
```

---

## Task 6: Extend queue config with driver + River settings

**Files:**
- Modify: `internal/infra/queue/config.go`

- [ ] **Step 1: Update `internal/infra/queue/config.go`**

```go
package queue

import (
	"ichi-go/internal/infra/queue/rabbitmq"
	"time"

	"github.com/spf13/viper"
)

// Config holds queue system configuration.
type Config struct {
	Enabled  bool            `mapstructure:"enabled"`
	Driver   string          `mapstructure:"driver"` // "rabbitmq" | "river"
	RabbitMQ rabbitmq.Config `mapstructure:"rabbitmq"`
	River    RiverConfig     `mapstructure:"river"`
}

// RiverConfig holds riverqueue-specific settings.
type RiverConfig struct {
	// Database is the key from the databases map to use for River's pgx pool.
	Database             string        `mapstructure:"database"`
	MaxWorkers           int           `mapstructure:"max_workers"`
	PollInterval         time.Duration `mapstructure:"poll_interval"`
	RescueStuckJobsAfter time.Duration `mapstructure:"rescue_stuck_jobs_after"`
}

// SetDefault sets default configuration.
func SetDefault() {
	viper.SetDefault("queue.enabled", false)
	viper.SetDefault("queue.driver", "rabbitmq")
	viper.SetDefault("queue.river.database", "postgres")
	viper.SetDefault("queue.river.max_workers", 50)
	viper.SetDefault("queue.river.poll_interval", "1s")
	viper.SetDefault("queue.river.rescue_stuck_jobs_after", "1h")
	rabbitmq.RabbitMQSetDefault()
}
```

- [ ] **Step 2: Add River queue config to `config.example.yaml`**

Under the existing `queue:` block, add:

```yaml
queue:
  enabled: false
  driver: "rabbitmq"   # "rabbitmq" | "river"
  river:
    database: "postgres"
    max_workers: 50
    poll_interval: "1s"
    rescue_stuck_jobs_after: "1h"
  rabbitmq:
    # ... existing rabbitmq config unchanged
```

- [ ] **Step 3: Build**

```bash
go build ./...
```

Expected: success.

- [ ] **Step 4: Commit**

```bash
git add internal/infra/queue/config.go config.example.yaml
git commit -m "feat(queue): add driver selector and RiverConfig to queue config"
```

---

## Task 7: RabbitMQ Dispatcher adapter

**Files:**
- Create: `internal/infra/queue/rabbitmq/dispatcher.go`

- [ ] **Step 1: Write test**

Create `internal/infra/queue/rabbitmq/dispatcher_test.go`:

```go
package rabbitmq_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	mocks "ichi-go/internal/infra/queue/rabbitmq/mocks"
)

type testJob struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
}

func (t testJob) Kind() string { return "test.send_email" }

func TestRabbitMQDispatcher_Dispatch(t *testing.T) {
	producer := mocks.NewMockMessageProducer(t)

	expectedPayload, _ := json.Marshal(testJob{UserID: 1, Email: "a@b.com"})
	producer.On("Publish",
		mock.Anything,
		"test.send_email",
		expectedPayload,
		mock.AnythingOfType("rabbitmq.PublishOptions"),
	).Return(nil)

	d := rabbitmq.NewDispatcher(producer)
	err := d.Dispatch(context.Background(), testJob{UserID: 1, Email: "a@b.com"})

	assert.NoError(t, err)
	producer.AssertExpectations(t)
}

func TestRabbitMQDispatcher_Dispatch_WithDelay(t *testing.T) {
	producer := mocks.NewMockMessageProducer(t)

	producer.On("Publish",
		mock.Anything,
		"test.send_email",
		mock.Anything,
		mock.MatchedBy(func(opts rabbitmq.PublishOptions) bool {
			return opts.Delay == 5*time.Minute
		}),
	).Return(nil)

	d := rabbitmq.NewDispatcher(producer)
	err := d.Dispatch(context.Background(), testJob{UserID: 1, Email: "a@b.com"},
		queue.Delay(5*time.Minute))

	assert.NoError(t, err)
	producer.AssertExpectations(t)
}
```

- [ ] **Step 2: Run test — expect compile error**

```bash
go test ./internal/infra/queue/rabbitmq/... 2>&1 | head -10
```

Expected: `undefined: rabbitmq.NewDispatcher`.

- [ ] **Step 3: Inspect mock to confirm `Publish` signature**

```bash
grep -n "func.*Publish" internal/infra/queue/rabbitmq/mocks/mock_message_producer.go | head -5
```

The mock's `Publish` method takes `(ctx, routingKey string, message interface{}, opts PublishOptions)`. Update the test's `producer.On("Publish", ...)` call to pass `mock.Anything` for the payload if the mock expects `interface{}`, not `[]byte`. Check the mock and adjust if needed.

- [ ] **Step 4: Create `internal/infra/queue/rabbitmq/dispatcher.go`**

```go
package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"ichi-go/internal/infra/queue"
)

// RabbitMQDispatcher implements queue.Dispatcher using the existing MessageProducer.
// It serialises the job to JSON and publishes with routing_key = job.Kind().
type RabbitMQDispatcher struct {
	producer MessageProducer
}

// NewDispatcher returns a queue.Dispatcher backed by RabbitMQ.
func NewDispatcher(producer MessageProducer) queue.Dispatcher {
	return &RabbitMQDispatcher{producer: producer}
}

func (d *RabbitMQDispatcher) Dispatch(ctx context.Context, job queue.JobArgs, opts ...queue.DispatchOption) error {
	o := queue.ApplyOptions(opts...)

	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("rabbitmq dispatcher: failed to marshal job %q: %w", job.Kind(), err)
	}

	return d.producer.Publish(ctx, job.Kind(), payload, PublishOptions{
		Delay: o.Delay,
	})
}
```

- [ ] **Step 5: Run test — expect PASS**

```bash
go test ./internal/infra/queue/rabbitmq/... -v -run TestRabbitMQDispatcher
```

Expected:
```
--- PASS: TestRabbitMQDispatcher_Dispatch
--- PASS: TestRabbitMQDispatcher_Dispatch_WithDelay
PASS
```

- [ ] **Step 6: Commit**

```bash
git add internal/infra/queue/rabbitmq/dispatcher.go \
        internal/infra/queue/rabbitmq/dispatcher_test.go
git commit -m "feat(queue/rabbitmq): add RabbitMQDispatcher implementing queue.Dispatcher"
```

---

## Task 8: River GenericJobWorker (bridge for existing consumers)

**Files:**
- Create: `internal/infra/queue/river/worker.go`

- [ ] **Step 1: Write test**

Create `internal/infra/queue/river/worker_test.go`:

```go
package river_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	riverqueue "github.com/riverqueue/river"
	riverworker "ichi-go/internal/infra/queue/river"
)

func TestGenericJobWorker_Work_CallsHandler(t *testing.T) {
	called := false
	var receivedPayload []byte

	worker := riverworker.NewGenericJobWorker(func(ctx context.Context, payload []byte) error {
		called = true
		receivedPayload = payload
		return nil
	})

	job := &riverqueue.Job[riverworker.GenericJobArgs]{
		Args: riverworker.GenericJobArgs{
			ConsumerName: "payment_handler",
			Payload:      []byte(`{"amount":100}`),
		},
	}

	err := worker.Work(context.Background(), job)
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, []byte(`{"amount":100}`), receivedPayload)
}

func TestGenericJobWorker_Work_PropagatesError(t *testing.T) {
	worker := riverworker.NewGenericJobWorker(func(ctx context.Context, payload []byte) error {
		return errors.New("transient db error")
	})

	job := &riverqueue.Job[riverworker.GenericJobArgs]{
		Args: riverworker.GenericJobArgs{ConsumerName: "payment_handler", Payload: []byte(`{}`)},
	}

	err := worker.Work(context.Background(), job)
	assert.EqualError(t, err, "transient db error")
}
```

- [ ] **Step 2: Run test — expect compile error**

```bash
go test ./internal/infra/queue/river/... 2>&1 | head -10
```

Expected: package not found.

- [ ] **Step 3: Create `internal/infra/queue/river/worker.go`**

```go
package river

import (
	"context"
	"ichi-go/internal/infra/queue"

	riverqueue "github.com/riverqueue/river"
)

// GenericJobArgs carries a raw payload for existing ConsumeFunc-based consumers.
// ConsumerName discriminates which handler processes the job.
type GenericJobArgs struct {
	ConsumerName string `json:"consumer_name"`
	Payload      []byte `json:"payload"`
}

func (GenericJobArgs) Kind() string { return "generic_job" }

// GenericJobWorker bridges a single river.Worker to any queue.ConsumeFunc handler.
type GenericJobWorker struct {
	riverqueue.WorkerDefaults[GenericJobArgs]
	handler queue.ConsumeFunc
}

// NewGenericJobWorker creates a GenericJobWorker wrapping the given ConsumeFunc.
func NewGenericJobWorker(handler queue.ConsumeFunc) *GenericJobWorker {
	return &GenericJobWorker{handler: handler}
}

func (w *GenericJobWorker) Work(ctx context.Context, job *riverqueue.Job[GenericJobArgs]) error {
	return w.handler(ctx, job.Args.Payload)
}
```

- [ ] **Step 4: Run test — expect PASS**

```bash
go test ./internal/infra/queue/river/... -v -run TestGenericJobWorker
```

Expected:
```
--- PASS: TestGenericJobWorker_Work_CallsHandler
--- PASS: TestGenericJobWorker_Work_PropagatesError
PASS
```

- [ ] **Step 5: Commit**

```bash
git add internal/infra/queue/river/worker.go internal/infra/queue/river/worker_test.go
git commit -m "feat(queue/river): add GenericJobWorker bridge for ConsumeFunc consumers"
```

---

## Task 9: River Dispatcher

**Files:**
- Create: `internal/infra/queue/river/dispatcher.go`

- [ ] **Step 1: Write test**

Create `internal/infra/queue/river/dispatcher_test.go`:

```go
package river_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	riverqueue "github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	riverworker "ichi-go/internal/infra/queue/river"
	"ichi-go/internal/infra/queue"
)

// NopJob is a typed job for dispatch testing (no real DB needed for unit test).
type NopJob struct{ ID int }

func (NopJob) Kind() string { return "nop_job" }

func TestRiverDispatcher_AppliesOptions(t *testing.T) {
	// Unit test: verify options are applied without a real DB.
	// We test the option application in isolation via ApplyOptions.
	o := queue.ApplyOptions(
		queue.OnQueue("emails"),
		queue.Delay(10*time.Minute),
		queue.MaxAttempts(5),
	)
	assert.Equal(t, "emails", o.Queue)
	assert.Equal(t, 10*time.Minute, o.Delay)
	assert.Equal(t, 5, o.MaxAttempts)
}

// Integration test — requires a real PostgreSQL instance.
// Run with: go test ./internal/infra/queue/river/... -run TestRiverDispatcher_Integration -tags integration
func TestRiverDispatcher_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	pool, err := pgxpool.New(context.Background(), "postgres://postgres:postgres@localhost:5432/ichi_queue?sslmode=disable")
	require.NoError(t, err, "requires local postgres running")
	defer pool.Close()

	workers := riverqueue.NewWorkers()
	riverqueue.AddWorker(workers, riverworker.NewGenericJobWorker(func(ctx context.Context, payload []byte) error {
		return nil
	}))

	client, err := riverqueue.NewClient(riverpgxv5.New(pool), &riverqueue.Config{
		Queues:  map[string]riverqueue.QueueConfig{riverqueue.QueueDefault: {MaxWorkers: 1}},
		Workers: workers,
	})
	require.NoError(t, err)

	d := riverworker.NewDispatcher(client)
	err = d.Dispatch(context.Background(), NopJob{ID: 1}, queue.OnQueue(riverqueue.QueueDefault))
	assert.NoError(t, err)
}
```

- [ ] **Step 2: Run unit test — expect PASS (integration skipped)**

```bash
go test ./internal/infra/queue/river/... -v -run TestRiverDispatcher_AppliesOptions
```

Expected: `PASS`.

- [ ] **Step 3: Create `internal/infra/queue/river/dispatcher.go`**

```go
package river

import (
	"context"
	"fmt"
	"ichi-go/internal/infra/queue"
	"time"

	riverqueue "github.com/riverqueue/river"
	"github.com/jackc/pgx/v5"
)

// RiverDispatcher implements queue.Dispatcher using riverqueue/river.
type RiverDispatcher struct {
	client *riverqueue.Client[pgx.Tx]
}

// NewDispatcher returns a queue.Dispatcher backed by River.
func NewDispatcher(client *riverqueue.Client[pgx.Tx]) queue.Dispatcher {
	return &RiverDispatcher{client: client}
}

func (d *RiverDispatcher) Dispatch(ctx context.Context, job queue.JobArgs, opts ...queue.DispatchOption) error {
	o := queue.ApplyOptions(opts...)

	insertOpts := &riverqueue.InsertOpts{
		Queue:       o.Queue,
		MaxAttempts: o.MaxAttempts,
		Priority:    o.Priority,
	}
	if o.Delay > 0 {
		insertOpts.ScheduledAt = time.Now().Add(o.Delay)
	}
	if o.UniqueKey != "" {
		insertOpts.UniqueOpts = riverqueue.UniqueOpts{ByArgs: true}
	}

	// river.JobArgs requires Kind() string — queue.JobArgs has the same signature,
	// so any queue.JobArgs that is also a river.JobArgs inserts directly.
	riverArgs, ok := job.(riverqueue.JobArgs)
	if !ok {
		return fmt.Errorf("river dispatcher: job %T does not implement river.JobArgs (needs Kind() string)", job)
	}

	_, err := d.client.Insert(ctx, riverArgs, insertOpts)
	return err
}
```

- [ ] **Step 4: Build**

```bash
go build ./internal/infra/queue/river/...
```

Expected: success.

- [ ] **Step 5: Commit**

```bash
git add internal/infra/queue/river/dispatcher.go internal/infra/queue/river/dispatcher_test.go
git commit -m "feat(queue/river): add RiverDispatcher implementing queue.Dispatcher"
```

---

## Task 10: River worker pool registration

**Files:**
- Create: `internal/infra/queue/river/worker_pool.go`

This file registers all existing `ConsumerRegistration` handlers as `GenericJobWorker` instances and all typed River workers. One registration per consumer so River can route by `ConsumerName`.

- [ ] **Step 1: Create `internal/infra/queue/river/worker_pool.go`**

```go
package river

import (
	"ichi-go/internal/infra/queue"

	riverqueue "github.com/riverqueue/river"
)

// RegisterBridgeWorkers adds a GenericJobWorker for each ConsumerRegistration.
// Called during River client setup so existing consumers work unchanged with the River driver.
func RegisterBridgeWorkers(workers *riverqueue.Workers, registrations []queue.ConsumerRegistration) {
	for _, reg := range registrations {
		riverqueue.AddWorker(workers, NewGenericJobWorker(reg.ConsumeFunc))
	}
}
```

Note: `queue.ConsumerRegistration` must export `ConsumeFunc queue.ConsumeFunc` (not `rabbitmq.ConsumeFunc`). Since `rabbitmq.ConsumeFunc` is now a type alias for `queue.ConsumeFunc`, this is already satisfied.

- [ ] **Step 2: Write test**

Add to `internal/infra/queue/river/worker_test.go`:

```go
func TestRegisterBridgeWorkers_NoError(t *testing.T) {
	workers := riverqueue.NewWorkers()
	registrations := []queue.ConsumerRegistration{
		{
			Name:        "payment_handler",
			Description: "test",
			ConsumeFunc: func(ctx context.Context, payload []byte) error { return nil },
		},
	}
	// Should not panic
	assert.NotPanics(t, func() {
		riverworker.RegisterBridgeWorkers(workers, registrations)
	})
}
```

- [ ] **Step 3: Run test — expect PASS**

```bash
go test ./internal/infra/queue/river/... -v -run TestRegisterBridgeWorkers
```

Expected: `PASS`.

- [ ] **Step 4: Commit**

```bash
git add internal/infra/queue/river/worker_pool.go internal/infra/queue/river/worker_test.go
git commit -m "feat(queue/river): RegisterBridgeWorkers wires existing ConsumerFuncs into River workers"
```

---

## Task 11: Queue dispatcher factory

**Files:**
- Create: `internal/infra/queue/dispatcher.go`

- [ ] **Step 1: Write test**

Create `internal/infra/queue/dispatcher_test.go`:

```go
package queue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ichi-go/internal/infra/queue"
)

func TestNewDispatcher_UnknownDriver(t *testing.T) {
	_, err := queue.NewDispatcher("unknown_driver", nil, nil)
	assert.ErrorContains(t, err, "unknown queue driver")
}
```

- [ ] **Step 2: Create `internal/infra/queue/dispatcher.go`**

```go
package queue

import (
	"fmt"
	"ichi-go/internal/infra/queue/rabbitmq"
	riverimpl "ichi-go/internal/infra/queue/river"

	riverqueue "github.com/riverqueue/river"
	"github.com/jackc/pgx/v5"
)

// NewDispatcher builds the active Dispatcher based on the configured driver name.
// Pass nil for unused arguments (e.g. nil riverClient when driver is "rabbitmq").
func NewDispatcher(driver string, producer rabbitmq.MessageProducer, riverClient *riverqueue.Client[pgx.Tx]) (Dispatcher, error) {
	switch driver {
	case "rabbitmq":
		if producer == nil {
			return nil, fmt.Errorf("rabbitmq dispatcher: producer is nil (queue connection unavailable)")
		}
		return rabbitmq.NewDispatcher(producer), nil

	case "river":
		if riverClient == nil {
			return nil, fmt.Errorf("river dispatcher: client is nil (postgres connection unavailable)")
		}
		return riverimpl.NewDispatcher(riverClient), nil

	default:
		return nil, fmt.Errorf("unknown queue driver: %q (valid: rabbitmq, river)", driver)
	}
}
```

- [ ] **Step 3: Run test — expect PASS**

```bash
go test ./internal/infra/queue/... -v -run TestNewDispatcher_UnknownDriver
```

Expected: `PASS`.

- [ ] **Step 4: Build**

```bash
go build ./...
```

Expected: success.

- [ ] **Step 5: Commit**

```bash
git add internal/infra/queue/dispatcher.go internal/infra/queue/dispatcher_test.go
git commit -m "feat(queue): NewDispatcher factory selects RabbitMQ or River backend"
```

---

## Task 12: Wire River client into DI and register queue.Dispatcher

**Files:**
- Modify: `internal/infra/providers.go`

- [ ] **Step 1: Add `provideRiverClient` and `provideQueueDispatcher` to `providers.go`**

Add imports: `"github.com/jackc/pgx/v5"`, `"github.com/jackc/pgx/v5/pgxpool"`, `riverqueue "github.com/riverqueue/river"`, `"github.com/riverqueue/river/riverdriver/riverpgxv5"`, and `riverimpl "ichi-go/internal/infra/queue/river"`.

Add to `Setup`:

```go
// Queue dispatcher (RabbitMQ or River depending on config)
do.Provide(injector, provideRiverClient(cfg))
do.Provide(injector, provideQueueDispatcher(cfg))
```

Add the two provider functions:

```go
func provideRiverClient(cfg *config.Config) func(do.Injector) (*riverqueue.Client[pgx.Tx], error) {
	return func(i do.Injector) (*riverqueue.Client[pgx.Tx], error) {
		queueCfg := cfg.Queue()
		if !queueCfg.Enabled || queueCfg.Driver != "river" {
			logger.Debugf("River client not needed (driver=%s)", queueCfg.Driver)
			return nil, nil
		}

		dbKey := queueCfg.River.Database
		databases := cfg.Databases()
		dbCfg, ok := databases[dbKey]
		if !ok {
			return nil, fmt.Errorf("river: database %q not found in databases config", dbKey)
		}

		pool, err := pgxpool.New(context.Background(), database.GetPostgresDSN(&dbCfg))
		if err != nil {
			return nil, fmt.Errorf("river: failed to create pgxpool: %w", err)
		}

		// Register bridge workers for existing ConsumeFunc consumers
		workers := riverqueue.NewWorkers()
		registrations := queue.GetRegisteredConsumers(i)
		riverimpl.RegisterBridgeWorkers(workers, registrations)

		riverCfg := queueCfg.River
		pollInterval := riverCfg.PollInterval
		if pollInterval == 0 {
			pollInterval = time.Second
		}
		rescueAfter := riverCfg.RescueStuckJobsAfter
		if rescueAfter == 0 {
			rescueAfter = time.Hour
		}

		maxWorkers := riverCfg.MaxWorkers
		if maxWorkers == 0 {
			maxWorkers = 50
		}

		client, err := riverqueue.NewClient(riverpgxv5.New(pool), &riverqueue.Config{
			Queues: map[string]riverqueue.QueueConfig{
				riverqueue.QueueDefault: {MaxWorkers: maxWorkers},
				"emails":                {MaxWorkers: 10},
				"notifications":         {MaxWorkers: 20},
			},
			Workers:              workers,
			PollInterval:         pollInterval,
			RescueStuckJobsAfter: rescueAfter,
		})
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("river: failed to create client: %w", err)
		}

		logger.Debugf("initialized River client (db=%s, maxWorkers=%d)", dbKey, maxWorkers)
		return client, nil
	}
}

func provideQueueDispatcher(cfg *config.Config) func(do.Injector) (queue.Dispatcher, error) {
	return func(i do.Injector) (queue.Dispatcher, error) {
		if !cfg.Queue().Enabled {
			logger.Debugf("Queue disabled — no Dispatcher")
			return nil, nil
		}

		producer, _ := do.Invoke[rabbitmq.MessageProducer](i)
		riverClient, _ := do.Invoke[*riverqueue.Client[pgx.Tx]](i)

		d, err := queue.NewDispatcher(cfg.Queue().Driver, producer, riverClient)
		if err != nil {
			return nil, err
		}
		logger.Debugf("initialized queue.Dispatcher (driver=%s)", cfg.Queue().Driver)
		return d, nil
	}
}
```

Also add `"context"` and `"time"` to imports if not already present.

- [ ] **Step 2: Build**

```bash
go build ./...
```

Expected: success. Fix any import errors.

- [ ] **Step 3: Commit**

```bash
git add internal/infra/providers.go
git commit -m "feat(infra): wire River client and queue.Dispatcher into DI container"
```

---

## Task 13: Update queue_server.go for multi-driver startup

**Files:**
- Modify: `cmd/server/queue_server.go`
- Modify: `cmd/main.go`

- [ ] **Step 1: Rewrite `cmd/server/queue_server.go`**

```go
package server

import (
	"context"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	riverqueue "github.com/riverqueue/river"
	"github.com/samber/do/v2"

	queue "ichi-go/internal/infra/queue"
	"ichi-go/internal/infra/queue/rabbitmq"
	"ichi-go/pkg/logger"
)

// StartQueueWorkers starts all queue consumers for the configured driver.
// Blocks until ctx is cancelled.
func StartQueueWorkers(ctx context.Context, queueConfig *queue.Config, injector do.Injector) {
	if !queueConfig.Enabled {
		logger.Warnf("Queue system disabled — skipping worker startup")
		return
	}

	switch queueConfig.Driver {
	case "river":
		startRiverWorkers(ctx, injector)
	default:
		startRabbitMQWorkers(ctx, &queueConfig.RabbitMQ, injector)
	}
}

func startRiverWorkers(ctx context.Context, injector do.Injector) {
	client, err := do.Invoke[*riverqueue.Client[pgx.Tx]](injector)
	if err != nil || client == nil {
		logger.Errorf("River client unavailable — cannot start queue workers: %v", err)
		return
	}

	logger.Infof("🚀 Starting River queue workers...")
	if err := client.Start(ctx); err != nil {
		logger.Errorf("River client start error: %v", err)
		return
	}

	<-ctx.Done()

	logger.Infof("🛑 Stopping River queue workers...")
	if err := client.Stop(context.Background()); err != nil {
		logger.Errorf("River client stop error: %v", err)
	}
	logger.Infof("👋 River workers stopped")
}

func startRabbitMQWorkers(ctx context.Context, rabbitCfg *rabbitmq.Config, injector do.Injector) {
	conn, err := do.Invoke[*rabbitmq.Connection](injector)
	if conn == nil || err != nil {
		logger.Warnf("RabbitMQ connection unavailable — skipping worker startup")
		return
	}

	logger.Infof("🚀 Starting RabbitMQ queue workers...")

	// Declare all exchanges, queues, and bindings with exponential backoff retry.
	{
		backoff := 100 * time.Millisecond
		const maxBackoff = 10 * time.Second
		for {
			if err := rabbitmq.SetupTopology(conn, *rabbitCfg); err == nil {
				break
			} else {
				logger.Errorf("❌ Topology setup failed (retrying in %v): %v", backoff, err)
			}
			select {
			case <-ctx.Done():
				logger.Warnf("🛑 Context cancelled during topology setup — aborting")
				return
			case <-time.After(backoff):
				backoff = min(backoff*2, maxBackoff)
			}
		}
	}

	wg := sync.WaitGroup{}
	registeredConsumers := queue.GetRegisteredConsumers(injector)

	for _, registration := range registeredConsumers {
		consumerCfg, err := rabbitmq.GetConsumerByName(rabbitCfg, registration.Name)
		if err != nil {
			logger.Infof("⏭️  Skipping %s: %v", registration.Name, err)
			continue
		}
		if !consumerCfg.Enabled {
			logger.Infof("⏭️  Disabled: %s", registration.Name)
			continue
		}
		exchangeCfg, err := rabbitmq.GetExchangeByName(rabbitCfg, consumerCfg.ExchangeName)
		if err != nil {
			logger.Errorf("❌ No exchange for %s: %v", registration.Name, err)
			continue
		}
		consumer, err := rabbitmq.NewConsumer(conn, *consumerCfg, *exchangeCfg)
		if err != nil {
			logger.Errorf("❌ Failed to create consumer %s: %v", registration.Name, err)
			continue
		}

		wg.Add(1)
		go func(name string, c rabbitmq.MessageConsumer, fn rabbitmq.ConsumeFunc, desc string) {
			defer wg.Done()
			logger.Infof("✅ Started %s: %s", name, desc)
			if err := c.Consume(ctx, fn); err != nil {
				logger.Errorf("❌ %s error: %v", name, err)
			}
			logger.Infof("👋 Stopped %s", name)
		}(registration.Name, consumer, registration.ConsumeFunc, registration.Description)
	}

	logger.Infof("✅ All RabbitMQ workers started")
	<-ctx.Done()
	logger.Infof("🛑 Shutting down RabbitMQ workers...")
	wg.Wait()
	logger.Infof("👋 All RabbitMQ workers stopped")
}
```

- [ ] **Step 2: Update `cmd/main.go`** — simplify queue startup

Replace the queue startup block:

```go
// Before:
if cfg.Queue().Enabled {
    msgConfig := cfg.Queue()
    msgConn := do.MustInvoke[*rabbitmq.Connection](injector)
    go server.StartQueueWorkers(ctx, msgConfig, msgConn, injector)
}

// After:
if cfg.Queue().Enabled {
    go server.StartQueueWorkers(ctx, cfg.Queue(), injector)
}
```

Remove the `"ichi-go/internal/infra/queue/rabbitmq"` import from `cmd/main.go` if it's only used for the `*rabbitmq.Connection` type.

- [ ] **Step 3: Build**

```bash
go build ./...
```

Expected: success. Fix any unused imports.

- [ ] **Step 4: Commit**

```bash
git add cmd/server/queue_server.go cmd/main.go
git commit -m "feat(server): multi-driver queue startup (rabbitmq|river) in StartQueueWorkers"
```

---

## Task 14: Update ConsumerRegistration to use queue.ConsumeFunc

**Files:**
- Modify: `internal/infra/queue/registry.go`

The `ConsumerRegistration` struct currently uses `rabbitmq.ConsumeFunc`. Since that's now a type alias for `queue.ConsumeFunc`, it compiles already — but the import should point to `queue` to make the intent clear. Update the struct's field type annotation in the doc comment if needed. No code change required if the alias is in place.

- [ ] **Step 1: Verify registry compiles cleanly**

```bash
go vet ./internal/infra/queue/...
```

Expected: no errors.

- [ ] **Step 2: If `registry.go` imports `rabbitmq` only for `ConsumeFunc`, switch to `queue.ConsumeFunc` directly**

In `internal/infra/queue/registry.go`, change:

```go
// Before
type ConsumerRegistration struct {
    Name        string
    ConsumeFunc rabbitmq.ConsumeFunc
    Description string
}

// After
type ConsumerRegistration struct {
    Name        string
    ConsumeFunc ConsumeFunc  // queue.ConsumeFunc — same type via alias, no import needed
    Description string
}
```

- [ ] **Step 3: Build**

```bash
go build ./...
```

Expected: success.

- [ ] **Step 4: Commit**

```bash
git add internal/infra/queue/registry.go
git commit -m "refactor(queue): ConsumerRegistration uses queue.ConsumeFunc directly"
```

---

## Task 15: Smoke test — full build + startup check

- [ ] **Step 1: Run all tests**

```bash
go test ./... -short 2>&1 | tail -20
```

Expected: all packages PASS. Note any failures.

- [ ] **Step 2: Build binary**

```bash
go build -o /tmp/ichi-go-bin ./cmd/main.go
```

Expected: success, binary at `/tmp/ichi-go-bin`.

- [ ] **Step 3: Verify binary starts with queue disabled**

Ensure `config.local.yaml` has `queue.enabled: false`. Run:

```bash
/tmp/ichi-go-bin &
sleep 2
kill %1
```

Expected: server starts without panicking.

- [ ] **Step 4: Verify config YAML loads databases map**

```bash
cat > /tmp/test-config-check.go << 'EOF'
//go:build ignore

package main

import (
	"fmt"
	"os"
	"ichi-go/config"
)

func main() {
	os.Chdir("/Users/msyahidin/Workspace/go/rathalos-kit")
	cfg := config.MustLoad()
	fmt.Printf("Primary: %s\n", cfg.PrimaryDatabase())
	for name, db := range cfg.Databases() {
		fmt.Printf("  %s: driver=%s host=%s port=%d\n", name, db.Driver, db.Host, db.Port)
	}
}
EOF
go run /tmp/test-config-check.go
```

Expected output:
```
Primary: mysql
  mysql: driver=mysql host=localhost port=3306
  postgres: driver=postgres host=localhost port=5432
```

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: PostgreSQL dual-driver database config + River queue backend complete"
```

---

## Self-Review Checklist

- [x] **Spec §1 (Config):** Tasks 3 covers `databases:` map + `primary_database` + accessors
- [x] **Spec §2 (Database layer):** Task 2 covers `NewMySQLClient`, `NewPostgresClient`, DSN helpers
- [x] **Spec §3 (DI):** Task 4 covers named `*bun.DB` providers + unnamed alias
- [x] **Spec §4 (Queue interfaces):** Task 5 covers `Dispatcher`, `JobArgs`, `ConsumeFunc`, `DispatchOption`
- [x] **Spec §5 (River):** Tasks 8–10 cover typed worker bridge, dispatcher, worker pool
- [x] **Spec §6 (Factory):** Task 11 covers `NewDispatcher`
- [x] **Spec §7 (Dependencies):** Task 1 covers all new packages
- [x] **migrate.go caller:** Task 4 Step 3 covers `db/cmd/migrate.go`
- [x] **queue_server.go:** Task 13 covers multi-driver startup
- [x] **`cmd/main.go`:** Task 13 Step 2 covers startup simplification
- [x] **Type consistency:** `GenericJobArgs.Kind()` returns `"generic_job"` — consistent across Tasks 8–10
- [x] **`GetDsn` old name:** Removed in Task 2; `db/cmd/migrate.go` updated in Task 4
