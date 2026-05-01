package database

import "github.com/spf13/viper"

type Config struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	SSLMode         string `mapstructure:"ssl_mode"` // postgres only; defaults to "disable"
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxConnLifeTime int    `mapstructure:"max_conn_life_time"`
	Debug           bool   `mapstructure:"debug"`
}

func SetDefault() {
	viper.SetDefault("database.default", "mysql")
	viper.SetDefault("database.connections.mysql.driver", "mysql")
	viper.SetDefault("database.connections.mysql.host", "localhost")
	viper.SetDefault("database.connections.mysql.port", 3306)
	viper.SetDefault("database.connections.mysql.user", "root")
	viper.SetDefault("database.connections.mysql.password", "password")
	viper.SetDefault("database.connections.mysql.name", "ichi_app")
	viper.SetDefault("database.connections.mysql.max_idle_conns", 10)
	viper.SetDefault("database.connections.mysql.max_open_conns", 100)
	viper.SetDefault("database.connections.mysql.max_conn_life_time", 3600)
	viper.SetDefault("database.connections.mysql.debug", false)
}
