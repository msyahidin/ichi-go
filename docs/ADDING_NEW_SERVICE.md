# Adding a New Service/Domain with Testing

This guide shows how to add a new service/domain to the project with proper testing setup.

## Example: Adding a Notification Service

### 1. Create Domain Structure

Use the code generator or create manually:

```bash
# Using ichigen (recommended)
make ichigen-example-full  # See example, then adapt
go run ./pkg/ichigen/cmd/main.go g full notification --domain=messaging --crud

# Manual structure
internal/applications/notification/
├── controller/
│   └── notification_controller.go
├── service/
│   └── notification_service.go
├── repository/
│   └── notification_repository.go
├── dto/
│   ├── notification_dto.go
│   └── ...
├── models/
│   └── notification.go
└── register.go
```

### 2. Define Interfaces

**service/notification_service.go:**
```go
package service

import "context"

// NotificationService handles notification operations
type NotificationService interface {
    Send(ctx context.Context, req SendNotificationRequest) error
    GetHistory(ctx context.Context, userID uint64) ([]*Notification, error)
}

type NotificationServiceImpl struct {
    repo NotificationRepository
}

func NewNotificationService(repo NotificationRepository) NotificationService {
    return &NotificationServiceImpl{repo: repo}
}
```

**repository/notification_repository.go:**
```go
package repository

import "context"

// NotificationRepository handles notification data access
type NotificationRepository interface {
    Create(ctx context.Context, notif *Notification) error
    FindByUserID(ctx context.Context, userID uint64) ([]*Notification, error)
}

type NotificationRepositoryImpl struct {
    db *database.DB
}

func NewNotificationRepository(db *database.DB) NotificationRepository {
    return &NotificationRepositoryImpl{db: db}
}
```

### 3. Add to .mockery.yaml

Edit `.mockery.yaml` and add your new service:

```yaml
packages:
  # ... existing packages ...

  # Notification Service (NEW)
  ./internal/applications/notification/service:
    config:
      all: true

  ./internal/applications/notification/repository:
    config:
      all: true
```

**Note:** Uses relative paths (`./internal/...`) so it works regardless of your module name (notification-service, cart-service, etc.)

### 4. Generate Mocks

```bash
make test-generate-mocks
```

This creates:
- `internal/applications/notification/service/mocks/mock_notification_service.go`
- `internal/applications/notification/repository/mocks/mock_notification_repository.go`

### 5. Write Service Tests

**service/notification_service_test.go:**
```go
package service

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "{your-module}/internal/applications/notification/repository/mocks"
    "{your-module}/internal/testutil"
)

func TestNotificationService_Send(t *testing.T) {
    tests := []struct {
        name        string
        request     SendNotificationRequest
        setupMock   func(*mocks.MockNotificationRepository)
        expectError bool
    }{
        {
            name: "Success - Send notification",
            request: SendNotificationRequest{
                UserID:  1,
                Message: "Test notification",
                Type:    "email",
            },
            setupMock: func(m *mocks.MockNotificationRepository) {
                m.EXPECT().
                    Create(mock.Anything, mock.AnythingOfType("*Notification")).
                    Return(nil)
            },
            expectError: false,
        },
        {
            name: "Error - Repository failure",
            request: SendNotificationRequest{
                UserID:  1,
                Message: "Test",
                Type:    "email",
            },
            setupMock: func(m *mocks.MockNotificationRepository) {
                m.EXPECT().
                    Create(mock.Anything, mock.AnythingOfType("*Notification")).
                    Return(errors.New("database error"))
            },
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockRepo := mocks.NewMockNotificationRepository(t)
            tt.setupMock(mockRepo)

            service := NewNotificationService(mockRepo)
            ctx := testutil.NewTestContext(t)

            // Act
            err := service.Send(ctx, tt.request)

            // Assert
            if tt.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

func TestNotificationService_GetHistory(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockNotificationRepository(t)
    service := NewNotificationService(mockRepo)
    ctx := testutil.NewTestContext(t)

    expected := []*Notification{
        {ID: 1, UserID: 1, Message: "Test 1"},
        {ID: 2, UserID: 1, Message: "Test 2"},
    }

    mockRepo.EXPECT().
        FindByUserID(ctx, uint64(1)).
        Return(expected, nil)

    // Act
    result, err := service.GetHistory(ctx, 1)

    // Assert
    require.NoError(t, err)
    assert.Len(t, result, 2)
    assert.Equal(t, expected, result)
}
```

### 6. Write Repository Integration Tests

**repository/notification_repository_integration_test.go:**
```go
package repository

import (
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/suite"

    "{your-module}/internal/testutil"
)

type NotificationRepositoryIntegrationTestSuite struct {
    suite.Suite
    db   *sql.DB
    repo NotificationRepository
    ctx  context.Context
}

func (suite *NotificationRepositoryIntegrationTestSuite) SetupSuite() {
    if testing.Short() {
        suite.T().Skip("Skipping integration tests")
    }

    mysql := testutil.SetupMySQLContainer(suite.T())
    suite.db = mysql.GetDB(suite.T())
    suite.repo = NewNotificationRepository(suite.db)
    suite.ctx = context.Background()
}

func (suite *NotificationRepositoryIntegrationTestSuite) SetupTest() {
    testutil.TruncateTables(suite.T(), suite.db)
}

func (suite *NotificationRepositoryIntegrationTestSuite) TearDownSuite() {
    if suite.db != nil {
        suite.db.Close()
    }
}

func (suite *NotificationRepositoryIntegrationTestSuite) TestCreate_Success() {
    notif := &Notification{
        UserID:  1,
        Message: "Test notification",
        Type:    "email",
    }

    err := suite.repo.Create(suite.ctx, notif)
    require.NoError(suite.T(), err)
}

func (suite *NotificationRepositoryIntegrationTestSuite) TestFindByUserID() {
    // Create test notifications
    notif1 := &Notification{UserID: 1, Message: "Test 1"}
    notif2 := &Notification{UserID: 1, Message: "Test 2"}

    suite.repo.Create(suite.ctx, notif1)
    suite.repo.Create(suite.ctx, notif2)

    // Find by user ID
    results, err := suite.repo.FindByUserID(suite.ctx, 1)

    require.NoError(suite.T(), err)
    assert.Len(suite.T(), results, 2)
}

func TestNotificationRepositoryIntegrationTestSuite(t *testing.T) {
    suite.Run(t, new(NotificationRepositoryIntegrationTestSuite))
}
```

