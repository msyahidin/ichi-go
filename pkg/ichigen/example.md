# Complete Example: Product CRUD

## Generate

```bash
./ichigen g full product --domain=catalog --crud
```

## Files Created

```
internal/applications/catalog/
├── controller/product_controller.go    ✓ Ready
├── dto/product_dto.go                 → Complete fields
├── repository/product_repository.go    → Complete model
├── service/product_service.go          ✓ Ready
├── validators/product_validator.go     ✓ Ready
├── providers.go                        ✓ Ready
└── registry.go                         ✓ Ready
```

## Step 1: Complete Model

`repository/product_repository.go`:

```go
type ProductModel struct {
    bunBase.CoreModel
    bun.BaseModel `bun:"table:products,alias:product"`

    Name        string  `bun:"name,notnull"`
    Description string  `bun:"description"`
    Price       float64 `bun:"price,notnull"`
    Stock       int     `bun:"stock,notnull,default:0"`
}
```

## Step 2: Complete DTOs

`dto/product_dto.go`:

```go
type CreateProductRequest struct {
    Name        string  `json:"name" validate:"required,min=2,max=100"`
    Description string  `json:"description" validate:"max=500"`
    Price       float64 `json:"price" validate:"required,gt=0"`
    Stock       int     `json:"stock" validate:"required,gte=0"`
}

func (r *CreateProductRequest) ToModel() *repository.ProductModel {
    return &repository.ProductModel{
        Name:        r.Name,
        Description: r.Description,
        Price:       r.Price,
        Stock:       r.Stock,
    }
}

type ProductResponse struct {
    ID          int64     `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Price       float64   `json:"price"`
    Stock       int       `json:"stock"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

## Step 3: Wire Up

`cmd/server/rest_server.go`:

```go
import "ichi-go/internal/applications/catalog"

func SetupRestRoutes(injector do.Injector, e *echo.Echo, cfg *config.Config) {
    // ... existing
    catalog.Register(injector, cfg.App().Name, e)
}
```

## Test

```bash
# Create
curl -X POST http://localhost:8080/ichi-go/api/products \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Laptop Gaming",
    "description": "High performance",
    "price": 15000000,
    "stock": 5
  }'

# List
curl "http://localhost:8080/ichi-go/api/products?page=1&limit=10"

# Get
curl http://localhost:8080/ichi-go/api/products/1

# Update
curl -X PUT http://localhost:8080/ichi-go/api/products/1 \
  -d '{"stock": 3}'

# Delete
curl -X DELETE http://localhost:8080/ichi-go/api/products/1
```

## Time: 30 minutes vs 2.5 hours manual