# ichi-go

A production-ready Go backend template built on clean architecture principles, designed for rapid development of scalable REST API services with enterprise-grade features.

## 🎯 Overview

**ichi-go** accelerates backend development through proven architectural patterns, comprehensive tooling, and developer-friendly abstractions. Built for teams with varying skill levels, it reduces cognitive load while maintaining flexibility for complex business requirements.

### Key Highlights

- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **Production Ready**: JWT auth, validation, logging, error handling out of the box
- **Developer Experience**: CLI generators, hot reload, comprehensive examples
- **Message Queue**: RabbitMQ integration with graceful shutdown and worker management
- **Database**: Bun ORM with migration system and seeding support
- **Caching**: Redis with LZ4 compression
- **Validation**: Multilingual support (English/Indonesian) with custom validators
- **Observability**: Structured logging with zerolog, request tracing

## 📋 Table of Contents

- [Tech Stack](#-tech-stack)
- [Getting Started](#-getting-started)
- [Project Structure](#-project-structure)
- [Core Features](#-core-features)
- [Development Workflow](#-development-workflow)
- [Database Management](#-database-management)
- [Code Generation](#-code-generation)
- [Configuration](#-configuration)
- [Testing](#-testing)
- [Examples](#-examples)
- [Best Practices](#-best-practices)

## 🛠 Tech Stack

### Core Framework
- **Go 1.24+** - Primary language
- **Echo v4** - HTTP framework
- **Bun ORM** - Database operations with MySQL dialect
- **Viper** - Type-safe configuration management

### Infrastructure
- **Redis** - Caching with LZ4 compression
- **RabbitMQ** - Message queue with topic exchanges
- **MySQL** - Primary database

### Developer Tools
- **Air** - Hot reload for development
- **Goose** - SQL migrations
- **samber/do** - Runtime dependency injection
- **Mockery** - Interface mocking for tests

### Security & Validation
- **JWT** - Multiple algorithm support (HMAC, RSA, ECDSA)
- **go-playground/validator** - Request validation with i18n

### Observability
- **zerolog** - Structured logging
- **Request ID** - Distributed tracing support

## 🚀 Getting Started

### Prerequisites
```bash
# Required
go >= 1.24
mysql >= 8.0
redis >= 7.0
rabbitmq >= 3.12 (optional, if queue.enabled = true)

# Development tools
make
air  # for hot reload
```

### Installation

1. **Clone and setup**
```bash
   git clone <repository-url>
   cd ichi-go
   cp config.example.yaml config.local.yaml
```

2. **Configure environment**
```yaml
   # config.local.yaml
   database:
     host: "localhost"
     port: 3306
     user: "your_user"
     password: "your_password"
     name: "your_database"
   
   cache:
     host: "localhost"
     port: 6379
   
   queue:
     enabled: true  # Set false if RabbitMQ not needed
```

3. **Install dependencies**
```bash
   go mod download
   go install github.com/pressly/goose/v3/cmd/goose@latest
   go install github.com/vektra/mockery/v2@latest
   go install github.com/air-verse/air@latest
```

4. **Setup database**
```bash
   make migration-up        # Run migrations
   make seed-run           # Seed initial data
```

5. **Run the application**
```bash
   # Development with hot reload
   air
   
   # Or direct execution
   go run cmd/main.go
```

The server will start on `http://localhost:8080`

## 📁 Project Structure
```
ichi-go/
├── cmd/
│   ├── main.go                    # Application entry point
│   ├── server/
│   │   ├── rest_server.go        # HTTP routes setup
│   │   ├── queue_server.go       # RabbitMQ workers
│   │   └── web_server.go         # Web/template routes
│   └── validator/
│       └── setup.go              # Validator initialization
├── internal/
│   ├── applications/             # Domain applications
│   │   ├── auth/                # Authentication domain
│   │   │   ├── controllers/
│   │   │   ├── services/
│   │   │   ├── repositories/
│   │   │   ├── dto/
│   │   │   └── validators/
│   │   └── user/                # User management domain
│   ├── infra/                   # Infrastructure layer
│   │   ├── database/           # Bun ORM setup
│   │   ├── cache/              # Redis client
│   │   └── queue/              # RabbitMQ integration
│   │       ├── rabbitmq/
│   │       │   ├── connection.go
│   │       │   ├── producer.go
│   │       │   └── consumer.go
│   │       └── registry.go     # Consumer registration
│   └── middlewares/            # HTTP middlewares
├── db/
│   ├── cmd/
│   │   └── migrate.go          # Migration CLI
│   └── migrations/
│       ├── schema/             # Schema migrations
│       ├── data/               # Data migrations
│       └── seeds/              # Seed files
├── pkg/                        # Reusable packages
│   ├── authenticator/         # JWT implementation
│   ├── validator/             # Validation framework
│   ├── logger/                # Structured logging
│   ├── requestctx/            # Request context
│   └── utils/
│       ├── response/          # HTTP response builders
│       └── dto/               # DTO mapping
├── config/                    # Configuration schemas
│   ├── config.go             # Main config loader
│   ├── app/
│   └── http/
└── Makefile                  # Build automation
```

## 🎯 Core Features

### 1. JWT Authentication

Multi-algorithm support with automatic key management:
```yaml
# config.local.yaml
auth:
  jwt:
    enabled: true
    signing_method: "HS256"  # or RS256, ES256
    secret_key: "your-secret"
    access_token_ttl: "1h"
    refresh_token_ttl: "720h"
    skip_paths:
      - "/api/v1/auth/*"
      - "/api/v1/public/*"
```

**Supported Algorithms:**
- **HS256/HS384/HS512**: HMAC with secret key
- **RS256/RS384/RS512**: RSA with public/private keys
- **ES256/ES384/ES512**: ECDSA with public/private keys

### 2. Validation System

Domain-isolated validation with multilingual support:
```go
// Define custom validator
type EmailValidator struct {
    Tag: "email_domain",
    Fn: func(fl validator.FieldLevel) bool {
        // Custom validation logic
    },
    RegisterTrans: func(trans ut.Translator) error {
        return trans.Add("email_domain", 
            "Email must be from allowed domain", true)
    },
}

// Register in domain
validators.RegisterAuthValidators(validator)

// Use in handler
if err := validator.BindAndValidate(c, &req); err != nil {
    return response.Error(c, http.StatusBadRequest, err)
}
```

**Language Detection:**
- `X-Language` header (priority)
- `Accept-Language` header
- Falls back to configured default

### 3. Message Queue (RabbitMQ)

Producer/Consumer pattern with graceful shutdown:
```go
// Register consumer
queue.RegisterConsumer(queue.ConsumerRegistration{
    Name: "payment_handler",
    Description: "Processes payment events",
    ConsumeFunc: func(ctx context.Context, msg rabbitmq.Message) error {
        // Process message
        return nil
    },
})

// Publish message
producer.Publish(ctx, rabbitmq.PublishParams{
    RoutingKey: "payment.completed",
    Body: paymentData,
})
```

**Features:**
- Topic-based routing with delayed message support
- Configurable worker pools per consumer
- Automatic reconnection handling
- Context-based lifecycle management

### 4. Dependency Injection with samber/do

Runtime dependency injection with automatic lifecycle management:
```go
// Register dependencies
func Register(injector do.Injector, serviceName string, e *echo.Echo) {
    // Repositories
    do.ProvideNamed(injector, serviceName+".user.repository", 
        repositories.NewUserRepository)
    
    // Services
    do.ProvideNamed(injector, serviceName+".user.service",
        services.NewUserService)
    
    // Controllers
    do.ProvideNamed(injector, serviceName+".user.controller",
        controllers.NewUserController)
}

// Invoke dependencies
userService := do.MustInvokeNamed[*services.UserService](
    injector, 
    serviceName+".user.service",
)
```

**Benefits:**
- No code generation required
- Runtime flexibility
- Automatic shutdown handling via `injector.Shutdown()`
- Named dependencies for multi-tenancy support

### 5. Request Context

Centralized request metadata:
```go
type RequestContext struct {
    RequestID     string
    UserID        string
    Language      string
    ClientIP      string
    ValidatedClaims any
}

// Extract from Echo context
rc := requestctx.FromRequest(c.Request())

// Use in business logic
userID := requestctx.GetUserID(ctx)
```

## 💻 Development Workflow

### Hot Reload
```bash
# Start with Air (auto-reload on file changes)
air

# Air watches:
# - Go files
# - Config files (*.yaml)
# - Templates (*.html)
```

### Code Generation

#### Generate CRUD Module
```bash
# Full CRUD with all components
make ichigen-example-full

# Or using the CLI directly
go run ./pkg/ichigen/cmd/main.go g full product --domain=catalog --crud

# Creates:
# - Controller (HTTP handlers)
# - Service (business logic)
# - Repository (data access)
# - DTO (request/response models)
# - Validators (domain-specific validation)
# - Dependency injection setup with samber/do
```

#### Generate Single Component
```bash
# Generate controller only
go run ./pkg/ichigen/cmd/main.go g controller notification --domain=alert

# Or use Makefile examples
make ichigen-example-controller
```

#### Install Generator Globally
```bash
# Build and install
make ichigen-install

# Now use anywhere
ichigen g full order --domain=sales --crud
```

### Mock Generation
```bash
# Generate all mocks
mockery --all --dir internal/applications --output mocks --keeptree

# Generate specific interface
mockery --name=UserRepository --dir internal/applications/user
```

## 🗄 Database Management

### Migrations

#### Schema Migrations (DDL)
```bash
# Create migration
make migration-create name=create_products_table

# Run migrations
make migration-up

# Rollback last
make migration-down

# Status
make migration-status

# Migrate to specific version
make migration-up-to version=20250407085044

# Reset all (DANGEROUS!)
make migration-reset
```

#### Data Migrations (DML)
```bash
# Create data migration
make data-migration-create name=fix_legacy_emails

# Run data migrations
make data-migration-up

# Check status
make data-migration-status
```

### Seeding
```bash
# Run all seeds
make seed-run

# Run specific seed
make seed-file name=00_base_roles.sql

# List available seeds
make seed-list
```

### Complete Database Reset
```bash
# Reset + Migrate + Seed (development only)
make db-reset-dev

# Fresh setup (no reset)
make db-fresh

# Complete status
make db-status
```

## 📝 Configuration

### Environment-based Config
```
config.local.yaml    # Local development
config.dev.yaml      # Development server
config.staging.yaml  # Staging environment
config.prod.yaml     # Production
```

Set environment:
```bash
export APP_ENV=prod
go run cmd/main.go
```

### Configuration Structure
```yaml
app:
  env: "local"
  name: "ichi-go"
  debug: true

http:
  port: 8080
  timeout: 10000
  cors:
    allow_origins: ["*"]

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  max_idle_conns: 10
  max_open_conns: 100

cache:
  driver: "redis"
  host: "localhost"
  pool_size: 20

validator:
  default_language: "en"
  supported_languages: ["en", "id"]

queue:
  enabled: true
  rabbitmq:
    exchanges:
      - name: "app.events"
        type: "x-delayed-message"
    consumers:
      - name: "payment_handler"
        queue:
          name: "order.payment.events"
        routing_keys:
          - "payment.completed"
        prefetch_count: 50
        worker_pool_size: 10
```

## 🧪 Testing

### Unit Tests
```bash
# Run all tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./internal/applications/user/services/...

# Verbose output
go test -v ./...
```

### Using Mocks
```go
func TestUserService_Create(t *testing.T) {
    mockRepo := new(mock_user.UserRepository)
    service := services.NewUserService(mockRepo)
    
    mockRepo.On("Create", mock.Anything, mock.Anything).
        Return(&user.User{ID: 1}, nil)
    
    result, err := service.Create(context.Background(), dto)
    
    assert.NoError(t, err)
    assert.Equal(t, 1, result.ID)
    mockRepo.AssertExpectations(t)
}
```

## 📚 Examples

### Complete CRUD Example

See `/internal/applications/user` for reference implementation:
```
user/
├── controllers/
│   └── user_controller.go      # HTTP handlers
├── services/
│   └── user_service.go         # Business logic
├── repositories/
│   └── user_repository.go      # Data access
├── dto/
│   ├── create_user_dto.go
│   └── update_user_dto.go
├── validators/
│   └── user_validators.go      # Custom validation
├── models/
│   └── user.go                 # Domain model
└── register.go                 # DI registration with samber/do
```

### JWT Authentication Flow
```go
// 1. Login endpoint
func (c *AuthController) Login(ctx echo.Context) error {
    var req dto.LoginRequest
    if err := validator.BindAndValidate(ctx, &req); err != nil {
        return response.Error(ctx, http.StatusBadRequest, err)
    }
    
    token, err := c.service.Login(ctx.Request().Context(), req)
    if err != nil {
        return response.Error(ctx, http.StatusUnauthorized, err)
    }
    
    return response.Success(ctx, token)
}

// 2. Protected endpoint
func (c *UserController) GetProfile(ctx echo.Context) error {
    userID := requestctx.GetUserID(ctx.Request().Context())
    profile, err := c.service.GetByID(ctx.Request().Context(), userID)
    return response.Success(ctx, profile)
}
```

### Queue Consumer Example
```go
// Register in internal/infra/queue/registry.go
func init() {
    RegisterConsumer(ConsumerRegistration{
        Name: "welcome_notifier",
        Description: "Sends welcome emails to new users",
        ConsumeFunc: handleWelcomeNotification,
    })
}

func handleWelcomeNotification(ctx context.Context, msg rabbitmq.Message) error {
    var userData UserCreatedEvent
    if err := json.Unmarshal(msg.Body, &userData); err != nil {
        return err
    }
    
    // Send email logic
    logger.Infof("Sending welcome email to %s", userData.Email)
    return nil
}
```

### Dependency Injection Pattern
```go
// In register.go
func Register(injector do.Injector, serviceName string, e *echo.Echo, auth *authenticator.Authenticator) {
    // Register repository with named dependency
    do.ProvideNamed(injector, serviceName+".user.repository", 
        func(i do.Injector) (*repositories.UserRepository, error) {
            db := do.MustInvoke[*database.DB](i)
            return repositories.NewUserRepository(db), nil
        })
    
    // Register service with dependency on repository
    do.ProvideNamed(injector, serviceName+".user.service",
        func(i do.Injector) (*services.UserService, error) {
            repo := do.MustInvokeNamed[*repositories.UserRepository](
                i, serviceName+".user.repository")
            return services.NewUserService(repo), nil
        })
    
    // Register controller and setup routes
    do.ProvideNamed(injector, serviceName+".user.controller",
        func(i do.Injector) (*controllers.UserController, error) {
            service := do.MustInvokeNamed[*services.UserService](
                i, serviceName+".user.service")
            return controllers.NewUserController(service), nil
        })
    
    // Setup routes
    setupRoutes(injector, serviceName, e, auth)
}
```

## 🎓 Best Practices

### 1. Error Handling
```go
// Use typed errors
var ErrUserNotFound = errors.New("user not found")

// Return domain errors
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}
```

### 2. Validation
```go
// Define DTOs with validation tags
type CreateUserDTO struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
}

// Use BindAndValidate in handlers
if err := validator.BindAndValidate(c, &dto); err != nil {
    return response.Error(c, http.StatusBadRequest, err)
}
```

### 3. Logging
```go
// Use contextual logging
logger.Infof("Processing order %s", orderID)
logger.Debugf("User data: %+v", user)
logger.Errorf("Database error: %v", err)

// With request context
logger.WithContext(ctx).Infof("Request processed")
```

### 4. Dependency Injection with samber/do
```go
// Always use named dependencies for domain isolation
serviceName := "myapp"
do.ProvideNamed(injector, serviceName+".user.repository", ...)

// Use MustInvoke for required dependencies
db := do.MustInvoke[*database.DB](injector)

// Proper shutdown in main.go
defer injector.Shutdown() // Cleans up all resources
```

### 5. Testing Strategy

- **Unit tests**: Test services with mocked repositories
- **Integration tests**: Test repositories with test database
- **E2E tests**: Test complete request flow

### 6. Queue Consumer Best Practices
```go
// Always handle context cancellation
func handleMessage(ctx context.Context, msg rabbitmq.Message) error {
    select {
    case <-ctx.Done():
        return ctx.Err() // Graceful shutdown
    default:
        // Process message
    }
}

// Use proper error handling for retry logic
func handleMessage(ctx context.Context, msg rabbitmq.Message) error {
    if err := process(msg); err != nil {
        if isRetryable(err) {
            return err // Will be requeued
        }
        logger.Errorf("Non-retryable error: %v", err)
        return nil // Ack to prevent infinite retry
    }
    return nil
}
```

## 📖 Additional Resources

### Documentation
- [Echo Framework Docs](https://echo.labstack.com/)
- [Bun ORM Guide](https://bun.uptrace.dev/)
- [samber/do Documentation](https://github.com/samber/do)
- [Go-Playground Validator](https://pkg.go.dev/github.com/go-playground/validator/v10)
- [RabbitMQ Tutorials](https://www.rabbitmq.com/tutorials)

### Common Commands Reference
```bash
# Development
make ichigen-build              # Build CLI generator
make ichigen-install            # Install generator globally
make migration-help             # Show all migration commands
make db-status                  # Complete database status

# Testing
go test ./...                   # Run all tests
go test -race ./...            # Race condition detection
go test -bench=.               # Run benchmarks

# Code Quality
go vet ./...                    # Static analysis
go fmt ./...                    # Format code
golangci-lint run              # Comprehensive linting
```

## 🤝 Contributing

1. Follow the existing code structure
2. Write tests for new features
3. Update documentation
4. Use conventional commit messages
5. Ensure all tests pass before submitting PR

## 📄 License

MIT License

Copyright (c) 2025 ichi-go contributors

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

---

**Built with ❤️ for teams that ship fast without compromising quality**
