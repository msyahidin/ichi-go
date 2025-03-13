package config

import (
	"fmt"
	appConfig "ichi-go/config/app"
	cacheConfig "ichi-go/config/cache"
	dbConfig "ichi-go/config/database"
	httpConfig "ichi-go/config/http"
	logConfig "ichi-go/config/log"
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App      appConfig.AppConfig
	Database dbConfig.DatabaseConfig
	Cache    cacheConfig.CacheConfig
	Log      logConfig.LogConfig
	Http     httpConfig.HttpConfig
}

var Cfg *Config

func setDefault() {
	appConfig.SetDefault()
	dbConfig.SetDefault()
	cacheConfig.SetDefault()
	logConfig.SetDefault()
	httpConfig.SetDefault()
}

func LoadConfig() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	setDefault()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}
	Cfg = &cfg

	log.Printf("Configuration loaded successfully for environment: %s", env)
	log.Printf("Loaded Config: %+v", *Cfg)
}
