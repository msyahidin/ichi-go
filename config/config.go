package config

import (
	"fmt"
	"log"
	"os"
	config "rathalos-kit/config/app"

	"github.com/spf13/viper"
)

var AppConfig config.AppConfig

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

	err := viper.Unmarshal(&AppConfig)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	log.Printf("Configuration loaded successfully for environment: %s", env)
}
