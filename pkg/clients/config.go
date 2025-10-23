package clients

type PkgClient struct {
	PokemonAPI PokemonAPIConfig `mapstructure:"pokemon_api"`
}

type PokemonAPIConfig struct {
	BaseURL    string `mapstructure:"base_url"`
	Timeout    int    `mapstructure:"timeout"` // in ms
	RetryCount int    `mapstructure:"retry_count"`
}
