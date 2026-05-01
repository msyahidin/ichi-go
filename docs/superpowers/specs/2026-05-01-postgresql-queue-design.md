# PostgreSQL Support + River Queue Design

**Date:** 2026-05-01  
**Updated:** 2026-05-02 (reflects actual Phase A implementation)  
**Branch:** feat/newqueue → main  
**Status:** Phase A complete; Phase B approved for implementation

---

## Overview

Two sequential phases:

- **Phase A (complete):** Add named multi-driver database config (MySQL + PostgreSQL), then wire riverqueue/river as a selectable queue backend alongside the existing RabbitMQ/AMQP driver.
- **Phase B (now):** Migrate all application domain tables from MySQL to PostgreSQL. The config infrastructure is already in place — this phase creates Postgres-compatible migrations, validates model compatibility, and runs a full integration smoke test.

RabbitMQ is kept as a selectable driver — no removal.

---

## 1. Config Structure

### 1.1 YAML shape

Both `database` and `queue` follow the same **Laravel-style named connections map** pattern: a `default` key names the active connection, and `connections` is a map of named entries.

```yaml
database:
  default: "mysql"              # which connection is the default *bun.DB
  connections:
    mysql:
      driver: "mysql"
      host: localhost
      port: 3306
      name: ichi_app
      user: root
      password: password
      max_idle_conns: 10
      max_open_conns: 100
      max_conn_life_time: 3600
      debug: true

    postgres:
      driver: "postgres"
      host: localhost
      port: 5432
      name: ichi_queue
      user: postgres
      password: postgres
      max_idle_conns: 5
      max_open_conns: 20
      max_conn_life_time: 3600
      debug: false

queue:
  default: "amqp"               # which connection handles dispatch when none specified
  connections:
    amqp:
      enabled: true
      driver: "amqp"
      amqp:
        # ... RabbitMQ config (unchanged)

    database:
      enabled: false
      driver: "database"
      database:
        connection: "postgres"  # key from database.connections
        max_workers: 50
        poll_interval: "1s"
        rescue_stuck_jobs_after: "1h"
```

### 1.2 Go config structs

**`config/config.go`:**

```go
type DatabaseSchema struct {
    Default     string                     `mapstructure:"default"`
    Connections map[string]database.Config `mapstructure:"connections"`
}

type Schema struct {
    App      AppConfig
    Database DatabaseSchema    `mapstructure:"database"`
    Cache    cache.Config
    Queue    queue.QueueSchema `mapstructure:"queue"`
    // ... rest unchanged
}
```

Accessors: `cfg.Databases()` returns the connections map; `cfg.PrimaryDatabase()` returns `database.default` (defaults to `"mysql"`); `cfg.Database()` returns the primary config for backward-compat.

**`internal/infra/queue/config.go`:**

```go
type QueueSchema struct {
    Default     string                      `mapstructure:"default"`
    Connections map[string]ConnectionConfig `mapstructure:"connections"`
}

type ConnectionConfig struct {
    Enabled  bool                  `mapstructure:"enabled"`
    Driver   string                `mapstructure:"driver"` // "amqp" | "database"
    AMQP     rabbitmq.Config       `mapstructure:"amqp"`
    Database DatabaseBackendConfig `mapstructure:"database"`
}

type DatabaseBackendConfig struct {
    Connection           string        `mapstructure:"connection"`
    MaxWorkers           int           `mapstructure:"max_workers"`
    PollInterval         time.Duration `mapstructure:"poll_interval"`
    RescueStuckJobsAfter time.Duration `mapstructure:"rescue_stuck_jobs_after"`
}
```

Helper methods on `QueueSchema`: `AnyEnabled()`, `EnabledConnections() []NamedConnection`, `DefaultConnection()`, `DefaultAMQPConfig()`.

---

## 2. Database Layer

### 2.1 Connection factories

`internal/infra/database/database.go`:

**`NewMySQLClient(cfg *Config) (*bun.DB, error)`** — renamed from `NewBunClient`.

**`NewPostgresClient(cfg *Config) (*bun.DB, error)`** — uses pgx in `database/sql` stdlib mode.

