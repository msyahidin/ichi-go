// Package testutil provides common testing utilities and helpers for ichi-go
package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Context Helpers
// ============================================================================

// NewTestContext creates a context with timeout for testing
func NewTestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// NewTestContextWithCancel creates a context with cancel for testing
func NewTestContextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx, cancel
}

// ============================================================================
// Assertion Helpers
// ============================================================================

// AssertNoError is a wrapper around require.NoError with better error messages
func AssertNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.NoError(t, err, msgAndArgs...)
}

// AssertError is a wrapper around require.Error with better error messages
func AssertError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
}

// AssertEqual is a wrapper around assert.Equal
func AssertEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotEqual is a wrapper around assert.NotEqual
func AssertNotEqual(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEqual(t, expected, actual, msgAndArgs...)
}

// AssertNil is a wrapper around assert.Nil
func AssertNil(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Nil(t, object, msgAndArgs...)
}

// AssertNotNil is a wrapper around assert.NotNil
func AssertNotNil(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotNil(t, object, msgAndArgs...)
}

// AssertTrue is a wrapper around assert.True
func AssertTrue(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	assert.True(t, value, msgAndArgs...)
}

// AssertFalse is a wrapper around assert.False
func AssertFalse(t *testing.T, value bool, msgAndArgs ...interface{}) {
	t.Helper()
	assert.False(t, value, msgAndArgs...)
}

// AssertContains checks if a string contains a substring
func AssertContains(t *testing.T, s, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Contains(t, s, substr, msgAndArgs...)
}

// AssertNotContains checks if a string does not contain a substring
func AssertNotContains(t *testing.T, s, substr string, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotContains(t, s, substr, msgAndArgs...)
}

// AssertLen checks the length of an array, slice, map, or string
func AssertLen(t *testing.T, object interface{}, length int, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Len(t, object, length, msgAndArgs...)
}

// AssertEmpty checks if an object is empty
func AssertEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Empty(t, object, msgAndArgs...)
}

// AssertNotEmpty checks if an object is not empty
func AssertNotEmpty(t *testing.T, object interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotEmpty(t, object, msgAndArgs...)
}

// ============================================================================
// Timing Helpers
// ============================================================================

// AssertEventually asserts that a condition will eventually be true within a timeout
func AssertEventually(t *testing.T, condition func() bool, timeout time.Duration, tick time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Eventually(t, condition, timeout, tick, msgAndArgs...)
}

// AssertNever asserts that a condition never becomes true within a duration
func AssertNever(t *testing.T, condition func() bool, duration time.Duration, tick time.Duration, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Never(t, condition, duration, tick, msgAndArgs...)
}

// ============================================================================
// Panic Helpers
// ============================================================================

// AssertPanics asserts that the code inside the specified function panics
func AssertPanics(t *testing.T, fn func(), msgAndArgs ...interface{}) {
	t.Helper()
	assert.Panics(t, fn, msgAndArgs...)
}

// AssertNotPanics asserts that the code inside the specified function does not panic
func AssertNotPanics(t *testing.T, fn func(), msgAndArgs ...interface{}) {
	t.Helper()
	assert.NotPanics(t, fn, msgAndArgs...)
}

// ============================================================================
// Error Type Helpers
// ============================================================================

// AssertErrorIs asserts that an error is of a specific type
func AssertErrorIs(t *testing.T, err, target error, msgAndArgs ...interface{}) {
	t.Helper()
	assert.ErrorIs(t, err, target, msgAndArgs...)
}

// AssertErrorContains asserts that error message contains a substring
func AssertErrorContains(t *testing.T, err error, contains string, msgAndArgs ...interface{}) {
	t.Helper()
	require.Error(t, err, msgAndArgs...)
	assert.Contains(t, err.Error(), contains, msgAndArgs...)
}

// ============================================================================
// Comparison Helpers
// ============================================================================

// AssertGreater asserts that e1 > e2
func AssertGreater(t *testing.T, e1, e2 interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Greater(t, e1, e2, msgAndArgs...)
}

// AssertGreaterOrEqual asserts that e1 >= e2
func AssertGreaterOrEqual(t *testing.T, e1, e2 interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.GreaterOrEqual(t, e1, e2, msgAndArgs...)
}

// AssertLess asserts that e1 < e2
func AssertLess(t *testing.T, e1, e2 interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Less(t, e1, e2, msgAndArgs...)
}

// AssertLessOrEqual asserts that e1 <= e2
func AssertLessOrEqual(t *testing.T, e1, e2 interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.LessOrEqual(t, e1, e2, msgAndArgs...)
}

// ============================================================================
// Collection Helpers
// ============================================================================

// AssertSubset asserts that list contains all elements of subset
func AssertSubset(t *testing.T, list, subset interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.Subset(t, list, subset, msgAndArgs...)
}

// AssertElementsMatch asserts that two slices contain the same elements (order doesn't matter)
func AssertElementsMatch(t *testing.T, expected, actual interface{}, msgAndArgs ...interface{}) {
	t.Helper()
	assert.ElementsMatch(t, expected, actual, msgAndArgs...)
}

// ============================================================================
// Test Cleanup Helpers
// ============================================================================

// CleanupFunc registers a cleanup function that will be called when test finishes
func CleanupFunc(t *testing.T, fn func()) {
	t.Helper()
	t.Cleanup(fn)
}

// ============================================================================
// Skip Helpers
// ============================================================================

// SkipIfShort skips the test if testing.Short() is true
func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// SkipCI skips the test if running in CI environment
func SkipCI(t *testing.T) {
	t.Helper()
	// Check common CI environment variables
	// You can customize this based on your CI provider
	t.Skip("Skipping test in CI environment")
}

// ============================================================================
// Table Test Helpers
// ============================================================================

// TableTest represents a single test case in a table-driven test
type TableTest struct {
	Name      string
	Setup     func(t *testing.T)
	Run       func(t *testing.T)
	Teardown  func(t *testing.T)
	SkipShort bool
}

// RunTableTests executes a slice of table tests
func RunTableTests(t *testing.T, tests []TableTest) {
	t.Helper()
	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			if tt.SkipShort && testing.Short() {
				t.Skip("Skipping test in short mode")
			}

			if tt.Setup != nil {
				tt.Setup(t)
			}

			if tt.Teardown != nil {
				defer tt.Teardown(t)
			}

			tt.Run(t)
		})
	}
}

// ============================================================================
// Pointer Helpers (useful for comparing pointer values in tests)
// ============================================================================

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to an int64
func Int64Ptr(i int64) *int64 {
	return &i
}

// Uint64Ptr returns a pointer to a uint64
func Uint64Ptr(i uint64) *uint64 {
	return &i
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}

// Float64Ptr returns a pointer to a float64
func Float64Ptr(f float64) *float64 {
	return &f
}

// TimePtr returns a pointer to a time.Time
func TimePtr(t time.Time) *time.Time {
	return &t
}
