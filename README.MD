# ichi-go-template

Boilerplate backend Go project based on clean architecture, built to accelerate backend service development with scalability, modularity, performance in mind, and DX (Developer Experience).

> Optimized for REST API services with Bun ORM, Echo, Redis, and RabbitMQ integration.

## Contents
- [Getting Started](#getting-started)
  - [Generate Mocks with Mockery](#generate-mocks-with-mockery)
  - [Generate Dependency Injection](#generate-dependency-injection)
  - [Migration](#migration)
  - [Run Configuration](#run-configuration)
- [Build Project](#build-project)
- [Example Code](#example-code)
- [References](#references)
- [Features](#features)

## Getting Started

### Generate Mocks with Mockery

```bash
go install github.com/vektra/mockery/v2@v2.32.0
mockery --all --dir internal/applications --output mocks --keeptree --packageprefix mock_
```

### Generate Dependency Injection

```bash
go install github.com/google/wire/cmd/wire@latest
make wire
```

### Migration

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
make migration-create name=create_users_table type=sql
make migration-up
make migration-status
```

### Run Configuration

```bash
cp config.example.yaml config.local.yaml
go run cmd/main.go
// or execute with air for live reload

air
```

## Build Project

Run all tasks including code generation, test, and binary build:

```bash
make build
```

## Example Code

- `/internal/applications/user`: Domain layer (controller, service, repository, DTO)
- `/internal/infra/database/bun`: Bun as ORM layer
- `/internal/infra/cache`: Redis implementation with LZ4 compression
- `/internal/infra/rabbitmq`: RabbitMQ integration (via Watermill)
- `/pkg/clients/pokemonapi`: External API client example

## References

- [Echo Framework](https://echo.labstack.com/)
- [Viper](https://github.com/spf13/viper)
- [Google Wire](https://github.com/google/wire)
- [Swaggo](https://github.com/swaggo/swag)
- [Mockery](https://github.com/vektra/mockery)
- [Goose Migration](https://github.com/pressly/goose)
- [Zerolog](https://github.com/rs/zerolog)

## Features

- [x] Clean architecture folder structure
- [x] REST API support with Echo
- [x] Uptrace Bun ORM
- [x] Redis cache with LZ4 compression
- [x] Dependency injection with Google Wire
- [x] Contextual Logging with Zerolog
- [X] Type safety configuration with Viper
- [x] SQL-based migrations via Goose
- [X] Live reload with Air
- [X] RabbitMQ as Messaging
- [ ] Delay, DLQ, Retry pattern
- [ ] Database Versioning with Retry
- [ ] Swagger documentation
- [ ] Interface mocking with Mockery
- [ ] Docker support
