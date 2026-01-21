package logger

type LogConfig struct {
	Level           string               `mapstructure:"level"`
	Pretty          bool                 `mapstructure:"pretty"`
	RequestIDConfig RequestIDConfig      `mapstructure:",squash"`
	RequestLogging  RequestLoggingConfig `mapstructure:"request_logging"`
}

func NewLogConfig() LogConfig {
	return LogConfig{
		Level:  "info",
		Pretty: false,
		RequestIDConfig: RequestIDConfig{
			Driver: "builtin",
		},
		RequestLogging: RequestLoggingConfig{
			Enabled: false,
			Driver:  "builtin",
		},
	}
}

type RequestIDConfig struct {
	Driver string `mapstructure:"driver"`
}

type RequestLoggingConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Driver  string `mapstructure:"driver"`
}

//func SetDefault() {
//	viper.SetDefault("log.level", "info")
//	viper.SetDefault("log.pretty", false)
//	viper.SetDefault("log.request_logging.enabled", false)
//	viper.SetDefault("log.request_logging.driver", "builtin")
//	viper.SetDefault("log.request_id.driver", "builtin")
//}
