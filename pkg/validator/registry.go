package validator

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// CustomValidator represents a custom validation function with translations
type CustomValidator struct {
	Tag            string
	Fn             validator.Func
	TranslationKey string
	RegisterTrans  func(ut.Translator) error
}

// RegisterCustomValidator registers a custom validator with translations
func (av *AppValidator) RegisterCustomValidator(cv CustomValidator) error {
	// Register validation function
	if err := av.validator.RegisterValidation(cv.Tag, cv.Fn); err != nil {
		return err
	}

	// Register translations for all languages
	for lang, trans := range av.transMap {
		if cv.RegisterTrans != nil {
			if err := cv.RegisterTrans(trans); err != nil {
				return err
			}
		} else {
			// Default translation registration
			if err := av.registerDefaultTranslation(cv.Tag, trans, lang); err != nil {
				return err
			}
		}
	}

	return nil
}

// RegisterCustomValidators registers multiple custom validators
func (av *AppValidator) RegisterCustomValidators(validators []CustomValidator) error {
	for _, cv := range validators {
		if err := av.RegisterCustomValidator(cv); err != nil {
			return err
		}
	}
	return nil
}

// registerDefaultTranslation provides a default translation if none specified
func (av *AppValidator) registerDefaultTranslation(tag string, trans ut.Translator, lang string) error {
	messages := map[string]map[string]string{
		"en": {
			"default": "{0} is invalid",
		},
		"id": {
			"default": "{0} tidak valid",
		},
	}

	message := messages[lang]["default"]

	return trans.Add(tag, message, true)
}

// RegisterTranslation registers a custom translation for a validation tag
func (av *AppValidator) RegisterTranslation(tag string, lang string, message string, override bool) error {
	trans := av.GetTranslator(lang)

	return av.validator.RegisterTranslation(
		tag,
		trans,
		func(ut ut.Translator) error {
			return ut.Add(tag, message, override)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(tag, fe.Field())
			return t
		},
	)
}
