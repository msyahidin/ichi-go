# Testing Cheat Sheet

Quick reference for testing in ichi-go.

## Quick Commands

```bash
# Setup
make test-install-mockery          # Install mockery
make test-generate-mocks           # Generate all mocks

# Run Tests
make test                          # All tests
make test-coverage                 # With coverage report
make test-race                     # With race detector
make test-bench                    # Benchmarks

# Quick Aliases
make qt                            # Quick test
make qtc                           # Quick coverage
make qtr                           # Quick race

# Specific Tests
make test-auth                     # Auth tests only
go test ./internal/applications/auth/service/...

# Skip Integration Tests
go test -short ./...
```

## Import Block

```go
import (
    "context"
    "testing"
    "ichi-go/internal/testutil"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
)
```

## Test Function Template

```go
func TestMyFunction_Success(t *testing.T) {
    // Arrange
    ctx := testutil.NewTestContext(t)
    mockRepo := mocks.NewMockRepository(t)
    service := NewService(mockRepo)

    mockRepo.EXPECT().
        Method(ctx, arg).
        Return(result, nil)

    // Act
    result, err := service.MyFunction(ctx, arg)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

## Table-Driven Test

```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectError bool
    }{
        {"success case", "input", "output", false},
        {"error case", "", "", true},
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

## Assertions

```go
// Require (stops on failure)
require.NoError(t, err)
require.Error(t, err)
require.Equal(t, expected, actual)
require.NotEqual(t, expected, actual)
require.Nil(t, obj)
require.NotNil(t, obj)
require.True(t, condition)
require.False(t, condition)

// Assert (continues on failure)
assert.NoError(t, err)
assert.Error(t, err)
assert.Equal(t, expected, actual)
assert.NotEqual(t, expected, actual)
assert.Contains(t, str, substr)
assert.Len(t, slice, 3)
assert.Empty(t, slice)
assert.NotEmpty(t, slice)
assert.Greater(t, val1, val2)
assert.Less(t, val1, val2)

// Test utilities
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, expected, actual)
testutil.AssertErrorContains(t, err, "message")
```

## Mock Setup

```go
// Basic mock
mockRepo.EXPECT().
    FindByID(ctx, uint64(1)).
    Return(&Model{ID: 1}, nil)

// With Any matchers
mockRepo.EXPECT().
    Create(mock.Anything, mock.AnythingOfType("Model")).
    Return(int64(1), nil)

// Multiple calls
mockRepo.EXPECT().
    Method(arg).
    Return(result1, nil).
    Once()

mockRepo.EXPECT().
    Method(arg).
    Return(result2, nil).
    Once()

// Verify expectations
mockRepo.AssertExpectations(t)
```

## Test Fixtures

```go
// User fixture
user := testutil.NewUserFixture().
    WithEmail("test@example.com").
    WithPassword("Password123!").
    WithID(123)

// Token fixture
token := testutil.NewTokenFixture().
    WithUserID(user.ID).
    WithAccessToken("custom_token")

// Pointers
email := testutil.StringPtr("test@example.com")
id := testutil.Uint64Ptr(123)
enabled := testutil.BoolPtr(true)
```

## Test Context

```go
// With timeout (30s default)
ctx := testutil.NewTestContext(t)

// With cancel
ctx, cancel := testutil.NewTestContextWithCancel(t)
defer cancel()

// Background
ctx := context.Background()
```

## Controller Test

```go
func TestController_Handler(t *testing.T) {
    // Setup
    e := echo.New()
    mockService := mocks.NewMockService(t)
    controller := NewController(mockService)

    // Mock expectation
    mockService.EXPECT().
        Method(mock.Anything, arg).
        Return(result, nil)

    // Create request
    req := httptest.NewRequest(http.MethodPost, "/path", body)
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // Execute
    err := controller.Handler(c)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

## Integration Test

```go
func TestRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    // Setup database
    mysql := testutil.SetupMySQLContainer(t)
    db := mysql.GetDB(t)
    defer db.Close()

    // Create repository
    repo := NewRepository(db)
    ctx := testutil.NewTestContext(t)

    // Test
    id, err := repo.Create(ctx, &Model{})
    require.NoError(t, err)
    assert.Greater(t, id, int64(0))

    // Cleanup is automatic via t.Cleanup()
}
```

## Benchmark Test

```go
func BenchmarkFunction(b *testing.B) {
    // Setup
    service := NewService()

    b.ResetTimer()
    b.ReportAllocs()

    for i := 0; i < b.N; i++ {
        service.Function(input)
    }
}
```

## Test Suite

```go
type MyTestSuite struct {
    suite.Suite
    service *Service
    mockRepo *mocks.MockRepository
}

