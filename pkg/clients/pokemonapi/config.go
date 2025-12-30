package pokemonapi

import "time"

type Config struct {
	BaseURL    string `mapstructure:"base_url"`
	Timeout    int    `mapstructure:"timeout"` // in milliseconds
	RetryCount int    `mapstructure:"retry_count"`

	// Optional: override default HTTP client settings
	RetryWaitTime *int  `mapstructure:"retry_wait_time"` // in ms
	RetryMaxWait  *int  `mapstructure:"retry_max_wait"`  // in ms
	LoggerEnabled *bool `mapstructure:"logger_enabled"`
}

// MergeWithDefaults merges Pokemon config with HTTP client defaults
func (c *Config) MergeWithDefaults(httpDefaults HTTPClientDefaults) HTTPClientOptions {
	opts := HTTPClientOptions{
		BaseURL:       c.BaseURL,
		Timeout:       time.Duration(c.Timeout) * time.Millisecond,
		RetryCount:    c.RetryCount,
		RetryWaitTime: time.Duration(httpDefaults.RetryWaitTime) * time.Millisecond,
		RetryMaxWait:  time.Duration(httpDefaults.RetryMaxWait) * time.Millisecond,
		LoggerEnabled: httpDefaults.LoggerEnabled,
	}

	// Override with Pokemon-specific settings if provided
	if c.RetryWaitTime != nil {
		opts.RetryWaitTime = time.Duration(*c.RetryWaitTime) * time.Millisecond
	}
	if c.RetryMaxWait != nil {
		opts.RetryMaxWait = time.Duration(*c.RetryMaxWait) * time.Millisecond
	}
	if c.LoggerEnabled != nil {
		opts.LoggerEnabled = *c.LoggerEnabled
	}

	return opts
}

// HTTPClientDefaults represents default HTTP client settings
type HTTPClientDefaults struct {
	RetryWaitTime int
	RetryMaxWait  int
	LoggerEnabled bool
}

// HTTPClientOptions represents final merged HTTP client options
type HTTPClientOptions struct {
	BaseURL       string
	Timeout       time.Duration
	RetryCount    int
	RetryWaitTime time.Duration
	RetryMaxWait  time.Duration
	LoggerEnabled bool
}
