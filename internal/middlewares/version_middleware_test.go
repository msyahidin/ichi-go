package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ichi-go/pkg/testutil"
	"ichi-go/pkg/versioning"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Test Setup
// ============================================================================

func setupVersionMiddlewareTest(t *testing.T) (*echo.Echo, *versioning.Config) {
	t.Helper()

	e := echo.New()

	// Create test configuration
	config := &versioning.Config{
		Enabled:           true,
		Strategy:          "date",
		DefaultVersion:    "202601",
		SupportedVersions: []string{"202511", "202512", "202601"},
		Deprecation: versioning.DeprecationConfig{
			HeaderEnabled: true,
		},
	}

	return e, config
}

// ============================================================================
// Unit Tests
// ============================================================================

func TestVersionMiddleware_DetectVersionFromHeader(t *testing.T) {
	tests := []struct {
		name            string
		headerName      string
		headerValue     string
		expectedVersion string
		expectError     bool
	}{
		{
			name:            "Valid version from API-Version header",
			headerName:      "API-Version",
			headerValue:     "202601",
			expectedVersion: "202601",
			expectError:     false,
		},
		{
			name:            "Valid version from X-API-Version header",
			headerName:      "X-API-Version",
			headerValue:     "202512",
			expectedVersion: "202512",
			expectError:     false,
		},
		{
			name:            "No version header - should use default",
			headerName:      "",
			headerValue:     "",
			expectedVersion: "202601",
			expectError:     false,
		},
		{
			name:            "Invalid version format",
			headerName:      "API-Version",
			headerValue:     "invalid",
			expectedVersion: "202601",
			expectError:     false,
		},
		{
			name:            "Unsupported version - should use default",
			headerName:      "API-Version",
			headerValue:     "999999",
			expectedVersion: "202601",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			e, config := setupVersionMiddlewareTest(t)
			middleware := VersionMiddleware(config)

			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			if tt.headerName != "" {
				req.Header.Set(tt.headerName, tt.headerValue)
			}

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handlerCalled := false
			handler := func(c echo.Context) error {
				handlerCalled = true

				// Check version was set in context
				version := c.Get("api_version")
				if version != nil {
					testutil.AssertEqual(t, tt.expectedVersion, version)
				}

				return c.String(http.StatusOK, "OK")
			}

			// Act
			err := middleware(handler)(c)

			// Assert
			testutil.AssertNoError(t, err)
			assert.True(t, handlerCalled, "Handler should be called")
		})
	}
}

func TestVersionMiddleware_VersionFromPath(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		expectedVersion string
		shouldExtract   bool
	}{
		{
			name:            "Version in path - v1",
			path:            "/api/v1/users",
			expectedVersion: "202601",
			shouldExtract:   true,
		},
		{
			name:            "Version in path - date format",
			path:            "/api/202512/users",
			expectedVersion: "202512",
			shouldExtract:   true,
		},
		{
			name:            "No version in path",
			path:            "/api/users",
			expectedVersion: "202601",
			shouldExtract:   false,
		},
		{
			name:            "Root path",
			path:            "/",
			expectedVersion: "202601",
			shouldExtract:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			e, config := setupVersionMiddlewareTest(t)
			middleware := VersionMiddleware(config)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			// Act
			err := middleware(handler)(c)

			// Assert
			testutil.AssertNoError(t, err)
		})
	}
}

func TestVersionMiddleware_Disabled(t *testing.T) {
	// Arrange
	e := echo.New()
	config := &versioning.Config{
		Enabled: false,
	}

	middleware := VersionMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("API-Version", "202601")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true

		// Version should not be set when middleware is disabled
		version := c.Get("api_version")
		testutil.AssertNil(t, version)

		return c.String(http.StatusOK, "OK")
	}

	// Act
	err := middleware(handler)(c)

	// Assert
	testutil.AssertNoError(t, err)
	assert.True(t, handlerCalled)
}

