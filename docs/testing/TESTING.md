# Testing Best Practices for ichi-go

This document outlines testing standards, patterns, and best practices for the ichi-go project.

## Table of Contents

- [Overview](#overview)
- [Testing Stack](#testing-stack)
- [Test Organization](#test-organization)
- [Testing Patterns](#testing-patterns)
- [Mock Generation](#mock-generation)
- [Test Utilities](#test-utilities)
- [Integration Testing](#integration-testing)
- [Best Practices](#best-practices)
- [Common Patterns](#common-patterns)
- [Examples](#examples)

## Overview

The ichi-go project follows a comprehensive testing strategy:

- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete user flows
- **Benchmark Tests**: Measure performance

### Testing Goals

- Minimum 80% code coverage for services
- 70% coverage for repositories
- 70% coverage for controllers
- 90% coverage for critical infrastructure (auth, cache, queue)

## Testing Stack

### Core Libraries

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
)
```

### Tools

- **testify**: Assertion and mocking framework
- **mockery**: Automatic mock generation
- **go test**: Native Go testing framework
- **testcontainers** (optional): Docker containers for integration tests

## Test Organization

### Directory Structure

```
internal/
├── applications/
│   └── auth/
│       ├── service/
│       │   ├── auth_service.go
│       │   └── auth_service_test.go      # Unit tests alongside code
│       ├── repository/
│       │   ├── user_repository.go
│       │   ├── user_repository_test.go
│       │   └── mocks/
│       │       └── mock_user_repository.go
│       └── controllers/
│           ├── auth_controller.go
│           └── auth_controller_test.go
└── testutil/                              # Shared test utilities
    ├── helpers.go                         # Test helper functions
    ├── fixtures.go                        # Test data builders
    └── containers.go                      # Test container setup

pkg/
└── authenticator/
    ├── jwt.go
    ├── jwt_test.go
    └── mocks/
        └── mock_jwt_service.go
```

### Naming Conventions

- **Test files**: `*_test.go`
- **Test functions**: `Test<FunctionName>`
- **Benchmark functions**: `Benchmark<FunctionName>`
- **Mock files**: `mock_<interface_name>.go`
- **Test helpers**: Place in `internal/testutil`

## Testing Patterns

### 1. Table-Driven Tests

**Best for**: Testing multiple scenarios with different inputs

```go
func TestUserService_Create(t *testing.T) {
    tests := []struct {
        name        string
        input       *CreateUserDTO
        setupMock   func(*mocks.MockUserRepository)
        expectError bool
        errorMsg    string
    }{
        {
            name: "Success - Valid user creation",
            input: &CreateUserDTO{
                Email:    "user@example.com",
                Username: "testuser",
                Password: "Password123!",
            },
            setupMock: func(m *mocks.MockUserRepository) {
                m.EXPECT().
                    Create(mock.Anything, mock.AnythingOfType("UserModel")).
                    Return(int64(1), nil)
            },
            expectError: false,
        },
        {
            name: "Error - Duplicate email",
            input: &CreateUserDTO{
                Email:    "existing@example.com",
                Username: "testuser",
                Password: "Password123!",
            },
            setupMock: func(m *mocks.MockUserRepository) {
                m.EXPECT().
                    Create(mock.Anything, mock.AnythingOfType("UserModel")).
                    Return(int64(0), errors.New("duplicate key"))
            },
            expectError: true,
            errorMsg:    "duplicate key",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Arrange
            mockRepo := mocks.NewMockUserRepository(t)
            tt.setupMock(mockRepo)
            service := NewUserService(mockRepo)
            ctx := context.Background()

            // Act
            result, err := service.Create(ctx, tt.input)

            // Assert
            if tt.expectError {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                require.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

### 2. Test Suites (for complex setup/teardown)

**Best for**: Tests requiring shared setup/teardown logic

```go
type AuthServiceTestSuite struct {
    suite.Suite
    mockRepo    *mocks.MockUserRepository
    mockJWT     *mocks.MockJWTService
    service     *AuthService
    ctx         context.Context
}

func (suite *AuthServiceTestSuite) SetupTest() {
    suite.mockRepo = mocks.NewMockUserRepository(suite.T())
    suite.mockJWT = mocks.NewMockJWTService(suite.T())
    suite.service = NewAuthService(suite.mockRepo, suite.mockJWT)
    suite.ctx = context.Background()
}

func (suite *AuthServiceTestSuite) TearDownTest() {
    // Cleanup if needed
}

func (suite *AuthServiceTestSuite) TestLogin_Success() {
    // Test implementation
}

func TestAuthServiceTestSuite(t *testing.T) {
    suite.Run(t, new(AuthServiceTestSuite))
}
```

### 3. Parallel Tests

**Best for**: Independent tests that can run concurrently

```go
func TestUserService_ParallelTests(t *testing.T) {
    tests := []struct {
        name string
        fn   func(*testing.T)
    }{
        {"Test1", testCase1},
        {"Test2", testCase2},
        {"Test3", testCase3},
    }

    for _, tt := range tests {
        tt := tt // capture range variable
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel() // Run in parallel
            tt.fn(t)
        })
    }
}
```

### 4. Subtests

**Best for**: Grouping related test scenarios

```go
func TestUserValidation(t *testing.T) {
    t.Run("Email validation", func(t *testing.T) {
        t.Run("Valid email", func(t *testing.T) {
            // Test valid email
        })
        t.Run("Invalid email", func(t *testing.T) {
            // Test invalid email
        })
    })

    t.Run("Password validation", func(t *testing.T) {
        t.Run("Strong password", func(t *testing.T) {
            // Test strong password
        })
        t.Run("Weak password", func(t *testing.T) {
            // Test weak password
        })
    })
}
```

## Mock Generation

### Using Mockery

#### Configuration

The `.mockery.yaml` configuration uses relative paths (starting with `./`) to make it portable across different project names:

```yaml
# Global settings
quiet: false
with-expecter: true
all: true
inpackage: false
keeptree: true

# Dynamic directory - mocks go in a "mocks" subdirectory next to the interface
dir: "{{.InterfaceDir}}/mocks"

mockname: "Mock{{.InterfaceName}}"
outpkg: "mocks"

# Use relative paths for portability
packages:
  ./internal/applications/auth/service:
    config:
      all: true

  ./internal/applications/user/repository:
    config:
      all: true
```

**Key Points:**
- Uses `./internal/...` instead of `module-name/internal/...`
- Works with any project name (ichi-go, notification-service, cart-service, etc.)
- `{{.InterfaceDir}}` template creates mocks next to source code
- `all: true` generates mocks for all exported interfaces in the package

#### Generate Mocks

```bash
# Install mockery
make test-install-mockery

# Generate all mocks (uses .mockery.yaml)
make test-generate-mocks

# Force regenerate all mocks
make test-generate-mocks-all

# Clean mock files
make test-clean-mocks
```

#### Using Generated Mocks

```go
func TestService(t *testing.T) {
    // Create mock with testify's expecter pattern
    mockRepo := mocks.NewMockUserRepository(t)

    // Set expectations
    mockRepo.EXPECT().
        FindByID(mock.Anything, uint64(1)).
        Return(&User{ID: 1, Email: "test@example.com"}, nil).
        Once()

    // Use mock
    service := NewService(mockRepo)
    result, err := service.GetUser(context.Background(), 1)

    // Assertions
    require.NoError(t, err)
    assert.Equal(t, uint64(1), result.ID)

    // Verify all expectations were met (automatic with testify's expecter)
}
```

### Manual Mocks (when needed)

```go
type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint64) (*User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*User), args.Error(1)
}

// Usage
mockRepo := new(MockUserRepository)
mockRepo.On("FindByID", mock.Anything, uint64(1)).
    Return(&User{ID: 1}, nil)
```

## Test Utilities

### Helpers (`internal/testutil/helpers.go`)

```go
import "ichi-go/internal/testutil"

func TestExample(t *testing.T) {
    // Context with timeout
    ctx := testutil.NewTestContext(t)

    // Assertions
    testutil.AssertNoError(t, err)
    testutil.AssertEqual(t, expected, actual)

    // Pointer helpers
    email := testutil.StringPtr("test@example.com")
    id := testutil.Uint64Ptr(123)
}
```

### Fixtures (`internal/testutil/fixtures.go`)

```go
import "ichi-go/internal/testutil"

func TestWithFixtures(t *testing.T) {
    // Create user fixture
    user := testutil.NewUserFixture().
        WithEmail("custom@example.com").
        WithPassword("CustomPass123!")

    // Create token fixture
    token := testutil.NewTokenFixture().
        WithUserID(user.ID)

    // Use in tests
    assert.Equal(t, "custom@example.com", user.Email)
}
```

### Test Containers (`internal/testutil/containers.go`)

```go
import "ichi-go/internal/testutil"

func TestIntegration(t *testing.T) {
    // Setup MySQL container
    mysql := testutil.SetupMySQLContainer(t)
    db := mysql.GetDB(t)

    // Setup Redis container
    redis := testutil.SetupRedisContainer(t)

    // Use in integration tests
    // Cleanup is automatic via t.Cleanup()
}
```

## Integration Testing

### Database Integration Tests

```go
func TestUserRepository_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Setup test database
    mysql := testutil.SetupMySQLContainer(t)
    db := mysql.GetDB(t)

    // Run migrations
    // ... migration logic

    // Create repository
    repo := NewUserRepository(db)
    ctx := testutil.NewTestContext(t)

    // Test Create
    t.Run("Create user", func(t *testing.T) {
        user := &User{
            Email:    "test@example.com",
            Username: "testuser",
        }

        id, err := repo.Create(ctx, user)
        require.NoError(t, err)
        assert.Greater(t, id, int64(0))
    })

    // Test FindByID
    t.Run("Find by ID", func(t *testing.T) {
        user, err := repo.FindByID(ctx, 1)
        require.NoError(t, err)
        assert.Equal(t, "test@example.com", user.Email)
    })
}
```

### Controller Integration Tests

```go
func TestAuthController_Integration(t *testing.T) {
    // Setup Echo
    e := echo.New()

    // Setup dependencies with mocks
    mockService := mocks.NewMockAuthService(t)
    controller := NewAuthController(mockService)

    // Setup expectations
    mockService.EXPECT().
        Login(mock.Anything, "test@example.com", "password").
        Return(&TokenPair{AccessToken: "token"}, nil)

    // Create request
    reqBody := `{"email":"test@example.com","password":"password"}`
    req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(reqBody))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    // Execute
    err := controller.Login(c)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)

    // Parse response
    var resp map[string]interface{}
    json.Unmarshal(rec.Body.Bytes(), &resp)
    assert.Equal(t, "token", resp["data"].(map[string]interface{})["access_token"])
}
```

## Best Practices

### 1. Arrange-Act-Assert (AAA) Pattern

```go
func TestService_Method(t *testing.T) {
    // Arrange
    mockRepo := mocks.NewMockRepository(t)
    mockRepo.EXPECT().FindByID(mock.Anything, uint64(1)).Return(&Model{}, nil)
    service := NewService(mockRepo)
    ctx := context.Background()

    // Act
    result, err := service.GetByID(ctx, 1)

    // Assert
    require.NoError(t, err)
    assert.NotNil(t, result)
}
```

### 2. Use require vs assert

- **require**: Stops test immediately on failure (for critical checks)
- **assert**: Continues test on failure (for multiple validations)

```go
func TestExample(t *testing.T) {
    result, err := someFunction()

    require.NoError(t, err)       // Stop if error
    require.NotNil(t, result)     // Stop if nil

    assert.Equal(t, "expected", result.Field1)  // Continue on failure
    assert.Greater(t, result.Value, 0)          // Continue on failure
}
```

### 3. Test Error Cases

Always test both success and error scenarios:

```go
func TestService_Create(t *testing.T) {
    t.Run("Success", func(t *testing.T) {
        // Test successful creation
    })

    t.Run("Error - Invalid input", func(t *testing.T) {
        // Test validation errors
    })

    t.Run("Error - Database failure", func(t *testing.T) {
        // Test infrastructure errors
    })

    t.Run("Error - Duplicate entry", func(t *testing.T) {
        // Test business logic errors
    })
}
```

### 4. Use Test Fixtures

Create reusable test data:

```go
func createTestUser(t *testing.T) *User {
    t.Helper()
    return &User{
        ID:       1,
        Email:    "test@example.com",
        Username: "testuser",
    }
}

func TestWithFixture(t *testing.T) {
    user := createTestUser(t)
    // Use user in test
}
```

### 5. Test Concurrency

```go
func TestConcurrentAccess(t *testing.T) {
    service := NewService()

    const workers = 100
    done := make(chan bool, workers)
    errors := make(chan error, workers)

    for i := 0; i < workers; i++ {
        go func(id int) {
            err := service.DoSomething(id)
            if err != nil {
                errors <- err
            }
            done <- true
        }(i)
    }

    // Wait for completion
    for i := 0; i < workers; i++ {
        <-done
    }

    close(errors)
    assert.Len(t, errors, 0, "No errors should occur")
}
```

### 6. Use Context

Always pass context to functions that support it:

```go
func TestWithContext(t *testing.T) {
    ctx := testutil.NewTestContext(t) // With timeout

    result, err := service.Method(ctx)
    require.NoError(t, err)
}

func TestContextCancellation(t *testing.T) {
    ctx, cancel := testutil.NewTestContextWithCancel(t)

    // Start operation
    go func() {
        time.Sleep(100 * time.Millisecond)
        cancel() // Cancel context
    }()

    // Should return error due to cancellation
    _, err := service.LongRunningOperation(ctx)
    assert.Error(t, err)
    assert.ErrorIs(t, err, context.Canceled)
}
```

### 7. Clean Test Database

```go
func TestRepository_Integration(t *testing.T) {
    db := setupTestDB(t)

    // Cleanup after test
    t.Cleanup(func() {
        testutil.TruncateTables(t, db)
    })

    // Run tests
}
```

### 8. Benchmark Critical Paths

```go
func BenchmarkService_CriticalMethod(b *testing.B) {
    service := NewService()
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        service.CriticalMethod(ctx, input)
    }
}

// Run with: go test -bench=. -benchmem
```

## Common Patterns

### Testing Middleware

```go
func TestAuthMiddleware(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodGet, "/", nil)
    req.Header.Set("Authorization", "Bearer valid_token")
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    handler := func(c echo.Context) error {
        return c.String(http.StatusOK, "OK")
    }

    middleware := AuthMiddleware(mockJWTService)
    err := middleware(handler)(c)

    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, rec.Code)
}
```

### Testing Validators

```go
func TestUserValidator(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserDTO
        wantErr bool
        errTag  string
    }{
        {
            name: "Valid user",
            input: CreateUserDTO{
                Email:    "test@example.com",
                Password: "Password123!",
            },
            wantErr: false,
        },
        {
            name: "Invalid email",
            input: CreateUserDTO{
                Email:    "invalid",
                Password: "Password123!",
            },
            wantErr: true,
            errTag:  "email",
        },
    }

    validator := validator.New()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validator.Struct(tt.input)
            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Testing with Transactions

```go
func TestWithTransaction(t *testing.T) {
    db := setupTestDB(t)

    testutil.WithTransaction(t, db, func(tx *sql.Tx) error {
        // All database operations use tx
        // Will be rolled back automatically

        repo := NewRepository(tx)
        _, err := repo.Create(ctx, &Model{})
        return err
    })

    // Changes are rolled back
}
```

## Running Tests

### Basic Commands

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/applications/auth/service/...

# Run specific test
go test -run TestAuthService_Login ./internal/applications/auth/service/...

# Run with race detector
make test-race

# Run benchmarks
make test-bench
```

### CI/CD

```bash
# CI-optimized test run
make test-ci

# Generate coverage for CI
make test-ci-short
```

### Advanced

```bash
# Run only short tests
go test -short ./...

# Verbose output
make test-verbose

# Fail fast
make test-failfast

# With timeout
make test-timeout

# Generate test profile
make test-profile
```

## Examples

See existing tests for reference:
- `internal/applications/auth/service/auth_service_test.go` - Service layer testing
- `pkg/authenticator/jwt_test.go` - Package testing with benchmarks
- `pkg/validator/validator_test.go` - Validation testing

## Coverage Goals

| Layer | Target Coverage |
|-------|----------------|
| Services | 80%+ |
| Repositories | 70%+ |
| Controllers | 70%+ |
| Middlewares | 80%+ |
| Core Packages (auth, cache, queue) | 90%+ |
| Utilities | 70%+ |

## Additional Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Mockery Documentation](https://vektra.github.io/mockery/)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

**Remember**: Good tests are:
- **Fast**: Run quickly
- **Independent**: Don't depend on other tests
- **Repeatable**: Same result every time
- **Self-validating**: Pass or fail clearly
- **Timely**: Written alongside code
