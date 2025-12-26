# Generator for ichi-go

Angular-like CLI generator for Go projects using ichi-go architecture.

## Install

```bash
go build -o ichigen ./pkg/ichigen/cmd/main.go
sudo mv gen /usr/local/bin/  # Optional: install globally
```

## Usage

```bash
# Generate full CRUD
ichigen g full product --domain=catalog --crud

# Generate single component
ichigen g controller notification --domain=alert
ichigen g service order --domain=sales
ichigen g repository user --domain=auth
```

## Generated Structure

```
internal/applications/catalog/
├── controller/product_controller.go
├── dto/product_dto.go
├── repository/product_repository.go
├── service/product_service.go
├── validators/product_validator.go
├── providers.go
└── registry.go
```

## Integration

1. **Wire in rest_server.go:**
```go
import "ichi-go/internal/applications/catalog"

catalog.Register(injector, cfg.App().Name, e)
```

2. **Complete TODOs in generated files**
3. **Test endpoints**

See `INTEGRATION.md` for detailed setup.