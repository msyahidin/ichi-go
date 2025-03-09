package config

type LogConfig struct {
	Level           string          `mapstructure:"level"`
	RequestIDConfig RequestIDConfig `mapstructure:",squash"`
}

type RequestIDConfig struct {
	Driver string `mapstructure:"driver"`
}
