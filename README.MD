

# ichi-go-template

Boilerplate backend Go project based on clean architecture, built to accelerate backend service development with scalability, modularity, performance in mind, and DX (Developer Experience).

> Optimized for REST API services with Ent ORM, Echo, Redis, and RabbitMQ integration.

## Contents
- [Getting Started](#getting-started)
  - [Install & Generate Ent ORM](#install--generate-ent-orm)
  - [Generate Mocks with Mockery](#generate-mocks-with-mockery)
  - [Generate Dependency Injection](#generate-dependency-injection)
  - [Migration](#migration)
  - [Run Configuration](#run-configuration)
- [Build Project](#build-project)
- [Example Code](#example-code)
- [References](#references)
- [Features](#features)

## Getting Started

### Install & Generate Ent ORM

```bash
go install entgo.io/ent/cmd/ent@latest
go run entgo.io/ent/cmd/ent init User
go generate ./internal/infra/database/ent
```

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
- `/internal/infra/database/ent`: Ent schema and generated files
- `/internal/infra/cache`: Redis implementation with LZ4 compression
- `/internal/infra/rabbitmq`: RabbitMQ integration (via Watermill)
- `/pkg/clients/pokemonapi`: External API client example

## References

- [Ent ORM](https://entgo.io/)
- [Echo Framework](https://echo.labstack.com/)
- [Viper](https://github.com/spf13/viper)
- [Google Wire](https://github.com/google/wire)
- [Swaggo](https://github.com/swaggo/swag)
- [Mockery](https://github.com/vektra/mockery)
- [Goose Migration](https://github.com/pressly/goose)
- [Zerolog](https://github.com/rs/zerolog)
- [Watermill](https://watermill.io/pubsubs/amqp/)

## Features

- [x] Clean architecture folder structure
- [x] REST API support with Echo
- [x] Ent ORM with versioning and hook support
- [x] Redis cache with LZ4 compression
- [x] Dependency injection with Google Wire
- [x] Logging with Zerolog
- [x] SQL-based migrations via Goose
- [X] Live reload with Air
- [ ] RabbitMQ pub/sub with Watermill (planned)
- [ ] Swagger documentation (planned)
- [ ] Interface mocking with Mockery (planned)
- [ ] Docker support (planned)
