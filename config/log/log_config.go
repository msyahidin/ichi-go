package config

import "github.com/spf13/viper"

type LogConfig struct {
	Level           string          `mapstructure:"level"`
	RequestIDConfig RequestIDConfig `mapstructure:",squash"`
}

type RequestIDConfig struct {
	Driver string `mapstructure:"driver"`
}

func SetDefault() {
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.request_id.driver", "header")
}
