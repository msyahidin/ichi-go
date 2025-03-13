package config

import "github.com/spf13/viper"

type AppConfig struct {
	Env     string `mapstructure:"env"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

func SetDefault() {
	viper.SetDefault("app.name", "MyApp")
	viper.SetDefault("app.version", "1.0.0")
}
