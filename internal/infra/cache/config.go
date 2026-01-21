package cache

type Config struct {
	Driver     string `mapstructure:"driver"`
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	Username   string `mapstructure:"username"`
	Db         int    `mapstructure:"db"`
	PoolSize   int    `mapstructure:"pool_size"`
	Timeout    int    `mapstructure:"timeout"`
	UseTLS     bool   `mapstructure:"use_tls"`
	SkipVerify bool   `mapstructure:"skip_verify"`
	ClientName string `mapstructure:"client_name"`
}

func NewConfig() Config {
	return Config{
		Driver:     "redis",
		Host:       "localhost",
		Port:       6379,
		Username:   "",
		Password:   "",
		Db:         0,
		PoolSize:   10,
		UseTLS:     false,
		SkipVerify: false,
		Timeout:    30,
		ClientName: "ichigo-cache",
	}
}

//func SetDefault() {
//	viper.SetDefault("cache.driver", "redis")
//	viper.SetDefault("cache.host", "localhost")
//	viper.SetDefault("cache.port", 6379)
//	viper.SetDefault("cache.username", "")
//	viper.SetDefault("cache.password", "")
//	viper.SetDefault("cache.db", 0)
//	viper.SetDefault("cache.pool_size", 10)
//	viper.SetDefault("cache.use_tls", false)
//	viper.SetDefault("cache.skip_verify", false)
//	viper.SetDefault("cache.timeout", 30)
//	viper.SetDefault("cache.client_name", "ichigo-cache")
//}
