package validator

import (
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	id_translations "github.com/go-playground/validator/v10/translations/id"
)

// AppValidator wraps go-playground validator with translation support
type AppValidator struct {
	validator   *validator.Validate
	uni         *ut.UniversalTranslator
	transMap    map[string]ut.Translator
	defaultLang string
}

// NewValidator creates a new validator with translation support
func NewValidator(config Config) (*AppValidator, error) {
	v := validator.New()

	// Setup translators
	enLocale := en.New()
	idLocale := id.New()
	uni := ut.New(enLocale, enLocale, idLocale)

	// Get translators
	enTrans, _ := uni.GetTranslator("en")
	idTrans, _ := uni.GetTranslator("id")

	// Register default translations
	if err := en_translations.RegisterDefaultTranslations(v, enTrans); err != nil {
		return nil, err
	}
	if err := id_translations.RegisterDefaultTranslations(v, idTrans); err != nil {
		return nil, err
	}

	transMap := map[string]ut.Translator{
		"en": enTrans,
		"id": idTrans,
	}

	defaultLang := config.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}

	return &AppValidator{
		validator:   v,
		uni:         uni,
		transMap:    transMap,
		defaultLang: defaultLang,
	}, nil
}

// GetValidator returns the underlying validator instance
// Use this to register custom validators
func (av *AppValidator) GetValidator() *validator.Validate {
	return av.validator
}

// GetTranslator returns translator for a specific language
func (av *AppValidator) GetTranslator(lang string) ut.Translator {
	if trans, ok := av.transMap[lang]; ok {
		return trans
	}
	return av.transMap[av.defaultLang]
}

// Validate validates a struct and returns translated errors
func (av *AppValidator) Validate(data interface{}, lang string) *ValidationErrors {
	err := av.validator.Struct(data)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return &ValidationErrors{
			Errors: []FieldError{
				{
					Field:   "unknown",
					Message: err.Error(),
				},
			},
		}
	}

	trans := av.GetTranslator(lang)
	return FormatValidationErrors(validationErrors, trans)
}

// ValidateStruct is a convenience method that uses the default language
func (av *AppValidator) ValidateStruct(data interface{}) *ValidationErrors {
	return av.Validate(data, av.defaultLang)
}