```go
import (
    "github.com/uptrace/bun/dialect/pgdialect"
    _ "github.com/jackc/pgx/v5/stdlib"  // registers "pgx" as a database/sql driver
)
```

Both share a common `applyPoolSettings(db, cfg)` helper.

### 2.2 DSN helpers

```go
func GetMySQLDSN(cfg *Config) string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true",
        cfg.User, cfg.Password, cfg.Host, strconv.Itoa(cfg.Port), cfg.Name)
}

func GetPostgresDSN(cfg *Config) string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}
```

**Why `riverdatabasesql` (poll-only) instead of `riverpgxv5`?**  
bun wraps `database/sql`; pgx stdlib mode is the correct bridge. River and bun share the same `*sql.DB` — no separate connection pool. Trade-off: no LISTEN/NOTIFY (poll-only), but zero extra connections at startup.

---

## 3. Dependency Injection

### 3.1 Named providers — database

`internal/infra/providers.go` — `provideDatabases`:

```text
Named providers registered:
  "db.<name>" → *bun.DB     for each key in database.connections map
  unnamed     → *bun.DB     alias to the default connection (backward compat)
```

### 3.2 Named providers — queue

`internal/infra/providers.go` — `provideQueueInfra` (single function handles all queue connections):

```text
Named providers per enabled connection:
  "queue.conn.<name>"     → *rabbitmq.Connection   (amqp driver only)
  "queue.producer.<name>" → rabbitmq.MessageProducer (amqp driver only)
  "queue.river.<name>"    → *river.Client[*sql.Tx]  (database driver only)

Unnamed backward-compat aliases (point to default connection):
  *rabbitmq.Connection      → nil when default is not amqp
  rabbitmq.MessageProducer  → nil when default is not amqp
```

The unnamed `queue.Dispatcher` is not registered in DI — it is constructed on demand by `queue.NewDispatcher` when needed. All application code that dispatches jobs calls `queue.NewDispatcher` directly via the factory.

### 3.3 Consumer usage

```go
// All existing repos — unchanged, still get the default *bun.DB
db := do.MustInvoke[*bun.DB](i)

// Explicit named connection
db := do.MustInvokeNamed[*bun.DB](i, "db.postgres")

// Phase B: when default is switched to postgres, repos need no changes
```

---

## 4. Queue Abstraction Layer

### 4.1 Interfaces (driver-agnostic)

`internal/infra/queue/interfaces.go`:

```go
type JobArgs interface {
    Kind() string
}

type ConsumeFunc func(ctx context.Context, payload []byte) error

type Dispatcher interface {
    Dispatch(ctx context.Context, job JobArgs, opts ...DispatchOption) error
}
```

### 4.2 Dispatch options

`internal/infra/queue/options.go`:

```go
type DispatchOptions struct {
    Queue       string
    Delay       time.Duration
    MaxAttempts int
    Priority    int
}

func ApplyOptions(opts ...DispatchOption) *DispatchOptions {
    return &DispatchOptions{Queue: "default", MaxAttempts: 3, Priority: 1}
    // then applies each opt
}

func OnQueue(name string) DispatchOption { ... }
func Delay(d time.Duration) DispatchOption { ... }
func MaxAttempts(n int) DispatchOption { ... }
func Priority(p int) DispatchOption { ... }
```

> **Note:** `UniqueKey` was removed from the final implementation. River unique jobs are handled by passing `river.UniqueOpts` directly to typed job args when needed.

**AMQP validation:** The AMQP dispatcher **rejects** `Queue`/`MaxAttempts`/`Priority` options at dispatch time with a clear error message. Only `Delay` is supported. This surfaces driver/option mismatches immediately rather than silently ignoring them.

### 4.3 Directory layout

