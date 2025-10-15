package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"

	appConfig "ichi-go/config/app"
	cacheConfig "ichi-go/config/cache"
	dbConfig "ichi-go/config/database"
	httpConfig "ichi-go/config/http"
	logConfig "ichi-go/config/log"
	pkgClientConfig "ichi-go/config/pkgclient"
	"ichi-go/pkg/logger"
)

type Schema struct {
	App        appConfig.AppConfig
	Database   dbConfig.DatabaseConfig
	Cache      cacheConfig.CacheConfig
	Log        logConfig.LogConfig
	Http       httpConfig.HttpConfig
	HttpClient httpConfig.ClientConfig
	PkgClient  pkgClientConfig.PkgClient
}

type Config struct {
	schema *Schema
}

var (
	instance *Config
	once     sync.Once
	mu       sync.RWMutex
)

func Load() (*Config, error) {
	var loadErr error

	once.Do(func() {
		schema, err := loadSchema()
		if err != nil {
			loadErr = err
			return
		}

		mu.Lock()
		instance = &Config{schema: schema}
		mu.Unlock()

		configureLogging(instance, getEnv())
	})

	if loadErr != nil {
		return nil, loadErr
	}

	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		return nil, fmt.Errorf("config instance is nil")
	}

	return instance, nil
}

func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		logger.Panicf("Failed to load config: %v", err)
	}
	return cfg
}

func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()

	if instance == nil {
		panic("config not loaded: call Load() or MustLoad() first")
	}
	return instance
}

func (c *Config) Schema() *Schema {
	c.ensureLoaded()
	return c.schema
}

func (c *Config) App() appConfig.AppConfig {
	c.ensureLoaded()
	return c.schema.App
}

func (c *Config) Database() dbConfig.DatabaseConfig {
	c.ensureLoaded()
	return c.schema.Database
}

func (c *Config) Cache() cacheConfig.CacheConfig {
	c.ensureLoaded()
	return c.schema.Cache
}

func (c *Config) Http() httpConfig.HttpConfig {
	c.ensureLoaded()
	return c.schema.Http
}

func (c *Config) HttpClient() httpConfig.ClientConfig {
	c.ensureLoaded()
	return c.schema.HttpClient
}

func (c *Config) Log() logConfig.LogConfig {
	c.ensureLoaded()
	return c.schema.Log
}

func (c *Config) PkgClient() pkgClientConfig.PkgClient {
	c.ensureLoaded()
	return c.schema.PkgClient
}

// Private helpers
func (c *Config) ensureLoaded() {
	if c == nil {
		panic("config receiver is nil")
	}
	if c.schema == nil {
		panic("config schema is nil: config not properly initialized")
	}
}

func loadSchema() (*Schema, error) {
	env := getEnv()

	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	setDefault()

	var schema Schema
	if err := viper.Unmarshal(&schema); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	return &schema, nil
}

func getEnv() string {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "local"
	}
	return env
}

func setDefault() {
	appConfig.SetDefault()
	dbConfig.SetDefault()
	cacheConfig.SetDefault()
	logConfig.SetDefault()
	httpConfig.SetDefault()
}

func configureLogging(cfg *Config, env string) {
	if cfg == nil || cfg.schema == nil {
		logger.Warnf("Cannot configure logging: config is nil")
		return
	}

	if cfg.App().Debug {
		log.SetLevel(log.DEBUG)
		log.Debugf("Debugging enabled")
		log.Debugf("Configuration loaded successfully for environment: %s", env)
		log.Debugf("Loaded Config: %+v", cfg.schema)
	} else {
		log.SetLevel(log.INFO)
	}
}

func SetDebugMode(e *echo.Echo, debug bool) {
	e.Debug = debug
	if debug {
		log.SetLevel(log.DEBUG)
	} else {
		log.SetLevel(log.INFO)
	}
	log.Debugf("Debug mode set to %v", debug)
}
