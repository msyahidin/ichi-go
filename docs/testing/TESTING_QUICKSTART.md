# Testing Quick Start Guide

Get started with testing in ichi-go in 5 minutes!

## Prerequisites

```bash
# Install mockery (if not already installed)
make test-install-mockery
```

## Quick Commands

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Generate mocks
make test-generate-mocks

# Run specific domain tests
make test-auth

# Run with race detector
make test-race

# Run benchmarks
make test-bench
```

## File Structure Created

```
.
├── .mockery.yaml                    # Mockery configuration
├── TESTING.md                       # Comprehensive testing guide
├── TESTING_QUICKSTART.md           # This file
│
├── internal/
│   ├── testutil/                   # Shared test utilities
│   │   ├── helpers.go             # Test helper functions
│   │   ├── fixtures.go            # Test data builders
│   │   └── containers.go          # Test container setup
│   │
│   ├── middlewares/
│   │   └── version_middleware_test.go    # Example middleware test
│   │
│   └── applications/
│       ├── auth/
│       │   └── controller/
│       │       └── auth_controller_test.go  # Example controller test
│       │
│       └── user/
│           └── repository/
│               └── user_repository_integration_test.go  # Example integration test
│
└── Makefile                        # Updated with test commands
```

## Step-by-Step: Write Your First Test

### 1. Unit Test (Service Layer)

```go
// internal/applications/myapp/service/myservice_test.go
package service

import (
    "context"
    "testing"
    "ichi-go/internal/testutil"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMyService_DoSomething(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockRepository(t)
    service := NewMyService(mockRepo)
    ctx := testutil.NewTestContext(t)

    mockRepo.EXPECT().
        FindByID(ctx, uint64(1)).
        Return(&Model{ID: 1}, nil)

    // Act
    result, err := service.DoSomething(ctx, 1)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, uint64(1), result.ID)
}
```

### 2. Integration Test (Repository Layer)

```go
// internal/applications/myapp/repository/myrepo_integration_test.go
package repository

import (
    "testing"
    "ichi-go/internal/testutil"
    "github.com/stretchr/testify/require"
)

func TestMyRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup database
    mysql := testutil.SetupMySQLContainer(t)
    db := mysql.GetDB(t)
    repo := NewMyRepository(db)
    ctx := testutil.NewTestContext(t)

    // Test
    id, err := repo.Create(ctx, &Model{Name: "test"})
    require.NoError(t, err)
    assert.Greater(t, id, int64(0))
}
```

### 3. Controller Test

```go
// internal/applications/myapp/controller/mycontroller_test.go
package controller

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/require"
)

func TestMyController_GetItem(t *testing.T) {
    // Setup
    e := echo.New()
    mockService := mocks.NewMockService(t)
    controller := NewMyController(mockService)

    // Mock expectations
    mockService.EXPECT().
        GetByID(mock.Anything, uint64(1)).
        Return(&Response{ID: 1}, nil)

    // Create request
    req := httptest.NewRequest(http.MethodGet, "/items/1", nil)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)
    c.SetParamNames("id")
    c.SetParamValues("1")

    // Execute
    err := controller.GetItem(c)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

## Generate Mocks for Your Interfaces

### 1. Define interfaces in your code

```go
// internal/applications/myapp/repository/interface.go
package repository

type MyRepository interface {
    Create(ctx context.Context, model *Model) (int64, error)
    FindByID(ctx context.Context, id uint64) (*Model, error)
}
```

### 2. Update .mockery.yaml

```yaml
packages:
  # Use relative paths - works with any project name
  ./internal/applications/myapp/repository:
    config:
      all: true  # Generate mocks for all interfaces in this package

  # Or specify specific interfaces:
  ./internal/applications/myapp/service:
    interfaces:
      MyService:
      AnotherService:
```

### 3. Generate mocks

```bash
make test-generate-mocks
```

### 4. Use generated mocks

```go
// Mocks will be in {package}/mocks directory
import "{your-module-name}/internal/applications/myapp/repository/mocks"

func TestWithMock(t *testing.T) {
    mockRepo := mocks.NewMockMyRepository(t)

    mockRepo.EXPECT().
        FindByID(mock.Anything, uint64(1)).
        Return(&Model{ID: 1}, nil)

    // Use mockRepo in your tests
}
```

## Using Test Utilities

### Test Helpers

```go
import "ichi-go/internal/testutil"

// Create test context with timeout
ctx := testutil.NewTestContext(t)

// Assertions
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, expected, actual)

// Pointers
email := testutil.StringPtr("test@example.com")
id := testutil.Uint64Ptr(123)
```

### Test Fixtures

```go
// Create user fixture
user := testutil.NewUserFixture().
    WithEmail("custom@example.com").
    WithPassword("CustomPass123!")

// Use in tests
assert.Equal(t, "custom@example.com", user.Email)
```

