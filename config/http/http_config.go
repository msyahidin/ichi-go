package http

import "github.com/spf13/viper"

type CorsConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type HttpConfig struct {
	Timeout int        `mapstructure:"timeout"`
	Cors    CorsConfig `mapstructure:"cors"`
	Port    int        `mapstructure:"port"`
}

func SetDefault() {
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("http.timeout", 30)
	viper.SetDefault("http.cors.allow_origins", []string{"*"})
}
