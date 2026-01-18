package http

import (
	"time"

	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type Config struct {
	Timeout int        `mapstructure:"timeout"`
	Cors    CorsConfig `mapstructure:"cors"`
	Port    int        `mapstructure:"port"`
}

type ClientConfig struct {
	Timeout       int  `mapstructure:"timeout"`         // request timeout (per call)
	RetryCount    int  `mapstructure:"retry_count"`     // total max retries
	RetryWaitTime int  `mapstructure:"retry_wait_time"` // delay between retries
	RetryMaxWait  int  `mapstructure:"retry_max_wait"`  // total max backoff wait
	LoggerEnabled bool `mapstructure:"logger_enabled"`  // enable/disable http log
}

// ToOptions converts ClientConfig to ClientOptions
func (c *ClientConfig) ToOptions() ClientOptions {
	return ClientOptions{
		Timeout:       time.Duration(c.Timeout) * time.Millisecond,
		RetryCount:    c.RetryCount,
		RetryWaitTime: time.Duration(c.RetryWaitTime) * time.Millisecond,
		RetryMaxWait:  time.Duration(c.RetryMaxWait) * time.Millisecond,
		LoggerEnabled: c.LoggerEnabled,
	}
}

func SetDefault() {
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("http.timeout", 30)
	viper.SetDefault("http.cors.allow_origins", []string{"*"})

	viper.SetDefault("http_client.timeout", 60000)
	viper.SetDefault("http_client.retry_count", 3)
	viper.SetDefault("http_client.retry_wait_time", 5000)
	viper.SetDefault("http_client.retry_max_wait", 5)
	viper.SetDefault("http_client.logger_enabled", false)
}
