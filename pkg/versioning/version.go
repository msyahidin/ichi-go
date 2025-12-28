package versioning

import (
	"fmt"
	"github.com/labstack/echo/v4"
)

// APIVersion represents an API version
type APIVersion string

// Default semantic versions
const (
	V1 APIVersion = "v1"
	V2 APIVersion = "v2"
	V3 APIVersion = "v3"

	TwentySixJan APIVersion = "202601"
)

// Global strategy - can be changed at runtime
var currentStrategy VersionStrategy = &SemanticVersionStrategy{}

// SetVersionStrategy sets the global version strategy
func SetVersionStrategy(strategy VersionStrategy) {
	currentStrategy = strategy
}

// GetVersionStrategy returns the current version strategy
func GetVersionStrategy() VersionStrategy {
	return currentStrategy
}

// String returns the string representation of the version
func (v APIVersion) String() string {
	return string(v)
}

// IsValid checks if the version is valid according to current strategy
func (v APIVersion) IsValid() bool {
	return currentStrategy.Validate(string(v))
}

// IsValidWithStrategy checks if version is valid for a specific strategy
func (v APIVersion) IsValidWithStrategy(strategy VersionStrategy) bool {
	return strategy.Validate(string(v))
}

// VersionedRoute represents a versioned API route configuration
type VersionedRoute struct {
	Version     APIVersion
	Domain      string
	ServiceName string
}

// NewVersionedRoute creates a new versioned route configuration
func NewVersionedRoute(serviceName string, version APIVersion, domain string) *VersionedRoute {
	return &VersionedRoute{
		Version:     version,
		Domain:      domain,
		ServiceName: serviceName,
	}
}

// BuildPath creates consistent versioned API paths
// Format: /serviceName/api/version/domain
// Example: /ichi-go/api/v1/auth
func (vr *VersionedRoute) BuildPath() string {
	return fmt.Sprintf("/%s/api/%s/%s", vr.ServiceName, vr.Version, vr.Domain)
}

// BuildPathWithEndpoint creates versioned API path with endpoint
// Format: /serviceName/api/version/domain/endpoint
// Example: /ichi-go/api/v1/auth/login
func (vr *VersionedRoute) BuildPathWithEndpoint(endpoint string) string {
	return fmt.Sprintf("%s/%s", vr.BuildPath(), endpoint)
}

// Group creates a versioned Echo group
func (vr *VersionedRoute) Group(e *echo.Echo) *echo.Group {
	return e.Group(vr.BuildPath())
}

// GroupWithMiddleware creates a versioned Echo group with middleware
func (vr *VersionedRoute) GroupWithMiddleware(e *echo.Echo, middleware ...echo.MiddlewareFunc) *echo.Group {
	group := e.Group(vr.BuildPath())
	group.Use(middleware...)
	return group
}

// BuildAPIPath is a helper function for backward compatibility
// Use NewVersionedRoute for new code
func BuildAPIPath(serviceName, version, domain string) string {
	return fmt.Sprintf("/%s/api/%s/%s", serviceName, version, domain)
}

// ParseVersion parses a string into an APIVersion using current strategy
func ParseVersion(version string) (APIVersion, error) {
	return currentStrategy.Parse(version)
}

// ParseVersionWithStrategy parses a string using a specific strategy
func ParseVersionWithStrategy(version string, strategy VersionStrategy) (APIVersion, error) {
	return strategy.Parse(version)
}

// AllVersions returns all supported API versions (for semantic versioning)
// For date-based versioning, this should be overridden in config
func AllVersions() []APIVersion {
	// This is only meaningful for semantic versioning
	return []APIVersion{V1, V2, V3}
}

// GetLatestVersion returns the latest API version
// For date-based versioning, returns current date
func GetLatestVersion() APIVersion {
	if currentStrategy.Name() == string(StrategyTypeDate) {
		return GetCurrentDateVersion()
	}
	if currentStrategy.Name() == string(StrategyTypeDateDaily) {
		return GetCurrentDateDailyVersion()
	}
	return V1 // Default for semantic
}