func TestVersionMiddleware_DeprecationWarning(t *testing.T) {
	tests := []struct {
		name                   string
		version                string
		expectDeprecatedHeader bool
		expectSunsetHeader     bool
	}{
		{
			name:                   "Current version - no deprecation",
			version:                "202601",
			expectDeprecatedHeader: false,
			expectSunsetHeader:     false,
		},
		{
			name:                   "Old version - should show deprecation",
			version:                "202511",
			expectDeprecatedHeader: true,
			expectSunsetHeader:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			e, config := setupVersionMiddlewareTest(t)
			middleware := VersionMiddleware(config)

			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			req.Header.Set("API-Version", tt.version)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			// Act
			err := middleware(handler)(c)

			// Assert
			testutil.AssertNoError(t, err)

			deprecatedHeader := rec.Header().Get("Deprecated")
			sunsetHeader := rec.Header().Get("Sunset")

			if tt.expectDeprecatedHeader {
				testutil.AssertNotEmpty(t, deprecatedHeader)
			} else {
				testutil.AssertEmpty(t, deprecatedHeader)
			}

			if tt.expectSunsetHeader {
				testutil.AssertNotEmpty(t, sunsetHeader)
			} else {
				testutil.AssertEmpty(t, sunsetHeader)
			}
		})
	}
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkVersionMiddleware(b *testing.B) {
	e := echo.New()
	config := &versioning.Config{
		Enabled:           true,
		Strategy:          "date",
		DefaultVersion:    "202601",
		SupportedVersions: []string{"202511", "202512", "202601"},
	}

	middleware := VersionMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("API-Version", "202601")

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		middleware(handler)(c)
	}
}

func BenchmarkVersionMiddleware_WithDeprecation(b *testing.B) {
	e := echo.New()
	config := &versioning.Config{
		Enabled:           true,
		Strategy:          "date",
		DefaultVersion:    "202601",
		SupportedVersions: []string{"202511", "202512", "202601"},
	}

	middleware := VersionMiddleware(config)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("API-Version", "202511") // Old version

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		middleware(handler)(c)
	}
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestVersionMiddleware_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Arrange
	e := echo.New()
	config := &versioning.Config{
		Enabled:           true,
		Strategy:          "date",
		DefaultVersion:    "202601",
		SupportedVersions: []string{"202511", "202512", "202601"},
	}

	e.Use(VersionMiddleware(config))

	// Register test route
	e.GET("/api/users", func(c echo.Context) error {
		version := c.Get("api_version")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"version": version,
			"data":    []string{"user1", "user2"},
		})
	})

	tests := []struct {
		name           string
		apiVersion     string
		expectedStatus int
	}{
		{
			name:           "Current version",
			apiVersion:     "202601",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Supported old version",
			apiVersion:     "202512",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "No version header",
			apiVersion:     "",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
			if tt.apiVersion != "" {
				req.Header.Set("API-Version", tt.apiVersion)
			}
			rec := httptest.NewRecorder()

			// Execute
			e.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// ============================================================================
// Table-Driven Test Example
// ============================================================================

func TestVersionMiddleware_TableDriven(t *testing.T) {
	testCases := []testutil.TableTest{
		{
			Name: "Valid version detection",
			Run: func(t *testing.T) {
				e, config := setupVersionMiddlewareTest(t)
				middleware := VersionMiddleware(config)

				req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
				req.Header.Set("API-Version", "202601")
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				handler := func(c echo.Context) error {
					return c.String(http.StatusOK, "OK")
				}

				err := middleware(handler)(c)
				require.NoError(t, err)
			},
		},
		{
			Name: "Fallback to default version",
			Run: func(t *testing.T) {
				e, config := setupVersionMiddlewareTest(t)
				middleware := VersionMiddleware(config)

				req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				handler := func(c echo.Context) error {
					return c.String(http.StatusOK, "OK")
				}

				err := middleware(handler)(c)
				require.NoError(t, err)
			},
		},
	}

	testutil.RunTableTests(t, testCases)
}
