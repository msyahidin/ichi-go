# PostgreSQL Support + River Queue Design

**Date:** 2026-05-01
**Branch:** feat/newqueue
**Status:** Approved for implementation

---

## Overview

Two sequential phases:

- **Phase C (now):** Add a PostgreSQL connection alongside the existing MySQL connection, used exclusively by [riverqueue/river](https://github.com/riverqueue/river) as a RabbitMQ-optional queue backend.
- **Phase B (later):** When queue and config are stable, migrate application domains from MySQL to PostgreSQL as the primary database.

RabbitMQ is kept as a selectable driver — no removal.

---

## 1. Config Structure

### 1.1 YAML shape

Replace the single `database:` block with a named `databases:` map. Any string key is a valid connection name. Multiple connections of the same driver (e.g. `mysql` and `mysql_legacy`) are supported.

```yaml
databases:
  mysql:                        # default app DB — key used in DI as "db.mysql"
    driver: mysql
    host: localhost
    port: 3306
    name: ichi_app
    user: root
    password: password
    max_idle_conns: 10
    max_open_conns: 100
    max_conn_life_time: 3600
    debug: true

  postgres:                     # queue DB — key used in DI as "db.postgres"
    driver: postgres
    host: localhost
    port: 5432
    name: ichi_queue
    user: postgres
    password: postgres
    max_idle_conns: 5
    max_open_conns: 20
    max_conn_life_time: 3600
    debug: false

  # Example of a second MySQL connection:
  # mysql_legacy:
  #   driver: mysql
  #   host: legacy-db.internal
  #   name: legacy_app
  #   ...

queue:
  enabled: true
  driver: "rabbitmq"            # "rabbitmq" | "river"
  rabbitmq:
    # ... existing config, unchanged
  river:
    database: "postgres"        # key from databases map to use for River
    max_workers: 50             # total worker slots across all queues
    poll_interval: "1s"         # fallback polling interval (LISTEN/NOTIFY is primary)
    rescue_stuck_jobs_after: "1h"
```

### 1.2 Go config structs

**`internal/infra/database/config.go`** — `Config` struct is unchanged. `SetDefault` is updated for the new key path.

**`config/config.go`:**

```go
type Schema struct {
    App        AppConfig
    Databases  map[string]database.Config  `mapstructure:"databases"` // renamed from Database
    Cache      cache.Config
    Queue      queue.Config
    // ... rest unchanged
}
```

A `databases.primary` field (optional, defaults to `"mysql"`) selects which named connection is exposed as the unnamed `*bun.DB` for backward compatibility:

```yaml
databases:
  primary: "mysql"   # optional — which key is the default *bun.DB
  mysql:
    ...
```

---

## 2. Database Layer

### 2.1 Connection factories

`internal/infra/database/database.go` is split into two functions:

**`NewMySQLClient(cfg *Config) (*bun.DB, error)`**
Existing logic, renamed from `NewBunClient`. No behaviour change.

**`NewPostgresClient(cfg *Config) (*bun.DB, error)`**
New function. Uses pgx in `database/sql` stdlib mode so bun wraps it identically to MySQL.

```go
import (
    "github.com/uptrace/bun/dialect/pgdialect"
    _ "github.com/jackc/pgx/v5/stdlib"  // registers "pgx" as a database/sql driver
)

func NewPostgresClient(cfg *Config) (*bun.DB, error) {
    sqldb, err := sql.Open("pgx", GetPostgresDSN(cfg))
    if err != nil {
        return nil, fmt.Errorf("failed to open postgres connection: %w", err)
    }
    if err := sqldb.Ping(); err != nil {
        sqldb.Close()
        return nil, fmt.Errorf("failed to ping postgres: %w", err)
    }
    db := bun.NewDB(sqldb, pgdialect.New())
    db.SetMaxIdleConns(cfg.MaxIdleConns)
    db.SetMaxOpenConns(cfg.MaxOpenConns)
    db.SetConnMaxLifetime(time.Duration(cfg.MaxConnLifeTime) * time.Second)
    if cfg.Debug {
        db.WithQueryHook(&hook.DebugHook{})
    }
    return db, nil
}

func GetPostgresDSN(cfg *Config) string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}
```

**Why pgx stdlib mode for bun, but native pgxpool for River?**

- bun wraps `database/sql` — pgx stdlib mode is the correct bridge.
- River needs pgx native pool (`pgxpool.Pool`) for its LISTEN/NOTIFY and advisory lock features. The two connection objects share the same PostgreSQL server but are independent pools. River's pool is private — never registered in DI.

### 2.2 DSN helpers

```go
// internal/infra/database/database.go
func GetMySQLDSN(cfg *Config) string {
    return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
}
```

---

## 3. Dependency Injection

### 3.1 Named providers

`internal/infra/providers.go` — `provideDatabase` is replaced by `provideDatabases`.

```
Named providers registered:
  "db.<name>" → *bun.DB     for each key in databases map
  unnamed     → *bun.DB     alias to the primary connection (backward compat)
```

```go
func provideDatabases(cfg *config.Config) func(do.Injector) error {
    return func(i do.Injector) error {
        primary := cfg.PrimaryDatabase() // reads databases.primary string, defaults to "mysql"
        if primary == "" {
            primary = "mysql"
        }

        for name, dbCfg := range cfg.Databases() {
            if name == "primary" { continue }          // skip the meta key
            name, dbCfg := name, dbCfg

            switch dbCfg.Driver {
            case "mysql":
                do.ProvideNamed(i, "db."+name, func(i do.Injector) (*bun.DB, error) {
                    return database.NewMySQLClient(&dbCfg)
                })
            case "postgres":
                do.ProvideNamed(i, "db."+name, func(i do.Injector) (*bun.DB, error) {
                    return database.NewPostgresClient(&dbCfg)
                })
            }
        }

        // Unnamed *bun.DB — points to primary, keeps all existing repos working with zero changes
        do.Provide(i, func(i do.Injector) (*bun.DB, error) {
            return do.InvokeNamed[*bun.DB](i, "db."+primary)
        })

        return nil
    }
}
```

### 3.2 Consumer usage

```go
// Existing repo — unchanged, still gets MySQL
db := do.MustInvoke[*bun.DB](i)

// Repo explicitly on a named connection
db := do.MustInvokeNamed[*bun.DB](i, "db.mysql_legacy")

// Phase B: postgres domain repo
db := do.MustInvokeNamed[*bun.DB](i, "db.postgres")
```

---

## 4. Queue Abstraction Layer

### 4.1 Interfaces (driver-agnostic)

Moved up from `internal/infra/queue/rabbitmq/` to `internal/infra/queue/`:

```go
// internal/infra/queue/interfaces.go

type JobArgs interface {
    Kind() string  // unique job type identifier, e.g. "notification.send_email"
}

type Dispatcher interface {
    Dispatch(ctx context.Context, job JobArgs, opts ...DispatchOption) error
}

type ConsumeFunc func(ctx context.Context, payload []byte) error
```

### 4.2 Dispatch options

Laravel equivalents: `->onQueue()`, `->delay()`, `->tries()`.

```go
// internal/infra/queue/options.go

type DispatchOptions struct {
    Queue       string
    Delay       time.Duration
    MaxAttempts int
    Priority    int
    UniqueKey   string
}

type DispatchOption func(*DispatchOptions)

func OnQueue(name string) DispatchOption  { return func(o *DispatchOptions) { o.Queue = name } }
func Delay(d time.Duration) DispatchOption { return func(o *DispatchOptions) { o.Delay = d } }
func MaxAttempts(n int) DispatchOption    { return func(o *DispatchOptions) { o.MaxAttempts = n } }
func Priority(p int) DispatchOption       { return func(o *DispatchOptions) { o.Priority = p } }
func UniqueKey(k string) DispatchOption   { return func(o *DispatchOptions) { o.UniqueKey = k } }
```

**Usage (Laravel parity):**

```go
// Laravel: SendEmailJob::dispatch($data)->onQueue('emails')->delay(now()->addMinutes(5))
queue.Dispatch(ctx, SendEmailJob{To: "user@example.com"},
    queue.OnQueue("emails"),
    queue.Delay(5*time.Minute),
    queue.MaxAttempts(3),
)
```

### 4.3 Config additions

```go
// internal/infra/queue/config.go
type Config struct {
    Enabled  bool            `mapstructure:"enabled"`
    Driver   string          `mapstructure:"driver"`   // "rabbitmq" | "river"
    RabbitMQ rabbitmq.Config `mapstructure:"rabbitmq"`
    River    RiverConfig     `mapstructure:"river"`
}

type RiverConfig struct {
    Database              string        `mapstructure:"database"`                // key from databases map
    MaxWorkers            int           `mapstructure:"max_workers"`
    PollInterval          time.Duration `mapstructure:"poll_interval"`
    RescueStuckJobsAfter  time.Duration `mapstructure:"rescue_stuck_jobs_after"`
}
```

### 4.4 Directory layout

```
internal/infra/queue/
├── interfaces.go        # Dispatcher, JobArgs, ConsumeFunc
├── options.go           # DispatchOption helpers
├── config.go            # Config + RiverConfig
├── registry.go          # ConsumerRegistration list (unchanged structure)
├── dispatcher.go        # factory: builds RabbitMQ or River Dispatcher per driver config
├── rabbitmq/            # existing implementation, now satisfies Dispatcher
│   └── ...              # unchanged internals
└── river/
    ├── config.go        # river-specific defaults
    ├── dispatcher.go    # river.Client Dispatch impl
    ├── worker.go        # BridgeWorker: wraps ConsumeFunc into river.Worker
    ├── worker_pool.go   # registers all typed workers with river.Workers
    └── migrations/      # river_job, river_queue, river_leader SQL (goose format)
```

---

## 5. River Implementation

### 5.1 Typed job (Laravel Job class equivalent)

```go
// internal/applications/notification/jobs/send_email_job.go
type SendEmailJob struct {
    To      string `json:"to"`
    Subject string `json:"subject"`
    Body    string `json:"body"`
}

func (SendEmailJob) Kind() string { return "notification.send_email" }
```

### 5.2 Typed worker (Laravel `handle()` equivalent)

```go
// internal/applications/notification/workers/send_email_worker.go
type SendEmailWorker struct {
    river.WorkerDefaults[jobs.SendEmailJob]
    mailer Mailer
}

func (w *SendEmailWorker) Work(ctx context.Context, job *river.Job[jobs.SendEmailJob]) error {
    return w.mailer.Send(ctx, job.Args.To, job.Args.Subject, job.Args.Body)
}
```

### 5.3 Bridge worker (backward compat for existing ConsumeFunc consumers)

Existing consumers (`PaymentConsumer`, `WelcomeNotificationConsumer`, etc.) use `func(ctx, []byte) error`. The bridge wraps them into a River-compatible worker with zero changes to the consumer logic.

```go
// internal/infra/queue/river/worker.go

// BridgeArgs carries a raw payload for legacy ConsumeFunc consumers.
// Kind is set per-registration so each consumer maps to its own River job kind.
type BridgeArgs struct {
    Kind_   string `json:"kind"`    // populated by the bridge dispatcher, matches the consumer name
    Payload []byte `json:"payload"`
}

func (a BridgeArgs) Kind() string { return a.Kind_ }

type BridgeWorker struct {
    river.WorkerDefaults[BridgeArgs]
    handler queue.ConsumeFunc
}

func (w *BridgeWorker) Work(ctx context.Context, job *river.Job[BridgeArgs]) error {
    return w.handler(ctx, job.Args.Payload)
}
```

### 5.4 River provider (pgxpool is private)

```go
// internal/infra/providers.go
func provideRiverClient(cfg *config.Config) func(do.Injector) (*river.Client[pgx.Tx], error) {
    return func(i do.Injector) (*river.Client[pgx.Tx], error) {
        riverCfg := cfg.Queue().River
        dbCfg := cfg.Databases()[riverCfg.Database]  // e.g. databases["postgres"]

        // pgxpool is River's private connection — not registered in DI
        pool, err := pgxpool.New(context.Background(), database.GetPostgresDSN(&dbCfg))
        if err != nil {
            return nil, fmt.Errorf("river: failed to create pgxpool: %w", err)
        }

        workers := river.NewWorkers()
        // typed workers registered here (see worker_pool.go)
        // bridge workers for existing ConsumeFunc consumers registered here

        return river.NewClient(riverpgxv5.New(pool), &river.Config{
            Queues: map[string]river.QueueConfig{
                river.QueueDefault: {MaxWorkers: riverCfg.MaxWorkers},
                "emails":           {MaxWorkers: 10},
                "notifications":    {MaxWorkers: 20},
            },
            Workers:              workers,
            PollInterval:         riverCfg.PollInterval,
            RescueStuckJobsAfter: riverCfg.RescueStuckJobsAfter,
        })
    }
}
```

### 5.5 River dispatcher (implements queue.Dispatcher)

```go
// internal/infra/queue/river/dispatcher.go
func (d *RiverDispatcher) Dispatch(ctx context.Context, job queue.JobArgs, opts ...queue.DispatchOption) error {
    o := &queue.DispatchOptions{Queue: river.QueueDefault, MaxAttempts: 3}
    for _, opt := range opts { opt(o) }

    insertOpts := &river.InsertOpts{
        Queue:       o.Queue,
        MaxAttempts: o.MaxAttempts,
        Priority:    o.Priority,
    }
    if o.Delay > 0 {
        insertOpts.ScheduledAt = time.Now().Add(o.Delay)
    }
    if o.UniqueKey != "" {
        insertOpts.UniqueOpts = river.UniqueOpts{ByArgs: true}
    }

    _, err := d.client.Insert(ctx, job, insertOpts)
    return err
}
```

### 5.6 Failed jobs

River stores all job lifecycle states in the `river_job` table:
- `available` → `running` → `completed`
- `running` → `retryable` → (retry after backoff) → `available`
- `running` → `failed` (exhausted attempts)
- `running` → `cancelled`

This is equivalent to Laravel's `jobs` + `failed_jobs` tables combined. The River Web UI provides inspection and manual retry without writing any admin code.

### 5.7 Migrations

River manages its own schema via `rivermigrate`. Called on application startup before serving traffic:

```go
// cmd/server/rest_server.go (or a dedicated migration step)
migrator := rivermigrate.New(riverpgxv5.New(pool), nil)
_, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
```

Tables created: `river_job`, `river_queue`, `river_leader`, `river_migration`.

Goose manages application schema migrations (unchanged). River manages its own tables independently.

---

## 6. Driver Factory

```go
// internal/infra/queue/dispatcher.go
func NewDispatcher(cfg *Config, injector do.Injector) (Dispatcher, error) {
    switch cfg.Driver {
    case "river":
        client := do.MustInvoke[*river.Client[pgx.Tx]](injector)
        return river.NewDispatcher(client), nil
    case "rabbitmq":
        producer := do.MustInvoke[rabbitmq.MessageProducer](injector)
        return rabbitmq.NewDispatcher(producer), nil
    default:
        return nil, fmt.Errorf("unknown queue driver: %s", cfg.Driver)
    }
}
```

The `Dispatcher` is registered in DI, and all application code that publishes jobs calls `do.MustInvoke[queue.Dispatcher](i)` — unaware of which backend is active.

---

## 7. New Go Dependencies

| Package | Purpose |
|---|---|
| `github.com/uptrace/bun/dialect/pgdialect` | bun PostgreSQL dialect |
| `github.com/jackc/pgx/v5/stdlib` | pgx as database/sql driver (for bun) |
| `github.com/jackc/pgx/v5/pgxpool` | pgx native pool (for River, internal) |
| `github.com/riverqueue/river` | River job queue core |
| `github.com/riverqueue/river/riverdriver/riverpgxv5` | River pgx driver |
| `github.com/riverqueue/river/rivermigrate` | River schema migrations |

---

## 8. Phase B Notes (future)

When queue and config are stable, Phase B promotes PostgreSQL to primary:

1. Rename `databases.postgres` to whatever name makes sense (e.g. keep as `postgres`)
2. Update `databases.primary: "postgres"` — the unnamed `*bun.DB` now points to Postgres
3. Run domain migrations against Postgres (new goose migration files)
4. Switch each domain's `register.go` from `do.MustInvoke[*bun.DB]` to `do.MustInvokeNamed[*bun.DB](i, "db.postgres")` — or just rely on the updated unnamed alias
5. MySQL connection stays in config for read-only legacy access until fully decommissioned

No changes to repository interfaces, service layer, or HTTP handlers required for Phase B.

---

## 9. Out of Scope

- Removing RabbitMQ — kept as selectable driver
- Migrating existing application data from MySQL to PostgreSQL (Phase B)
- River Web UI setup (separate operational concern)
- Multi-queue worker configuration tuning (handled at deployment time)