### 7. Write Controller Tests

**controller/notification_controller_test.go:**
```go
package controller

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "{your-module}/internal/applications/notification/service/mocks"
    "{your-module}/internal/testutil"
)

func TestNotificationController_Send(t *testing.T) {
    // Setup
    e := echo.New()
    mockService := mocks.NewMockNotificationService(t)
    controller := NewNotificationController(mockService)

    req := SendNotificationRequest{
        UserID:  1,
        Message: "Test",
        Type:    "email",
    }

    mockService.EXPECT().
        Send(mock.Anything, req).
        Return(nil)

    // Create HTTP request
    httpReq := httptest.NewRequest(http.MethodPost, "/notifications", createBody(req))
    httpReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(httpReq, rec)

    // Execute
    err := controller.Send(c)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

### 8. Run Tests

```bash
# Run all notification tests
go test ./internal/applications/notification/... -v

# With coverage
go test ./internal/applications/notification/... -cover

# Skip integration tests
go test -short ./internal/applications/notification/... -v

# Run all tests
make test
```

### 9. Check Coverage

```bash
make test-coverage
open coverage.html
```

Target coverage:
- Service: 80%+
- Repository: 70%+
- Controller: 70%+

## Example: Adding a Cart Service

Same steps, just replace "notification" with "cart":

### .mockery.yaml
```yaml
packages:
  # ... existing packages ...

  # Cart Service (NEW)
  ./internal/applications/cart/service:
    config:
      all: true

  ./internal/applications/cart/repository:
    config:
      all: true
```

### Generate Mocks
```bash
make test-generate-mocks
```

Mocks will be created in:
- `internal/applications/cart/service/mocks/`
- `internal/applications/cart/repository/mocks/`

## Quick Checklist for New Service

- [ ] Create domain structure (controller, service, repository, dto, models)
- [ ] Define service and repository interfaces
- [ ] Add packages to `.mockery.yaml` using relative paths (`./internal/...`)
- [ ] Run `make test-generate-mocks`
- [ ] Write service unit tests with mocks
- [ ] Write repository integration tests
- [ ] Write controller tests
- [ ] Run `make test-coverage` and verify >70% coverage
- [ ] Add to main test suite with `make test`

## Tips for Any Project Name

### Works with ANY module name:

```yaml
# .mockery.yaml - Same config works for all projects
packages:
  ./internal/applications/notification/service:    # ✅ Works with any module
    config:
      all: true
```

### NOT this (hardcoded module name):

```yaml
# ❌ DON'T do this - only works with "ichi-go" module
packages:
  ichi-go/internal/applications/notification/service:
    config:
      all: true
```

### Import paths in tests:

```go
// Replace {your-module} with your actual module name from go.mod
import "{your-module}/internal/applications/notification/service/mocks"

// Examples:
// import "notification-service/internal/applications/notification/service/mocks"
// import "cart-service/internal/applications/cart/service/mocks"
// import "ichi-go/internal/applications/auth/service/mocks"
```

## Common Patterns

### Service Test Template
```go
func TestService_Method(t *testing.T) {
    mockRepo := mocks.NewMockRepository(t)
    service := NewService(mockRepo)
    ctx := testutil.NewTestContext(t)

    mockRepo.EXPECT().Method(ctx, arg).Return(result, nil)

    result, err := service.Method(ctx, arg)
    require.NoError(t, err)
}
```

### Controller Test Template
```go
func TestController_Handler(t *testing.T) {
    e := echo.New()
    mockService := mocks.NewMockService(t)
    controller := NewController(mockService)

    mockService.EXPECT().Method(mock.Anything, arg).Return(result, nil)

    req := httptest.NewRequest(http.MethodPost, "/path", body)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    err := controller.Handler(c)
    require.NoError(t, err)
}
```

### Integration Test Template
```go
func TestRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    mysql := testutil.SetupMySQLContainer(t)
    db := mysql.GetDB(t)
    repo := NewRepository(db)
    ctx := testutil.NewTestContext(t)

    // Test code
}
```

## Resources

- [TESTING.md](testing/TESTING.md) - Complete testing guide
- [TESTING_QUICKSTART.md](testing/TESTING_QUICKSTART.md) - Quick start
- [TESTING_CHEATSHEET.md](testing/TESTING_CHEATSHEET.md) - Quick reference
- Example: `internal/applications/auth/` - See auth domain for examples

---

**Remember:** The key to portability is using relative paths (`./internal/...`) in `.mockery.yaml` instead of absolute module paths!
