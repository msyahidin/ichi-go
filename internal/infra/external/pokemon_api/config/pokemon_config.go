package config

type PokemonAPIConfig struct {
	BaseURL    string `yaml:"base_url"`
	Timeout    int    `yaml:"timeout"` // in ms
	RetryCount int    `yaml:"retry_count"`
}
