package pokemonapi

type Config struct {
	BaseURL    string `yaml:"base_url"`
	Timeout    int    `yaml:"timeout"` // in ms
	RetryCount int    `yaml:"retry_count"`
}