```text
internal/infra/queue/
├── interfaces.go        # Dispatcher, JobArgs, ConsumeFunc
├── options.go           # DispatchOption helpers (no UniqueKey)
├── config.go            # QueueSchema, ConnectionConfig, DatabaseBackendConfig
├── registry.go          # GetRegisteredConsumers — builds ConsumerRegistration list
├── dispatcher.go        # NewDispatcher factory + inline rabbitMQDispatcher + riverDispatcher
├── rabbitmq/            # AMQP connection, producer, consumer, topology
│   └── ...              # unchanged internals; ConsumeFunc is a type alias for queue.ConsumeFunc
└── river/
    ├── worker.go        # BridgeWorker (single worker, dispatches by ConsumerName)
    └── worker_pool.go   # RegisterBridgeWorkers → builds one BridgeWorker for all consumers
```

---

## 5. River Implementation

### 5.1 BridgeWorker (single instance, dispatch-by-name)

River panics if you add two workers with the same `Kind()`. Since all bridge jobs share `Kind() == "generic_job"`, exactly **one** `BridgeWorker` is registered. It holds a map of `consumerName → ConsumeFunc` and dispatches at runtime.

```go
// internal/infra/queue/river/worker.go

type GenericJobArgs struct {
    ConsumerName string `json:"consumer_name"`
    Payload      []byte `json:"payload"`
}

func (GenericJobArgs) Kind() string { return "generic_job" }

type BridgeWorker struct {
    river.WorkerDefaults[GenericJobArgs]
    handlers map[string]queue.ConsumeFunc
}

func (w *BridgeWorker) Work(ctx context.Context, job *river.Job[GenericJobArgs]) error {
    handler, ok := w.handlers[job.Args.ConsumerName]
    if !ok {
        return fmt.Errorf("bridge worker: no handler registered for consumer %q", job.Args.ConsumerName)
    }
    return handler(ctx, job.Args.Payload)
}
```

```go
// internal/infra/queue/river/worker_pool.go

// RegisterBridgeWorkers builds a single BridgeWorker from all ConsumerRegistrations.
// Returns an error on duplicate consumer names (caught at startup).
func RegisterBridgeWorkers(workers *river.Workers, registrations []queue.ConsumerRegistration) error {
    handlers := make(map[string]queue.ConsumeFunc, len(registrations))
    for _, reg := range registrations {
        if _, exists := handlers[reg.Name]; exists {
            return fmt.Errorf("river: duplicate consumer registration for name %q", reg.Name)
        }
        handlers[reg.Name] = reg.ConsumeFunc
    }
    river.AddWorker(workers, NewBridgeWorker(handlers))
    return nil
}
```

### 5.2 River client (shared bun *sql.DB, poll-only)

`buildRiverClient` in `providers.go` — constructs the River client using `riverdatabasesql.New(bunDB.DB)`:

```go
func buildRiverClient(bunDB *bun.DB, cfg queue.DatabaseBackendConfig, registrations []queue.ConsumerRegistration) (*river.Client[*sql.Tx], error) {
    workers := river.NewWorkers()
    if err := riverimpl.RegisterBridgeWorkers(workers, registrations); err != nil {
        return nil, fmt.Errorf("river: %w", err)
    }
    return river.NewClient(riverdatabasesql.New(bunDB.DB), &river.Config{
        Queues: map[string]river.QueueConfig{
            river.QueueDefault: {MaxWorkers: maxWorkers},
            "emails":           {MaxWorkers: 10},
            "notifications":    {MaxWorkers: 20},
        },
        Workers:              workers,
        FetchPollInterval:    pollInterval,
        RescueStuckJobsAfter: rescueAfter,
    })
}
```

Named as `"queue.river.<connName>"` in DI.

### 5.3 River dispatcher (inline in queue/dispatcher.go)

Both dispatchers (`rabbitMQDispatcher` and `riverDispatcher`) are unexported structs defined inline in `internal/infra/queue/dispatcher.go`. There is no separate `rabbitmq/dispatcher.go`.

### 5.4 Failed jobs

River stores all job lifecycle states in `river_job`:
- `available` → `running` → `completed`
- `running` → `retryable` → `available` (after backoff)
- `running` → `failed` (exhausted attempts)

### 5.5 Migrations

