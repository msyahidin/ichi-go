package config

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	appConfig "ichi-go/config/app"
	cacheConfig "ichi-go/config/cache"
	dbConfig "ichi-go/config/database"
	httpConfig "ichi-go/config/http"
	logConfig "ichi-go/config/log"
	msgConfig "ichi-go/config/message"
	pkgClientConfig "ichi-go/config/pkgclient"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	App        appConfig.AppConfig
	Database   dbConfig.DatabaseConfig
	Cache      cacheConfig.CacheConfig
	Log        logConfig.LogConfig
	Http       httpConfig.HttpConfig
	HttpClient httpConfig.ClientConfig
	PkgClient  pkgClientConfig.PkgClient
	Messages   msgConfig.MessageConfig
}

var Cfg *Config

func setDefault() {
	appConfig.SetDefault()
	dbConfig.SetDefault()
	cacheConfig.SetDefault()
	logConfig.SetDefault()
	httpConfig.SetDefault()
	msgConfig.SetDefault()
}

func LoadConfig(e *echo.Echo) *Config {
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
	setDefault()
	var cfg Config
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}
	Cfg = &cfg
	SetDebugMode(e, Cfg.App.Debug)
	if e.Debug {
		log.SetLevel(log.DEBUG)
		log.Debugf("Debugging enabled")
		log.Debugf("Configuration loaded successfully for environment: %s", env)
		log.Debugf("Loaded MessageConfig: %+v", *Cfg)
	} else {
		log.SetLevel(log.INFO)
	}
	return Cfg
}

func App() appConfig.AppConfig {
	return Cfg.App
}

func SetDebugMode(e *echo.Echo, debug bool) {
	Cfg.App.Debug = debug
	e.Debug = debug
	if debug {
		log.SetLevel(log.DEBUG)
	} else {
		log.SetLevel(log.INFO)
	}
	log.Debugf("Debug mode set to %v", debug)
}

func Database() dbConfig.DatabaseConfig {
	return Cfg.Database
}

func Cache() cacheConfig.CacheConfig {
	return Cfg.Cache
}

func Http() httpConfig.HttpConfig {
	return Cfg.Http
}
func HttpClient() httpConfig.ClientConfig {
	return Cfg.HttpClient
}

func Log() logConfig.LogConfig {
	return Cfg.Log
}

func PkgClient() pkgClientConfig.PkgClient {
	return Cfg.PkgClient
}

func Message() msgConfig.MessageConfig { return Cfg.Messages }
