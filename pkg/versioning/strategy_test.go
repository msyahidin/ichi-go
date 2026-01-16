package versioning

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemanticVersionStrategy(t *testing.T) {
	strategy := &SemanticVersionStrategy{}

	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"Valid v1", "v1", true},
		{"Valid v2", "v2", true},
		{"Valid v10", "v10", true},
		{"Valid v999", "v999", true},
		{"Invalid no v prefix", "1", false},
		{"Invalid text", "version1", false},
		{"Invalid empty", "", false},
		{"Invalid date format", "2026-01", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, strategy.Validate(tt.version))

			if tt.valid {
				version, err := strategy.Parse(tt.version)
				assert.NoError(t, err)
				assert.Equal(t, APIVersion(tt.version), version)
			} else {
				_, err := strategy.Parse(tt.version)
				assert.Error(t, err)
			}
		})
	}
}

func TestDateVersionStrategy(t *testing.T) {
	strategy := &DateVersionStrategy{}

	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"Valid 2026-01", "2026-01", true},
		{"Valid 2026-12", "2026-12", true},
		{"Valid 2025-06", "2025-06", true},
		{"Invalid month 13", "2026-13", false},
		{"Invalid month 00", "2026-00", false},
		{"Invalid format", "202601", false},
		{"Invalid v1", "v1", false},
		{"Invalid empty", "", false},
		{"Invalid year only", "2026", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, strategy.Validate(tt.version))

			if tt.valid {
				version, err := strategy.Parse(tt.version)
				assert.NoError(t, err)
				assert.Equal(t, APIVersion(tt.version), version)
			} else {
				_, err := strategy.Parse(tt.version)
				assert.Error(t, err)
			}
		})
	}
}

func TestDateDailyVersionStrategy(t *testing.T) {
	strategy := &DateDailyVersionStrategy{}

	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{"Valid 20260115", "20260115", true},
		{"Valid 20261231", "20261231", true},
		{"Valid 20250601", "20250601", true},
		{"Invalid month 13", "20261345", false},
		{"Invalid day 32", "20260132", false},
		{"Invalid format with dashes", "2026-01-15", false},
		{"Invalid v1", "v1", false},
		{"Invalid empty", "", false},
		{"Invalid too short", "2026011", false},
		{"Invalid too long", "202601156", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, strategy.Validate(tt.version))

			if tt.valid {
				version, err := strategy.Parse(tt.version)
				assert.NoError(t, err)
				assert.Equal(t, APIVersion(tt.version), version)
			} else {
				_, err := strategy.Parse(tt.version)
				assert.Error(t, err)
			}
		})
	}
}

func TestCustomVersionStrategy(t *testing.T) {
	t.Run("Pattern-based validation", func(t *testing.T) {
		// Custom strategy: release-YYYY-QQ (e.g., release-2026-Q1)
		strategy, err := NewCustomVersionStrategy(
			"quarterly",
			`^release-\d{4}-Q[1-4]$`,
			nil,
		)
		require.NoError(t, err)

		tests := []struct {
			version string
			valid   bool
		}{
			{"release-2026-Q1", true},
			{"release-2026-Q4", true},
			{"release-2026-Q5", false},
			{"2026-Q1", false},
			{"v1", false},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.valid, strategy.Validate(tt.version), "version: %s", tt.version)
		}
	})

	t.Run("Explicit versions list", func(t *testing.T) {
		strategy, err := NewCustomVersionStrategy(
			"named",
			"",
			[]string{"alpha", "beta", "stable", "legacy"},
		)
		require.NoError(t, err)

		tests := []struct {
			version string
			valid   bool
		}{
			{"alpha", true},
			{"beta", true},
			{"stable", true},
			{"legacy", true},
			{"gamma", false},
			{"v1", false},
		}

		for _, tt := range tests {
			assert.Equal(t, tt.valid, strategy.Validate(tt.version), "version: %s", tt.version)
		}
	})

	t.Run("Invalid pattern", func(t *testing.T) {
		_, err := NewCustomVersionStrategy("bad", "[invalid(", nil)
		assert.Error(t, err)
	})
}

func TestGetStrategy(t *testing.T) {
	tests := []struct {
		name         string
		strategyType VersionStrategyType
		expectedName string
	}{
		{"Semantic", StrategyTypeSemantic, "semantic"},
		{"Date", StrategyTypeDate, "date"},
		{"DateDaily", StrategyTypeDateDaily, "date_daily"},
		{"Unknown defaults to semantic", "unknown", "semantic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := GetStrategy(tt.strategyType)
			assert.Equal(t, tt.expectedName, strategy.Name())
		})
	}
}

func TestNewDateVersion(t *testing.T) {
	version := NewDateVersion(2026, 1, "")
	assert.Equal(t, APIVersion("2026-01"), version)

	version = NewDateVersion(2026, 12, "")
	assert.Equal(t, APIVersion("2026-12"), version)
}

func TestNewDateDailyVersion(t *testing.T) {
	version := NewDateDailyVersion(2026, 1, 15)
	assert.Equal(t, APIVersion("20260115"), version)

	version = NewDateDailyVersion(2026, 12, 31)
	assert.Equal(t, APIVersion("20261231"), version)
}

func TestGetCurrentDateVersion(t *testing.T) {
	version := GetCurrentDateVersion()
	now := time.Now()
	expected := NewDateVersion(now.Year(), int(now.Month()), "-")
	assert.Equal(t, expected, version)
}

func TestGetCurrentDateDailyVersion(t *testing.T) {
	version := GetCurrentDateDailyVersion()
	now := time.Now()
	expected := NewDateDailyVersion(now.Year(), int(now.Month()), now.Day())
	assert.Equal(t, expected, version)
}

func TestParseDateVersion(t *testing.T) {
	version := APIVersion("2026-01")
	parsed, err := ParseDateVersion(version)
	require.NoError(t, err)
	assert.Equal(t, 2026, parsed.Year())
	assert.Equal(t, time.January, parsed.Month())
}

func TestParseDateDailyVersion(t *testing.T) {
	version := APIVersion("20260115")
	parsed, err := ParseDateDailyVersion(version)
	require.NoError(t, err)
	assert.Equal(t, 2026, parsed.Year())
	assert.Equal(t, time.January, parsed.Month())
	assert.Equal(t, 15, parsed.Day())
}

func TestSetVersionStrategy(t *testing.T) {
	// Save original strategy
	original := GetVersionStrategy()
	defer SetVersionStrategy(original)

	// Set to date strategy
	dateStrategy := &DateVersionStrategy{}
	SetVersionStrategy(dateStrategy)

	assert.Equal(t, dateStrategy, GetVersionStrategy())

	// Test that validation uses new strategy
	version := APIVersion("2026-01")
	assert.True(t, version.IsValid())

	semanticVersion := APIVersion("v1")
	assert.False(t, semanticVersion.IsValid())
}

func TestAPIVersion_IsValidWithStrategy(t *testing.T) {
	version := APIVersion("v1")

	// Valid for semantic
	semantic := &SemanticVersionStrategy{}
	assert.True(t, version.IsValidWithStrategy(semantic))

	// Invalid for date
	date := &DateVersionStrategy{}
	assert.False(t, version.IsValidWithStrategy(date))
}
