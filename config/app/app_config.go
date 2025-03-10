package config

type AppConfig struct {
	Env  string `mapstructure:"env"`
	Name string `mapstructure:"name"`
}