River manages its own schema via `rivermigrate`. Called on application startup before serving traffic. Tables: `river_job`, `river_queue`, `river_leader`, `river_migration`.

Goose manages application domain schema migrations (unchanged).

---

## 6. Driver Factory

`internal/infra/queue/dispatcher.go` — `NewDispatcher`:

```go
func NewDispatcher(driver string, producer rabbitmq.MessageProducer, riverClient *river.Client[*sql.Tx]) (Dispatcher, error) {
    switch driver {
    case "amqp":
        // Validates producer != nil; returns rabbitMQDispatcher
    case "database":
        // Validates riverClient != nil; returns riverDispatcher
    default:
        return nil, fmt.Errorf("unknown queue driver: %q (valid: amqp, database)", driver)
    }
}
```

Driver names: `"amqp"` (RabbitMQ) and `"database"` (River). These match the `driver:` field in the queue `connections` map.

---

## 7. Server Startup

`cmd/server/queue_server.go` — `StartQueueWorkers`:

```go
func StartQueueWorkers(ctx context.Context, queueCfg *queue.QueueSchema, injector do.Injector)
```

Iterates `queueCfg.EnabledConnections()` and launches each driver's worker set in a separate goroutine — AMQP and River can run **concurrently** in the same process. `cmd/main.go` calls `StartQueueWorkers` when `cfg.Queue().AnyEnabled()`.

---

## 8. New Go Dependencies

| Package | Purpose |
|---|---|
| `github.com/uptrace/bun/dialect/pgdialect` | bun PostgreSQL dialect |
| `github.com/jackc/pgx/v5/stdlib` | pgx as database/sql driver (for bun) |
| `github.com/riverqueue/river` | River job queue core |
| `github.com/riverqueue/river/riverdriver/riverdatabasesql` | River driver sharing bun's `*sql.DB` (poll-only) |
| `github.com/riverqueue/river/rivermigrate` | River schema migrations |

---

## 9. Phase B: Migrate Domains to PostgreSQL

**Goal:** Promote PostgreSQL to the primary application database. Switch `database.default` from `"mysql"` to `"postgres"`, create Postgres-compatible migrations for all domain tables, and verify the full stack runs against Postgres.

**Preconditions (all satisfied by Phase A):**
- `database.connections.postgres` exists in config
- `NewPostgresClient` + `GetPostgresDSN` are implemented
- Named DI providers `"db.postgres"` and unnamed alias work
- `db/cmd/migrate.go` auto-detects driver from `database.default`

**Steps:**

1. **Set `database.default: "postgres"`** in `config.example.yaml` and `config.local.yaml`. The unnamed `*bun.DB` in DI now points to Postgres — zero changes to existing repos.

2. **Dialect-specific migration subdirectories.** The existing migrations in `db/migrations/schema/` use MySQL syntax. Move them to `db/migrations/schema/mysql/` and create `db/migrations/schema/postgres/` for Postgres-compatible DDL. Update `migrate.go` to auto-select the subdirectory based on `database.default` (overridable via `--dir`). Key Postgres differences:
   - `AUTO_INCREMENT` → `BIGSERIAL`
   - `TINYINT(1)` → `BOOLEAN`
   - Backtick identifiers → unquoted (or double-quoted)
   - `DATETIME`/`TIMESTAMP ON UPDATE` → `TIMESTAMPTZ`
   - `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4` → remove entirely
   - `JSON` → `JSONB` (preferred)

3. **Run migrations** against the Postgres connection: `make migration-up`

4. **Update seed files** if any use MySQL-specific syntax.

5. **Integration test** — run the full test suite against Postgres: `go test ./... -count=1`

6. **Decommission MySQL** — remove `database.connections.mysql` from config once all data is migrated and verified. Keep it in `config.example.yaml` as a commented-out example.

No changes to repository interfaces, service layer, or HTTP handlers are required.

---

## 10. Out of Scope

- Removing RabbitMQ — kept as selectable driver
- Migrating existing production data from MySQL to PostgreSQL (ops concern)
- River Web UI setup (separate operational concern)
- Multi-queue worker configuration tuning (handled at deployment time)
