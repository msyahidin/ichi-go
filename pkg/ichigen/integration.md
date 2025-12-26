# Integration Guide

## Quick Setup

```bash
# 1. Build generator
go build -o ichigen ./pkg/ichigen/cmd/main.go

# 2. Generate module
./ichigen g full product --domain=catalog --crud

# 3. Wire in rest_server.go
```

## Wire in cmd/server/rest_server.go

```go
import "ichi-go/internal/applications/catalog"

func SetupRestRoutes(injector do.Injector, e *echo.Echo, cfg *config.Config) {
    // Existing
    auth.Register(injector, cfg.App().Name, e, authenticator)
    
    // Add new domain
    catalog.Register(injector, cfg.App().Name, e)
}
```

## Complete Model

Edit `repository/product_repository.go`:

```go
type ProductModel struct {
    bunBase.CoreModel
    bun.BaseModel `bun:"table:products,alias:product"`

    Name  string  `bun:"name,notnull"`
    Price float64 `bun:"price,notnull"`
    Stock int     `bun:"stock,default:0"`
}
```

## Complete DTOs

Edit `dto/product_dto.go`:

```go
type CreateProductRequest struct {
    Name  string  `json:"name" validate:"required,min=2,max=100"`
    Price float64 `json:"price" validate:"required,gt=0"`
    Stock int     `json:"stock" validate:"required,gte=0"`
}

func (r *CreateProductRequest) ToModel() *repository.ProductModel {
    return &repository.ProductModel{
        Name:  r.Name,
        Price: r.Price,
        Stock: r.Stock,
    }
}
```

## Test

```bash
curl -X POST http://localhost:8080/ichi-go/api/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Laptop","price":15000000,"stock":5}'
```

## Time Saved

Manual: **~2.5 hours** → Generator: **~30 minutes**