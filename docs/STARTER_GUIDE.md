# ichi-go Starter Guide

> Get the server running and make your first API call in 15 minutes.

## Prerequisites

| Requirement | Version | Notes |
|-------------|---------|-------|
| Go | 1.24+ | Required |
| MySQL | 8.0+ | Required |
| Redis | 7.0+ | Required |
| make | any | Required |
| RabbitMQ | 3.12+ | Optional — set `queue.enabled: false` to skip |

## Step 1: Clone and Configure (2 min)

```bash
git clone <repository-url> ichi-go
cd ichi-go
cp config.example.yaml config.local.yaml
```

Open `config.local.yaml` and fill in **three required values**:

```yaml
database:
  user: "your_db_user"
  password: "your_db_password"
  name: "your_db_name"
```

Everything else uses sane defaults for local development.

> **No RabbitMQ?** Add `queue.enabled: false` to `config.local.yaml` to skip queue setup.

## Step 2: Install Tools (2 min)

```bash
go mod download
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/air-verse/air@latest
```

To also use the code generator globally:
```bash
make ichigen-install
```

## Step 3: Set Up Database (3 min)

```bash
make migration-up   # creates all tables
make seed-run       # loads default roles, permissions, and test data
```

Verify everything looks good:
```bash
make db-status
```

## Step 4: Run the Server (1 min)

```bash
air
```

The server starts at `http://localhost:8080` with hot reload enabled.

API documentation is available at `http://localhost:8080/docs/index.html`

## Step 5: First API Calls (5 min)

Routes follow the format `/{app-name}/api/{version}/{domain}`. The default app name is `ichi-go` and the current API version is `202601`.

**Register a user:**
```bash
curl -s -X POST http://localhost:8080/ichi-go/api/202601/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "Dev User", "email": "dev@example.com", "password": "SecurePass123!"}' \
  | jq .
```

**Login and save the token:**
```bash
TOKEN=$(curl -s -X POST http://localhost:8080/ichi-go/api/202601/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "dev@example.com", "password": "SecurePass123!"}' \
  | jq -r '.data.access_token')

echo "Token: $TOKEN"
```

**Call a protected endpoint:**
```bash
curl -s http://localhost:8080/ichi-go/api/202601/auth/me \
  -H "Authorization: Bearer $TOKEN" \
  | jq .
```

**Check system health:**
```bash
curl -s http://localhost:8080/ichi-go/api/202601/health | jq .
```

## The 5 Folders You'll Touch Most

| Folder | Purpose |
|--------|---------|
| `internal/applications/{domain}/` | Feature code — controllers, services, repositories |
| `db/migrations/schema/` | Database table definitions |
| `config.local.yaml` | Your local settings (gitignored, never committed) |
| `internal/infra/queue/registry.go` | Register new queue message consumers |
| `cmd/server/rest_server.go` | Wire new domains into the HTTP server |

## Your First Feature: Add an Endpoint (5 min)

Use the code generator to scaffold a complete CRUD module:

```bash
# Generate a full domain with all layers
go run ./pkg/ichigen/cmd/main.go g full product --domain=catalog --crud

# Or, if you installed the generator globally:
ichigen g full product --domain=catalog --crud
```

This creates 7 files under `internal/applications/catalog/`:
- `controllers/product_controller.go` — HTTP handlers
- `services/product_service.go` — business logic (fill in the TODO stubs)
- `repositories/product_repository.go` — database access
- `dto/` — request/response models
- `validators/product_validators.go` — custom validation rules
- `providers.go` — dependency injection setup
- `registry.go` — route registration

**Three steps to wire it in:**

1. Register the domain in `cmd/server/rest_server.go`:
   ```go
   import catalogapp "ichi-go/internal/applications/catalog"

   // In SetupRestRoutes():
   catalogapp.Register(injector, cfg.App().Name, e, appAuth)
   ```

2. Create the database migration:
   ```bash
   make migration-create name=create_products_table
   # Edit the generated file in db/migrations/schema/
   ```

3. Run the migration:
   ```bash
   make migration-up
   ```

See [docs/ADDING_NEW_SERVICE.md](./ADDING_NEW_SERVICE.md) for the complete step-by-step walkthrough including testing patterns.

## Troubleshooting

| Problem | Fix |
|---------|-----|
| `connection refused` on database | Check `database.*` in `config.local.yaml`; verify MySQL is running |
| Migration fails | Run `mysql -u root -p` and confirm the database exists |
| Port 8080 already in use | Change `http.port` in `config.local.yaml` |
| RabbitMQ errors on start | Add `queue: enabled: false` to `config.local.yaml` |

## Next Steps

Read in this order:

1. **[README.md](../README.md)** — Complete feature documentation (JWT, validation, queues, DI patterns)
2. **[docs/ADDING_NEW_SERVICE.md](./ADDING_NEW_SERVICE.md)** — Full domain creation walkthrough with testing
3. **[docs/testing/TESTING_QUICKSTART.md](./testing/TESTING_QUICKSTART.md)** — Write your first test in 5 minutes
4. **[docs/RBAC.md](./RBAC.md)** — Understand the authorization system before adding protected routes
