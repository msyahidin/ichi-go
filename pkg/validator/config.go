package validator

import "github.com/spf13/viper"

// Config represents validator configuration
type Config struct {
	DefaultLanguage    string   `mapstructure:"default_language"`
	SupportedLanguages []string `mapstructure:"supported_languages"`
}

// SetDefault sets default validator configuration
func SetDefault() {
	viper.SetDefault("validator.default_language", "en")
	viper.SetDefault("validator.supported_languages", []string{"en", "id"})
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Check if default language is supported
	if c.DefaultLanguage == "" {
		c.DefaultLanguage = "en"
	}

	if len(c.SupportedLanguages) == 0 {
		c.SupportedLanguages = []string{"en", "id"}
	}

	return nil
}

// IsLanguageSupported checks if a language is supported
func (c *Config) IsLanguageSupported(lang string) bool {
	for _, supported := range c.SupportedLanguages {
		if supported == lang {
			return true
		}
	}
	return false
}

// GetDefaultLanguage returns the default language
func (c *Config) GetDefaultLanguage() string {
	if c.DefaultLanguage == "" {
		return "en"
	}
	return c.DefaultLanguage
}
