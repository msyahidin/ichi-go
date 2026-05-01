# PostgreSQL + River Queue Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add named multi-driver database config (MySQL + PostgreSQL via bun), then wire riverqueue/river as a selectable queue backend alongside the existing RabbitMQ/AMQP driver.

**Architecture:** Both `database` and `queue` use the same Laravel-style **named connections map** (`{default, connections: map}`). The `database` map produces named `*bun.DB` providers; the `queue` map produces named `*rabbitmq.Connection`, `rabbitmq.MessageProducer`, and `*river.Client[*sql.Tx]` providers. A driver-agnostic `queue.Dispatcher` factory (`NewDispatcher`) routes jobs to either RabbitMQ or River based on the chosen driver. Existing consumers are bridged to River via a single `BridgeWorker` that dispatches by `ConsumerName` at runtime.

**Tech Stack:** Go 1.25, uptrace/bun + pgdialect, jackc/pgx/v5/stdlib, riverqueue/river, riverdatabasesql (poll-only, shares bun's `*sql.DB`), rivermigrate, samber/do/v2

---

## File Map

### Phase A (complete)

**Created:**
- `internal/infra/queue/interfaces.go` — `Dispatcher`, `JobArgs`, `ConsumeFunc`
- `internal/infra/queue/options.go` — `DispatchOption`, `ApplyOptions`, `OnQueue`, `Delay`, `MaxAttempts`, `Priority`
- `internal/infra/queue/dispatcher.go` — `NewDispatcher` factory + inline `rabbitMQDispatcher` + `riverDispatcher`
- `internal/infra/queue/river/worker.go` — `GenericJobArgs`, `BridgeWorker` (single worker, dispatch-by-name)
- `internal/infra/queue/river/worker_pool.go` — `RegisterBridgeWorkers` (returns error on duplicate names)

**Modified:**
- `go.mod` / `go.sum` — pgx, pgdialect, river, riverdatabasesql, rivermigrate
- `config.example.yaml` — `database.connections` map + `database.default` + `queue.connections` map
- `config/config.go` — `DatabaseSchema` struct, `Databases()`, `PrimaryDatabase()`, `Database()` accessors; `Queue queue.QueueSchema`
- `internal/infra/database/config.go` — `SetDefault` updated to `databases.mysql.*` key paths
- `internal/infra/database/database.go` — `NewBunClient`→`NewMySQLClient`; add `NewPostgresClient`, `GetMySQLDSN`, `GetPostgresDSN`; shared `applyPoolSettings`
- `internal/infra/queue/config.go` — `QueueSchema`, `ConnectionConfig`, `DatabaseBackendConfig`; `SetDefault` updated
- `internal/infra/queue/rabbitmq/consumer_interface.go` — `ConsumeFunc` is now a type alias for `queue.ConsumeFunc`
- `internal/infra/queue/registry.go` — `ConsumerRegistration.ConsumeFunc` uses `queue.ConsumeFunc` directly
- `internal/infra/providers.go` — `provideDatabases` (named `*bun.DB`); `provideQueueInfra` (replaces `provideMessaging` + `provideMessageProducer`); `buildRiverClient` helper
- `cmd/main.go` — queue startup uses `cfg.Queue().AnyEnabled()` + `server.StartQueueWorkers(ctx, cfg.Queue(), injector)`
- `cmd/server/queue_server.go` — `StartQueueWorkers(*queue.QueueSchema, ...)` runs all enabled connections concurrently
- `db/cmd/migrate.go` — auto-detects driver; `GetMySQLDSN` / `GetPostgresDSN` based on `database.default`

---

## Phase A: Tasks (all complete)

### Task 1: Add Go dependencies ✅

- [x] Install pgx, pgdialect, river, riverdatabasesql, rivermigrate
- [x] `go build ./...` passes

### Task 2: Rename database client functions + add PostgreSQL client ✅

- [x] `NewBunClient` → `NewMySQLClient`
- [x] `GetDsn` → `GetMySQLDSN`
- [x] Added `NewPostgresClient`, `GetPostgresDSN`
- [x] Shared `applyPoolSettings` helper
- [x] Tests: `TestGetMySQLDSN`, `TestGetPostgresDSN`

### Task 3: Update config schema for named databases map ✅

- [x] `Schema.Database DatabaseSchema` (has `Default` + `Connections map`)
- [x] `Databases()`, `PrimaryDatabase()`, `Database()` accessors
- [x] `SetDefault` updated to `databases.mysql.*` key paths
- [x] `config.example.yaml` updated

### Task 4: Update DI providers for named *bun.DB connections ✅

- [x] `provideDatabases` registers `"db.<name>"` per driver
- [x] Unnamed `*bun.DB` alias → `"db." + primary`
- [x] `db/cmd/migrate.go` updated to auto-detect driver

### Task 5: Create driver-agnostic queue interfaces ✅

- [x] `interfaces.go` — `Dispatcher`, `JobArgs`, `ConsumeFunc`
- [x] `options.go` — `DispatchOptions`, `ApplyOptions`, option funcs (no `UniqueKey`)
- [x] `rabbitmq/consumer_interface.go` — `ConsumeFunc = queue.ConsumeFunc` alias

### Task 6: Queue config schema ✅

- [x] `QueueSchema` with `Default` + `Connections map[string]ConnectionConfig`
- [x] `ConnectionConfig` with `Driver` (`"amqp"` | `"database"`), `AMQP`, `Database`
- [x] `DatabaseBackendConfig` with `Connection`, `MaxWorkers`, `PollInterval`, `RescueStuckJobsAfter`
- [x] Helper methods: `AnyEnabled()`, `EnabledConnections()`, `DefaultConnection()`, `DefaultAMQPConfig()`
- [x] `SetDefault` updated

### Task 7: Inline RabbitMQ dispatcher ✅

Both dispatchers are unexported structs in `internal/infra/queue/dispatcher.go`:
- [x] `rabbitMQDispatcher` — serializes to JSON, publishes via `MessageProducer`; **rejects** `Queue`/`MaxAttempts`/`Priority` options fast-fail (only `Delay` supported)
- [x] `riverDispatcher` — builds `river.InsertOpts` from `DispatchOptions`

### Task 8: BridgeWorker (single instance, dispatch-by-ConsumerName) ✅

- [x] `GenericJobArgs{ConsumerName, Payload}`, `Kind() == "generic_job"`
- [x] `BridgeWorker` holds `handlers map[string]queue.ConsumeFunc`
- [x] `Work()` dispatches by `job.Args.ConsumerName`; returns error for unknown names

### Task 9: River dispatcher ✅

- [x] `riverDispatcher.Dispatch` builds `InsertOpts` from `DispatchOptions`
- [x] `ScheduledAt` set when `Delay > 0`

### Task 10: Worker pool registration ✅

- [x] `RegisterBridgeWorkers` returns `error` on duplicate consumer names
- [x] Adds exactly **one** `BridgeWorker` (avoids duplicate-kind panic)

### Task 11: Queue dispatcher factory ✅

- [x] `NewDispatcher(driver, producer, riverClient)` — drivers: `"amqp"`, `"database"`
- [x] Returns descriptive error for unknown driver

### Task 12: Wire queue infra into DI ✅

- [x] `provideQueueInfra` iterates `EnabledConnections()`
- [x] Named providers: `queue.conn.<name>`, `queue.producer.<name>`, `queue.river.<name>`
- [x] Unnamed backward-compat aliases for `*rabbitmq.Connection` and `rabbitmq.MessageProducer`
- [x] `buildRiverClient` helper shared by all database-driver connections

### Task 13: Multi-driver queue_server.go ✅

- [x] `StartQueueWorkers(ctx, *queue.QueueSchema, injector)` — concurrent per-driver goroutines
- [x] `startAMQPWorkers` — topology setup with backoff, per-consumer goroutines
- [x] `startRiverWorkers` — `client.Start` / `client.Stop` with 30 s timeout
- [x] `cmd/main.go` — `cfg.Queue().AnyEnabled()` guard

### Task 14: ConsumerRegistration cleanup ✅

- [x] `ConsumerRegistration.ConsumeFunc` is `queue.ConsumeFunc` directly (no rabbitmq import)

### Task 15: Smoke test ✅

- [x] `go test ./... -short` passes
- [x] `go build ./...` produces binary
- [x] Server starts with queue disabled

---

## Phase B: Migrate Domains to PostgreSQL

**Goal:** Switch `database.default` from `"mysql"` to `"postgres"` and migrate all domain table schemas to PostgreSQL-compatible migrations. No code changes required to repositories, services, or HTTP handlers — they all use the unnamed `*bun.DB` which will automatically point to Postgres.

**File Map:**

**Modify:**
- `config.example.yaml` — `database.default: "postgres"` (switch default)
- `config.local.yaml` — same
- `db/migrations/schema/` — add new `.sql` migration files with Postgres-compatible DDL

**Create:**
- `db/migrations/schema/YYYYMMDDHHMMSS_postgres_domain_tables.sql` — one or more migration files porting domain tables to Postgres syntax

---

### Task B1: Switch database.default to postgres

**Files:** `config.example.yaml`, `config.local.yaml`

- [ ] **Step 1: Update `config.example.yaml`**

Change:
```yaml
database:
  default: "mysql"
```
To:
```yaml
database:
  default: "postgres"
```

- [ ] **Step 2: Update `config.local.yaml`**

Same change. Verify Postgres connection settings point to a running instance.

- [ ] **Step 3: Verify build still passes**

```bash
go build ./...
```

Expected: no errors.

- [ ] **Step 4: Verify config loads correctly**

```bash
go test ./config/... -v -run TestDatabasesMapUnmarshal
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add config.example.yaml config.local.yaml
git commit -m "feat(config): switch database.default to postgres"
```

---

### Task B2: Add dialect-specific migration directory support to migrate.go

**Problem:** The existing migrations in `db/migrations/schema/` use MySQL-specific syntax (`AUTO_INCREMENT`, backtick identifiers, `ENGINE=InnoDB`, etc.). Running them against Postgres will fail. Since this is a template (no live MySQL data to preserve), the clean solution is a **per-dialect subdirectory**: `db/migrations/schema/mysql/` and `db/migrations/schema/postgres/`. `migrate.go` auto-selects the subdirectory based on `database.default`.

**Files:** `db/cmd/migrate.go`, `db/migrations/schema/`

- [ ] **Step 1: Create dialect subdirectories**

```bash
mkdir -p db/migrations/schema/mysql
mkdir -p db/migrations/schema/postgres
```

Move existing MySQL migration files into the `mysql/` subdirectory:

```bash
mv db/migrations/schema/*.sql db/migrations/schema/mysql/
```

- [ ] **Step 2: Update `connectDatabase` in `migrate.go` to set `dir` based on dialect**

In `connectDatabase`, after resolving `driverName`/`dialect`, update the default migration dir:

```go
func connectDatabase(environment string) (*sql.DB, string) {
    config.MustLoad()
    dbConfig := config.Get().Database()

    var driverName, dsn, dialect string
    switch dbConfig.Driver {
    case "postgres":
        driverName = "pgx"
        dsn = database.GetPostgresDSN(dbConfig)
        dialect = "postgres"
    default:
        driverName = "mysql"
        dsn = database.GetMySQLDSN(dbConfig)
        dialect = "mysql"
    }

    // ... open + ping ...

    // Return the auto-detected dialect subdir so callers can default to it.
    return db, dialect
}
```

Update callers to use the returned dialect for the default `--dir`:

```go
db, dialect := connectDatabase(*env)
if *dir == defaultMigrationDir {
    *dir = filepath.Join(defaultMigrationDir, dialect)
}
```

This keeps `--dir` overridable (for CI/CD or custom paths) while defaulting to the dialect subdirectory.

- [ ] **Step 3: Build and verify**

```bash
go build ./db/cmd/...
```

Expected: success.

- [ ] **Step 4: Commit**

```bash
git add db/cmd/migrate.go db/migrations/schema/
git commit -m "feat(migrate): dialect-specific subdirectory (mysql/ postgres/) auto-selected by driver"
```

---

### Task B3: Create Postgres-compatible schema migrations

**Files:** `db/migrations/schema/postgres/`

Common MySQL → Postgres conversions:

| MySQL | PostgreSQL |
|---|---|
| `INT AUTO_INCREMENT PRIMARY KEY` | `BIGSERIAL PRIMARY KEY` |
| `BIGINT AUTO_INCREMENT PRIMARY KEY` | `BIGSERIAL PRIMARY KEY` |
| `TINYINT(1)` | `BOOLEAN` |
| `DATETIME` / `TIMESTAMP ON UPDATE` | `TIMESTAMPTZ` (no `ON UPDATE` — handle in app or trigger) |
| `` `column_name` `` | `column_name` (no backticks) |
| `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4` | *(remove entirely)* |
| `JSON` | `JSONB` (preferred in Postgres) |
| `COMMENT '...'` | *(remove — Postgres uses `COMMENT ON COLUMN` separately)* |
| `UNIQUE INDEX name (cols)` | `CREATE UNIQUE INDEX name ON table(cols)` |

- [ ] **Step 1: Create Postgres migration files**

Create one file per domain (or one large baseline file):

```bash
# Option A: one file per domain (easier to review)
touch db/migrations/schema/postgres/20260502000001_create_users_table.sql
touch db/migrations/schema/postgres/20260502000002_create_order_tables.sql
touch db/migrations/schema/postgres/20260502000003_create_rbac_tables.sql
touch db/migrations/schema/postgres/20260502000004_create_notification_tables.sql
```

- [ ] **Step 2: Write Postgres DDL**

Example for users:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(100) NOT NULL UNIQUE,
    password    VARCHAR(255) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;
```

Mirror the same tables from `db/migrations/schema/mysql/` for each domain.

- [ ] **Step 3: Run migrations against Postgres**

```bash
make migration-status   # should show 0 applied (fresh postgres DB)
make migration-up
make migration-status   # should show all applied
```

Expected: all migrations show `applied`.

- [ ] **Step 4: Commit migrations**

```bash
git add db/migrations/schema/postgres/
git commit -m "feat(db): add Postgres-compatible schema migrations for all domain tables"
```

---

### Task B4: Verify bun model compatibility with PostgreSQL

**Files:** domain model files in `internal/applications/*/models/`

- [ ] **Step 1: Check bun model tags**

```bash
grep -rn "bun:" internal/applications/*/models/*.go | grep -i "mysql\|mariadb\|auto_increment"
```

Look for any model tags that use MySQL-specific dialect features.

- [ ] **Step 2: Check for raw SQL in repositories**

```bash
grep -rn "\.Exec\|\.Query\|\.QueryRow" internal/applications/ | grep -v "_test.go"
```

Review any raw SQL for MySQL-specific functions (e.g., `NOW()` works in both; `IFNULL` → `COALESCE`, `DATE_FORMAT` → `TO_CHAR`, etc.).

- [ ] **Step 3: Update any MySQL-specific raw SQL**

Replace MySQL functions with Postgres equivalents where found. Common ones:
- `IFNULL(x, y)` → `COALESCE(x, y)` (bun handles this at ORM level, but check raw queries)
- `GROUP_CONCAT` → `STRING_AGG`
- `LIMIT x,y` → `LIMIT y OFFSET x`

- [ ] **Step 4: Commit any fixes**

```bash
git add internal/applications/
git commit -m "fix(db): update raw SQL to Postgres-compatible syntax"
```

---

### Task B5: Update seed files for PostgreSQL

**Files:** `db/migrations/seeds/`

- [ ] **Step 1: Check seed files for MySQL-specific syntax**

```bash
grep -rn "AUTO_INCREMENT\|\`\|ENGINE=" db/migrations/seeds/
```

- [ ] **Step 2: Fix any incompatibilities**

Same conversion rules as Task B2.

- [ ] **Step 3: Run seeds**

```bash
make seed-run
```

Expected: all seeds apply without errors.

- [ ] **Step 4: Commit fixes**

```bash
git add db/migrations/seeds/
git commit -m "fix(db): update seed files for Postgres compatibility"
```

---

### Task B6: Full integration test against PostgreSQL

- [ ] **Step 1: Run all tests against Postgres**

Ensure `config.local.yaml` has `database.default: "postgres"` and Postgres is running.

```bash
go test ./... -count=1 2>&1 | tail -30
```

Expected: all packages PASS. Note any failures.

- [ ] **Step 2: Start server and verify health endpoint**

```bash
go build -o /tmp/ichi-go-bin ./cmd/main.go
/tmp/ichi-go-bin &
sleep 2
curl -s http://localhost:8080/health | jq .
kill %1
```

Expected: `{"status": "ok"}` with all components healthy.

- [ ] **Step 3: Smoke test core API endpoints**

```bash
# Register + login
curl -s -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"Password123!"}' | jq .

curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Password123!"}' | jq .
```

Expected: `200 OK` responses.

- [ ] **Step 4: Commit**

```bash
git add .
git commit -m "feat: Phase B complete — all domains migrated to PostgreSQL as primary DB"
```

---

### Task B7: Clean up MySQL from config (optional)

Once Postgres is verified as primary and MySQL is no longer needed:

- [ ] Comment out `mysql` connection in `config.example.yaml` with a migration note
- [ ] Remove MySQL connection from `config.local.yaml`
- [ ] Commit: `chore(config): deprecate MySQL connection — Postgres is primary`

---

## Self-Review Checklist

### Phase A
- [x] **Spec §1 (Config):** Named connections map for both database and queue; driver names `"amqp"` / `"database"`
- [x] **Spec §2 (Database layer):** `NewMySQLClient`, `NewPostgresClient`, DSN helpers, `applyPoolSettings`
- [x] **Spec §3 (DI):** Named `*bun.DB` providers; `provideQueueInfra` for all queue connections
- [x] **Spec §4 (Queue interfaces):** `Dispatcher`, `JobArgs`, `ConsumeFunc`, `DispatchOption` (no UniqueKey)
- [x] **Spec §5 (River):** Single `BridgeWorker` with handler map; `RegisterBridgeWorkers` returns error
- [x] **Spec §6 (Factory):** `NewDispatcher` with `"amqp"` and `"database"` driver names
- [x] **Spec §7 (Dependencies):** All new packages installed
- [x] **migrate.go:** Auto-detects driver from `database.default`
- [x] **queue_server.go:** `StartQueueWorkers(*queue.QueueSchema)` — concurrent multi-driver
- [x] **cmd/main.go:** `cfg.Queue().AnyEnabled()` guard
- [x] **AMQP validation:** Fast-fail for unsupported options (`Queue`/`MaxAttempts`/`Priority`)

### Phase B
- [ ] `database.default` set to `"postgres"` in both example and local config
- [ ] All domain table migrations created with Postgres-compatible DDL
- [ ] All raw SQL in repositories updated for Postgres
- [ ] Seed files updated
- [ ] Full test suite passes against Postgres
- [ ] Health endpoint returns ok
- [ ] Core API endpoints smoke tested
