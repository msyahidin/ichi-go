package config

import (
	configDB "rathalos-kit/config/database"
	configLog "rathalos-kit/config/log"
)

type AppConfig struct {
	Env      string                  `mapstructure:"env"`
	Name     string                  `mapstructure:"name"`
	Request  RequestConfig           `mapstructure:"request_timeout"`
	Server   ServerConfig            `mapstructure:"server"`
	Database configDB.DatabaseConfig `mapstructure:"database"`
	Log      configLog.LogConfig     `mapstructure:"log"`
}

type ServerConfig struct {
	Rest RestConfig `mapstructure:"rest"`
	Web  WebConfig  `mapstructure:"web"`
}

type RestConfig struct {
	Port int `mapstructure:"port"`
}
type WebConfig struct {
	Port int `mapstructure:"port"`
}

type RequestConfig struct {
	Timeout int `mapstructure:"timeout"`
}
