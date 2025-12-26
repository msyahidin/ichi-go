package validator

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// GetLanguageFromContext extracts language from Accept-Language or X-Language header
func GetLanguageFromContext(c echo.Context) string {
	// Try X-Language custom header first (highest priority)
	customLang := c.Request().Header.Get("X-Language")
	if customLang == "id" || customLang == "en" {
		return customLang
	}

	// Try Accept-Language header
	acceptLang := c.Request().Header.Get("Accept-Language")
	if acceptLang != "" {
		// Parse "en-US,en;q=0.9,id;q=0.8" format
		langs := strings.Split(acceptLang, ",")
		if len(langs) > 0 {
			// Get the first language preference
			primary := strings.Split(langs[0], "-")[0]
			primary = strings.Split(primary, ";")[0]
			primary = strings.TrimSpace(primary)

			// Support en and id
			if primary == "id" {
				return "id"
			}
		}
	}

	return "en" // default
}

// GetLanguageFromRequest extracts language from HTTP headers
func GetLanguageFromRequest(headers map[string]string) string {
	// Try X-Language custom header first
	if lang, ok := headers["X-Language"]; ok {
		if lang == "id" || lang == "en" {
			return lang
		}
	}

	// Try Accept-Language header
	if acceptLang, ok := headers["Accept-Language"]; ok && acceptLang != "" {
		langs := strings.Split(acceptLang, ",")
		if len(langs) > 0 {
			primary := strings.Split(langs[0], "-")[0]
			primary = strings.Split(primary, ";")[0]
			primary = strings.TrimSpace(primary)

			if primary == "id" {
				return "id"
			}
		}
	}

	return "en" // default
}

// NormalizeLanguage ensures the language code is valid
func NormalizeLanguage(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))

	switch lang {
	case "id", "indonesia", "indonesian":
		return "id"
	case "en", "english":
		return "en"
	default:
		return "en"
	}
}