### Test Containers (Integration Tests)

```go
// Setup MySQL for integration tests
mysql := testutil.SetupMySQLContainer(t)
db := mysql.GetDB(t)

// Setup Redis
redis := testutil.SetupRedisContainer(t)
addr := redis.GetAddress()

// Cleanup is automatic via t.Cleanup()
```

## Best Practices Checklist

- [ ] Use table-driven tests for multiple scenarios
- [ ] Test both success and error cases
- [ ] Use `require` for critical checks, `assert` for validations
- [ ] Always pass context to functions
- [ ] Mock external dependencies
- [ ] Run integration tests with `-short` flag support
- [ ] Add benchmarks for critical paths
- [ ] Clean up test data with `t.Cleanup()`
- [ ] Use fixtures for test data
- [ ] Generate mocks with mockery

## Test Coverage Goals

| Layer | Target |
|-------|--------|
| Services | 80%+ |
| Repositories | 70%+ |
| Controllers | 70%+ |
| Middlewares | 80%+ |
| Core Packages | 90%+ |

## Common Test Patterns

### Table-Driven Test

```go
func TestSomething(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectError bool
    }{
        {"Valid input", "test", "TEST", false},
        {"Empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)

            if tt.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Test Suite Pattern

```go
type MyTestSuite struct {
    suite.Suite
    service *MyService
}

func (s *MyTestSuite) SetupTest() {
    s.service = NewMyService()
}

func (s *MyTestSuite) TestMethod() {
    result := s.service.Method()
    s.Assert().NotNil(result)
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

### Benchmark Test

```go
func BenchmarkFunction(b *testing.B) {
    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        Function(input)
    }
}
```

## Running Tests

### Basic

```bash
# All tests
go test ./...

# Specific package
go test ./internal/applications/auth/service/...

# Specific test
go test -run TestLogin ./internal/applications/auth/service/...

# Verbose
go test -v ./...
```

### Coverage

```bash
# Generate coverage
make test-coverage

# View in browser
open coverage.html

# Coverage for specific package
go test -cover ./internal/applications/auth/service/...
```

### Advanced

```bash
# Skip long-running tests
go test -short ./...

# Race detector
make test-race

# Benchmarks
make test-bench

# Profile CPU
make test-profile
```

## Troubleshooting

### Mocks not generating?

```bash
# Clean and regenerate
make test-clean-mocks
make test-generate-mocks-all
```

### Tests failing randomly?

```bash
# Run with race detector
make test-race
```

### Database tests failing?

```bash
# Check if MySQL is running
mysql -u root -p

# Verify connection in test
mysql := testutil.SetupMySQLContainer(t)
```

### Import cycle errors?

- Move shared interfaces to separate package
- Use dependency injection
- Check circular dependencies

## Next Steps

1. Read the full [TESTING.md](TESTING.md) guide
2. Check example tests:
   - Service: `internal/applications/auth/service/auth_service_test.go`
   - Controller: `internal/applications/auth/controller/auth_controller_test.go`
   - Middleware: `internal/middlewares/version_middleware_test.go`
   - Repository: `internal/applications/user/repository/user_repository_integration_test.go`

3. Write tests for your domain:
   - Start with service layer (easiest)
   - Add repository integration tests
   - Add controller tests
   - Add middleware tests

4. Aim for coverage goals:
   ```bash
   make test-coverage
   ```

## Quick Reference

### Assertions (testify)

```go
require.NoError(t, err)           // Stop on error
require.Equal(t, expected, actual) // Stop if not equal
assert.NoError(t, err)            // Continue on error
assert.Equal(t, expected, actual)  // Continue if not equal
assert.True(t, condition)
assert.False(t, condition)
assert.Nil(t, obj)
assert.NotNil(t, obj)
assert.Contains(t, str, substr)
assert.Len(t, slice, length)
```

### Mock Expectations (testify/mock)

```go
// Basic
mock.On("Method", arg1, arg2).Return(result, nil)

// With matchers
mock.On("Method", mock.Anything).Return(result, nil)
mock.On("Method", mock.AnythingOfType("string")).Return(result, nil)

// Multiple calls
mock.On("Method", arg).Return(result1, nil).Once()
mock.On("Method", arg).Return(result2, nil).Once()

// Verify
mock.AssertExpectations(t)
mock.AssertCalled(t, "Method", arg1, arg2)
mock.AssertNumberOfCalls(t, "Method", 3)
```

### Test Context

```go
// With timeout
ctx := testutil.NewTestContext(t)

// With cancel
ctx, cancel := testutil.NewTestContextWithCancel(t)
defer cancel()

// Background
ctx := context.Background()
```

---

**Ready to test!** Start with `make test-coverage` to see current coverage, then pick an untested component and write your first test.
