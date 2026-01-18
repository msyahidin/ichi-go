package versioning

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config represents API versioning configuration
type Config struct {
	Enabled           bool              `mapstructure:"enabled"`
	Strategy          string            `mapstructure:"strategy"` // "semantic", "date", "date_daily", "custom"
	DefaultVersion    string            `mapstructure:"default_version"`
	SupportedVersions []string          `mapstructure:"supported_versions"`
	Deprecation       DeprecationConfig `mapstructure:"deprecation"`

	// Custom strategy settings (optional)
	CustomPattern       string   `mapstructure:"custom_pattern"`
	CustomValidVersions []string `mapstructure:"custom_valid_versions"`
}

// DeprecationConfig represents deprecation settings
type DeprecationConfig struct {
	HeaderEnabled          bool `mapstructure:"header_enabled"`
	SunsetNotificationDays int  `mapstructure:"sunset_notification_days"`
}

// SetDefault sets default versioning configuration
func SetDefault() {
	viper.SetDefault("api.versioning.enabled", true)
	viper.SetDefault("api.versioning.strategy", "semantic")
	viper.SetDefault("api.versioning.default_version", "v1")
	viper.SetDefault("api.versioning.supported_versions", []string{"v1"})
	viper.SetDefault("api.versioning.deprecation.header_enabled", true)
	viper.SetDefault("api.versioning.deprecation.sunset_notification_days", 90)
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DefaultVersion == "" {
		c.DefaultVersion = "v1"
	}

	if len(c.SupportedVersions) == 0 {
		c.SupportedVersions = []string{"v1"}
	}

	// Set strategy if not specified
	if c.Strategy == "" {
		c.Strategy = "semantic"
	}

	return nil
}

// GetStrategy returns the version strategy based on config
func (c *Config) GetStrategy() (VersionStrategy, error) {
	strategyType := VersionStrategyType(c.Strategy)

	switch strategyType {
	case StrategyTypeCustom:
		if c.CustomPattern == "" && len(c.CustomValidVersions) == 0 {
			return nil, fmt.Errorf("custom_pattern or custom_valid_versions is required for custom strategy")
		}
		return NewCustomVersionStrategy("custom", c.CustomPattern, c.CustomValidVersions)
	default:
		return GetStrategy(strategyType), nil
	}
}

// InitializeStrategy initializes the global version strategy from config
func (c *Config) InitializeStrategy() error {
	strategy, err := c.GetStrategy()
	if err != nil {
		return err
	}
	SetVersionStrategy(strategy)
	return nil
}

// GetDefaultVersion returns the default API version
func (c *Config) GetDefaultVersion() APIVersion {
	version, err := ParseVersion(c.DefaultVersion)
	if err != nil {
		return V1 // Fallback to V1
	}
	return version
}

// IsVersionSupported checks if a version is supported
func (c *Config) IsVersionSupported(version string) bool {
	for _, v := range c.SupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}

// GetSupportedVersions returns all supported versions as APIVersion
func (c *Config) GetSupportedVersions() []APIVersion {
	versions := make([]APIVersion, 0, len(c.SupportedVersions))
	for _, v := range c.SupportedVersions {
		if version, err := ParseVersion(v); err == nil {
			versions = append(versions, version)
		}
	}
	return versions
}
