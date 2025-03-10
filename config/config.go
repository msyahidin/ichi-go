package config

import (
	"fmt"
	"log"
	"os"
	appConfig "rathalos-kit/config/app"
	cacheConfig "rathalos-kit/config/cache"
	dbConfig "rathalos-kit/config/database"
	httpConfig "rathalos-kit/config/http"
	logConfig "rathalos-kit/config/log"

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
