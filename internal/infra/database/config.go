package database

import "github.com/spf13/viper"

type Config struct {
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
	viper.SetDefault("primary_database", "mysql")
	viper.SetDefault("databases.mysql.driver", "mysql")
	viper.SetDefault("databases.mysql.host", "localhost")
	viper.SetDefault("databases.mysql.port", 3306)
	viper.SetDefault("databases.mysql.user", "root")
	viper.SetDefault("databases.mysql.password", "password")
	viper.SetDefault("databases.mysql.name", "ichi_app")
	viper.SetDefault("databases.mysql.max_idle_conns", 10)
	viper.SetDefault("databases.mysql.max_open_conns", 100)
	viper.SetDefault("databases.mysql.max_conn_life_time", 3600)
	viper.SetDefault("databases.mysql.debug", false)
}
