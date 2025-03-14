package http

import (
	"github.com/spf13/viper"
)

type CorsConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type HttpConfig struct {
	Timeout int        `mapstructure:"timeout"`
	Cors    CorsConfig `mapstructure:"cors"`
	Port    int        `mapstructure:"port"`
}

type ClientConfig struct {
	Timeout       int  `yaml:"timeout"`         // request timeout (per call)
	RetryCount    int  `yaml:"retry_count"`     // total max retries
	RetryWaitTime int  `yaml:"retry_wait_time"` // delay between retries
	RetryMaxWait  int  `yaml:"retry_max_wait"`  // total max backoff wait
	LoggerEnabled bool `yaml:"logger_enabled"`  // enable/disable http log
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
