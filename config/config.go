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

var Cfg Config

func LoadConfig() {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}

	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	err := viper.Unmarshal(&Cfg)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	log.Printf("Configuration loaded successfully for environment: %s", env)
}
