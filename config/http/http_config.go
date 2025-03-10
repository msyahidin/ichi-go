package http

type CorsConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type HttpConfig struct {
	Timeout int        `mapstructure:"timeout"`
	Cors    CorsConfig `mapstructure:"cors"`
	Port    int        `mapstructure:"port"`
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
