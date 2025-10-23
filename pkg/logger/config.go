package logger

import "github.com/spf13/viper"

type LogConfig struct {
	Level           string               `mapstructure:"level"`
	Pretty          bool                 `mapstructure:"pretty"`
	RequestIDConfig RequestIDConfig      `mapstructure:",squash"`
	RequestLogging  RequestLoggingConfig `mapstructure:"request_logging"`
}

type RequestIDConfig struct {
	Driver string `mapstructure:"driver"`
}

type RequestLoggingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Driver  string `mapstructure:"driver"`
}

func SetDefault() {
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.pretty", false)
	viper.SetDefault("log.request_logging.enabled", false)
	viper.SetDefault("log.request_logging.driver", "builtin")
	viper.SetDefault("log.request_id.driver", "builtin")
}