func (s *MyTestSuite) SetupTest() {
    s.mockRepo = mocks.NewMockRepository(s.T())
    s.service = NewService(s.mockRepo)
}

func (s *MyTestSuite) TearDownTest() {
    // Cleanup
}

func (s *MyTestSuite) TestMethod() {
    // Test code
    s.Assert().NotNil(result)
}

func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

## Parallel Tests

```go
func TestParallel(t *testing.T) {
    tests := []struct{
        name string
        fn   func(*testing.T)
    }{
        {"Test1", test1},
        {"Test2", test2},
    }

    for _, tt := range tests {
        tt := tt // capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            tt.fn(t)
        })
    }
}
```

## Concurrency Test

```go
func TestConcurrent(t *testing.T) {
    const workers = 100
    done := make(chan bool, workers)

    for i := 0; i < workers; i++ {
        go func(id int) {
            // Test code
            done <- true
        }(i)
    }

    for i := 0; i < workers; i++ {
        <-done
    }
}
```

## Transaction Test

```go
func TestWithTransaction(t *testing.T) {
    db := setupDB(t)

    testutil.WithTransaction(t, db, func(tx *sql.Tx) error {
        // Operations using tx
        // Will be rolled back automatically
        return nil
    })
}
```

## HTTP Request Helper

```go
func createTestRequest(method, path string, body interface{}) (*http.Request, error) {
    reqBody, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }

    req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    return req, nil
}
```

## Test Data Builders

```go
// Random data
email := testutil.RandomEmail()
username := testutil.RandomUsername()
uuid := testutil.RandomUUID()
id := testutil.RandomUint64()

// Password
password := testutil.ValidPassword()        // "Password123!"
hashed := testutil.HashPassword(password)

// Time
fixedTime := testutil.FixedTime            // 2024-01-01 12:00:00 UTC
now := testutil.NowUTC()
```

## Mock Matchers

```go
mock.Anything                           // Any value
mock.AnythingOfType("string")          // Any string
mock.AnythingOfType("*Model")          // Any *Model pointer
mock.MatchedBy(func(v interface{}) bool {
    return v.(string) == "expected"
})
```

## Coverage Commands

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -func=coverage.out

# HTML report
go tool cover -html=coverage.out -o coverage.html
open coverage.html

# Or use make
make test-coverage
```

## Common Patterns

```go
// Setup helper
func setupTest(t *testing.T) (*Service, *mocks.MockRepo) {
    t.Helper()
    mockRepo := mocks.NewMockRepo(t)
    service := NewService(mockRepo)
    return service, mockRepo
}

// Cleanup helper
func cleanup(t *testing.T, db *sql.DB) {
    t.Helper()
    t.Cleanup(func() {
        testutil.TruncateTables(t, db)
    })
}

// Error assertion
testutil.AssertErrorContains(t, err, "expected message")
testutil.AssertErrorIs(t, err, ErrNotFound)

// Eventually assertion
testutil.AssertEventually(t, func() bool {
    return condition == true
}, 5*time.Second, 100*time.Millisecond)
```

## Skip Tests

```go
// Skip in short mode
if testing.Short() {
    t.Skip("Skipping in short mode")
}

// Skip in CI
testutil.SkipCI(t)

// Skip with helper
testutil.SkipIfShort(t)
```

## Test File Names

```
my_service.go           → my_service_test.go
user_repository.go      → user_repository_test.go
                          user_repository_integration_test.go
auth_controller.go      → auth_controller_test.go
```

## Coverage Goals

- Services: **80%+**
- Repositories: **70%+**
- Controllers: **70%+**
- Middlewares: **80%+**
- Core packages: **90%+**

## Quick Tips

1. Use `require` for critical checks (stops test)
2. Use `assert` for multiple validations (continues)
3. Always pass context to functions
4. Test both success and error cases
5. Use table-driven tests for multiple scenarios
6. Mock external dependencies
7. Clean up with `t.Cleanup()`
8. Use fixtures for test data
9. Skip integration tests in short mode
10. Add benchmarks for critical code

## Resources

- [TESTING.md](TESTING.md) - Full guide
- [TESTING_QUICKSTART.md](TESTING_QUICKSTART.md) - Quick start
- Example tests in `internal/applications/auth/`
