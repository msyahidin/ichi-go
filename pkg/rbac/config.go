package rbac

import "github.com/spf13/viper"

// Config defines RBAC system configuration
type Config struct {
	// Mode determines deployment strategy
	// - "multi_tenant": Strict tenant isolation, tenant_id required
	// - "single_tenant": Auto-inject default tenant, simplified for admin tools
	// - "hybrid": Optional tenant_id with fallback to default
	Mode string `mapstructure:"mode"`

	// DefaultTenant is used in single_tenant and hybrid modes
	DefaultTenant string `mapstructure:"default_tenant"`

	// ModelPath is the file path to Casbin model configuration
	ModelPath string `mapstructure:"model_path"`

	// Performance tuning settings
	Performance PerformanceConfig `mapstructure:"performance"`

	// Cache configuration for multi-tier caching
	Cache CacheConfig `mapstructure:"cache"`

	// Audit logging configuration for compliance
	Audit AuditConfig `mapstructure:"audit"`

	// Features toggles for optional functionality
	Features FeaturesConfig `mapstructure:"features"`
}

// PerformanceConfig configures RBAC performance optimizations
type PerformanceConfig struct {
	// LoadingStrategy determines how policies are loaded
	// - "full": Load all policies into memory (best for single-tenant)
	// - "filtered": Load only specific tenant policies (best for multi-tenant)
	// - "adaptive": Switch between full/filtered based on policy count
	LoadingStrategy string `mapstructure:"loading_strategy"`

	// FilteredTTL is how long filtered policies stay in enforcer
	FilteredTTL string `mapstructure:"filtered_ttl"`

	// MaxConcurrent limits concurrent permission checks
	MaxConcurrent int `mapstructure:"max_concurrent"`

	// AdaptiveThreshold is policy count threshold for adaptive mode
	AdaptiveThreshold int `mapstructure:"adaptive_threshold"`
}

// CacheConfig configures multi-tier decision caching
type CacheConfig struct {
	// Enabled toggles all caching (disable for debugging)
	Enabled bool `mapstructure:"enabled"`

	// MemoryTTL is L1 in-memory cache TTL
	MemoryTTL string `mapstructure:"memory_ttl"`

	// RedisTTL is L2 Redis cache TTL
	RedisTTL string `mapstructure:"redis_ttl"`

	// MaxSize is max number of entries in L1 cache
	MaxSize int `mapstructure:"max_size"`

	// Compression enables LZ4 compression for Redis cache
	Compression bool `mapstructure:"compression"`
}

// AuditConfig configures audit logging for compliance (SOC2/GDPR)
type AuditConfig struct {
	// Enabled toggles all audit logging
	Enabled bool `mapstructure:"enabled"`

	// LogDecisions logs every CheckPermission call (high volume)
	LogDecisions bool `mapstructure:"log_decisions"`

	// LogMutations logs policy/role changes (recommended for compliance)
	LogMutations bool `mapstructure:"log_mutations"`

	// RetentionDays is how long to keep audit logs (2555 = 7 years for SOC2)
	RetentionDays int `mapstructure:"retention_days"`

	// AnonymizePII enables SHA-256 hashing of email addresses (GDPR)
	AnonymizePII bool `mapstructure:"anonymize_pii"`

	// SampleRate controls decision logging sampling (0.0-1.0)
	// 0.01 = log 1% of decisions
	SampleRate float64 `mapstructure:"sample_rate"`
}

// FeaturesConfig toggles optional RBAC features
type FeaturesConfig struct {
	// TimeBoundRoles enables role assignments with expiration
	TimeBoundRoles bool `mapstructure:"time_bound_roles"`

	// Impersonation enables platform admin user impersonation
	Impersonation bool `mapstructure:"impersonation"`

	// ApprovalWorkflows enables multi-party approval for sensitive operations
	ApprovalWorkflows bool `mapstructure:"approval_workflows"`

	// ResourceLevelABAC enables fine-grained resource permissions
	ResourceLevelABAC bool `mapstructure:"resource_level_abac"`
}

// SetDefault sets default RBAC configuration values
func SetDefault() {
	// Mode defaults
	viper.SetDefault("rbac.mode", "hybrid")
	viper.SetDefault("rbac.default_tenant", "system")
	viper.SetDefault("rbac.model_path", "config/rbac_model.conf")

	// Performance defaults
	viper.SetDefault("rbac.performance.loading_strategy", "filtered")
	viper.SetDefault("rbac.performance.filtered_ttl", "10m")
	viper.SetDefault("rbac.performance.max_concurrent", 50)
	viper.SetDefault("rbac.performance.adaptive_threshold", 5000)

	// Cache defaults
	viper.SetDefault("rbac.cache.enabled", true)
	viper.SetDefault("rbac.cache.memory_ttl", "60s")
	viper.SetDefault("rbac.cache.redis_ttl", "5m")
	viper.SetDefault("rbac.cache.max_size", 10000)
	viper.SetDefault("rbac.cache.compression", true)

	// Audit defaults
	viper.SetDefault("rbac.audit.enabled", true)
	viper.SetDefault("rbac.audit.log_decisions", false) // High volume
	viper.SetDefault("rbac.audit.log_mutations", true)
	viper.SetDefault("rbac.audit.retention_days", 2555) // 7 years
	viper.SetDefault("rbac.audit.anonymize_pii", true)
	viper.SetDefault("rbac.audit.sample_rate", 0.01) // 1% of decisions

	// Feature flags (all disabled by default for Phase 1)
	viper.SetDefault("rbac.features.time_bound_roles", false)
	viper.SetDefault("rbac.features.impersonation", false)
	viper.SetDefault("rbac.features.approval_workflows", false)
	viper.SetDefault("rbac.features.resource_level_abac", false)
}

// Validate validates the RBAC configuration
func (c *Config) Validate() error {
	// Mode validation
	validModes := map[string]bool{
		"multi_tenant":  true,
		"single_tenant": true,
		"hybrid":        true,
	}
	if !validModes[c.Mode] {
		return &ValidationError{
			Field:   "mode",
			Value:   c.Mode,
			Message: "must be one of: multi_tenant, single_tenant, hybrid",
		}
	}

	// Loading strategy validation
	validStrategies := map[string]bool{
		"full":     true,
		"filtered": true,
		"adaptive": true,
	}
	if !validStrategies[c.Performance.LoadingStrategy] {
		return &ValidationError{
			Field:   "performance.loading_strategy",
			Value:   c.Performance.LoadingStrategy,
			Message: "must be one of: full, filtered, adaptive",
		}
	}

	// Model path validation
	if c.ModelPath == "" {
		return &ValidationError{
			Field:   "model_path",
			Value:   c.ModelPath,
			Message: "cannot be empty",
		}
	}

	return nil
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e *ValidationError) Error() string {
	return "rbac config validation failed: " + e.Field + " (" + e.Value + ") " + e.Message
}

// IsMultiTenant returns true if running in strict multi-tenant mode
func (c *Config) IsMultiTenant() bool {
	return c.Mode == "multi_tenant"
}

// IsSingleTenant returns true if running in single-tenant mode
func (c *Config) IsSingleTenant() bool {
	return c.Mode == "single_tenant"
}

// IsHybrid returns true if running in hybrid mode
func (c *Config) IsHybrid() bool {
	return c.Mode == "hybrid"
}

// RequiresTenantID returns true if tenant_id must be provided in requests
func (c *Config) RequiresTenantID() bool {
	return c.Mode == "multi_tenant"
}
