package database

import "github.com/spf13/viper"

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxConnLifeTime int    `mapstructure:"max_conn_life_time"`
	Debug           bool   `mapstructure:"debug"`
}

func SetDefault() {
	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.maxIdleConns", 10)
	viper.SetDefault("database.maxOpenConns", 100)
	viper.SetDefault("database.maxLifetime", 3600)
	viper.SetDefault("database.debug", false)
}
