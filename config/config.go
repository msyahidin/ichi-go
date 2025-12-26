package config

import (
	"fmt"
	"ichi-go/internal/infra/cache"
	"ichi-go/internal/infra/database"
	"ichi-go/internal/infra/messaging"
	"ichi-go/pkg/authenticator"
	"ichi-go/pkg/validator"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"

	appConfig "ichi-go/config/app"

	httpConfig "ichi-go/config/http"
	pkgClientConfig "ichi-go/pkg/clients"
	"ichi-go/pkg/logger"
)

type Schema struct {
	App        appConfig.AppConfig
	Database   database.Config
	Cache      cache.Config
	Log        logger.LogConfig
	Http       httpConfig.HttpConfig
	HttpClient httpConfig.ClientConfig
	PkgClient  pkgClientConfig.PkgClient
	Messaging  messaging.Config
	Auth       authenticator.Config
	Validator  validator.Config
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

// Get is kept for backward compatibility
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

func (c *Config) App() *appConfig.AppConfig {
	c.ensureLoaded()
	return &c.schema.App
}

func (c *Config) Database() *database.Config {
	c.ensureLoaded()
	return &c.schema.Database
}

func (c *Config) Cache() *cache.Config {
	c.ensureLoaded()
	return &c.schema.Cache
}

func (c *Config) Http() *httpConfig.HttpConfig {
	c.ensureLoaded()
	return &c.schema.Http
}

func (c *Config) HttpClient() *httpConfig.ClientConfig {
	c.ensureLoaded()
	return &c.schema.HttpClient
}

func (c *Config) Log() *logger.LogConfig {
	c.ensureLoaded()
	return &c.schema.Log
}

func (c *Config) PkgClient() *pkgClientConfig.PkgClient {
	c.ensureLoaded()
	return &c.schema.PkgClient
}

func (c *Config) Messaging() *messaging.Config {
	c.ensureLoaded()
	return &c.schema.Messaging
}

func (c *Config) Auth() *authenticator.Config {
	c.ensureLoaded()
	return &c.schema.Auth
}

func (c *Config) Validator() *validator.Config {
	c.ensureLoaded()
	return &c.schema.Validator
}

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
	database.SetDefault()
	cache.SetDefault()
	logger.SetDefault()
	httpConfig.SetDefault()
	messaging.SetDefault()
	authenticator.SetDefault()
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
