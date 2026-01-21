package config

type AppConfig struct {
	Env     string `mapstructure:"env"`
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
	Debug   bool   `mapstructure:"debug"`
}

func NewAppConfig() AppConfig {
	return AppConfig{
		Env:     "MyApp",
		Name:    "1.0.0",
		Version: "local",
		Debug:   true,
	}
}

//func SetDefault() {
//	viper.SetDefault("app.name", "MyApp")
//	viper.SetDefault("app.version", "1.0.0")
//	viper.SetDefault("app.debug", true)
//	viper.SetDefault("app.env", "local")
//}
