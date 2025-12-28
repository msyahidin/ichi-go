package versioning

import (
	"fmt"
	"regexp"
	"time"
)

// VersionStrategy defines how versions are formatted and validated
type VersionStrategy interface {
	// Format returns the version string format
	Format(version APIVersion) string

	// Validate checks if a version string is valid for this strategy
	Validate(version string) bool

	// Parse converts a string to APIVersion
	Parse(version string) (APIVersion, error)

	// Name returns the strategy name
	Name() string
}

// VersionStrategyType represents different versioning strategies
type VersionStrategyType string

const (
	StrategyTypeSemantic  VersionStrategyType = "semantic"   // v1, v2, v3
	StrategyTypeDate      VersionStrategyType = "date"       // 2026-01, 2026-12
	StrategyTypeDateDaily VersionStrategyType = "date_daily" // 20260115, 20261231
	StrategyTypeCustom    VersionStrategyType = "custom"     // user-defined
)

// SemanticVersionStrategy implements semantic versioning (v1, v2, v3)
type SemanticVersionStrategy struct{}

func (s *SemanticVersionStrategy) Name() string {
	return string(StrategyTypeSemantic)
}

func (s *SemanticVersionStrategy) Format(version APIVersion) string {
	return string(version)
}

func (s *SemanticVersionStrategy) Validate(version string) bool {
	matched, _ := regexp.MatchString(`^v\d+$`, version)
	return matched
}

func (s *SemanticVersionStrategy) Parse(version string) (APIVersion, error) {
	if !s.Validate(version) {
		return "", fmt.Errorf("invalid semantic version format: %s (expected: v1, v2, etc.)", version)
	}
	return APIVersion(version), nil
}

// DateVersionStrategy implements date-based versioning (2026-01, 2026-12)
type DateVersionStrategy struct{}

func (s *DateVersionStrategy) Name() string {
	return string(StrategyTypeDate)
}

func (s *DateVersionStrategy) Format(version APIVersion) string {
	return string(version)
}

func (s *DateVersionStrategy) Validate(version string) bool {
	// Validates YYYY-MM format
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}$`, version)
	if !matched {
		return false
	}

	// Parse to ensure it's a valid date
	_, err := time.Parse("2006-01", version)
	return err == nil
}

func (s *DateVersionStrategy) Parse(version string) (APIVersion, error) {
	if !s.Validate(version) {
		return "", fmt.Errorf("invalid date version format: %s (expected: YYYY-MM, e.g., 2026-01)", version)
	}
	return APIVersion(version), nil
}

// DateDailyVersionStrategy implements daily date-based versioning (20260115, 20261231)
type DateDailyVersionStrategy struct{}

func (s *DateDailyVersionStrategy) Name() string {
	return string(StrategyTypeDateDaily)
}

func (s *DateDailyVersionStrategy) Format(version APIVersion) string {
	return string(version)
}

func (s *DateDailyVersionStrategy) Validate(version string) bool {
	// Validates YYYYMMDD format
	matched, _ := regexp.MatchString(`^\d{8}$`, version)
	if !matched {
		return false
	}

	// Parse to ensure it's a valid date
	_, err := time.Parse("20060102", version)
	return err == nil
}

func (s *DateDailyVersionStrategy) Parse(version string) (APIVersion, error) {
	if !s.Validate(version) {
		return "", fmt.Errorf("invalid date version format: %s (expected: YYYYMMDD, e.g., 20260115)", version)
	}
	return APIVersion(version), nil
}

// CustomVersionStrategy allows user-defined version patterns
type CustomVersionStrategy struct {
	pattern       *regexp.Regexp
	validVersions map[string]bool
	name          string
}

// NewCustomVersionStrategy creates a custom version strategy
func NewCustomVersionStrategy(name string, pattern string, validVersions []string) (*CustomVersionStrategy, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	versionMap := make(map[string]bool)
	for _, v := range validVersions {
		versionMap[v] = true
	}

	return &CustomVersionStrategy{
		pattern:       regex,
		validVersions: versionMap,
		name:          name,
	}, nil
}

func (s *CustomVersionStrategy) Name() string {
	return s.name
}

func (s *CustomVersionStrategy) Format(version APIVersion) string {
	return string(version)
}

func (s *CustomVersionStrategy) Validate(version string) bool {
	// If specific versions are defined, check against them
	if len(s.validVersions) > 0 {
		return s.validVersions[version]
	}

	// Otherwise, validate against pattern
	return s.pattern.MatchString(version)
}

func (s *CustomVersionStrategy) Parse(version string) (APIVersion, error) {
	if !s.Validate(version) {
		return "", fmt.Errorf("invalid version format: %s", version)
	}
	return APIVersion(version), nil
}

// GetStrategy returns the appropriate strategy based on type
func GetStrategy(strategyType VersionStrategyType) VersionStrategy {
	switch strategyType {
	case StrategyTypeSemantic:
		return &SemanticVersionStrategy{}
	case StrategyTypeDate:
		return &DateVersionStrategy{}
	case StrategyTypeDateDaily:
		return &DateDailyVersionStrategy{}
	default:
		return &SemanticVersionStrategy{} // Default to semantic
	}
}

// Helper functions for common date-based versions

// NewDateVersion creates a date-based version from year and month
func NewDateVersion(year int, month int, separator string) APIVersion {
	if separator == "" {
		separator = "-"
	}
	return APIVersion(fmt.Sprintf("%04d%s%02d", year, separator, month))
}

// NewDateDailyVersion creates a daily date-based version
func NewDateDailyVersion(year int, month int, day int) APIVersion {
	return APIVersion(fmt.Sprintf("%04d%02d%02d", year, month, day))
}

// GetCurrentDateVersion returns the current month as a version
func GetCurrentDateVersion() APIVersion {
	now := time.Now()
	return NewDateVersion(now.Year(), int(now.Month()), "")
}

// GetCurrentDateDailyVersion returns the current date as a version
func GetCurrentDateDailyVersion() APIVersion {
	now := time.Now()
	return NewDateDailyVersion(now.Year(), int(now.Month()), now.Day())
}

// ParseDateVersion parses a date version and returns the time
func ParseDateVersion(version APIVersion) (time.Time, error) {
	return time.Parse("2006-01", string(version))
}

// ParseDateDailyVersion parses a daily date version and returns the time
func ParseDateDailyVersion(version APIVersion) (time.Time, error) {
	return time.Parse("20060102", string(version))
}
