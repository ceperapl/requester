package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	valid "github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

var (
	ErrTranslatorNotFound = errors.New("translator for en locale is not found")
	ErrValidation         = errors.New("validation error")
)

type Translation struct {
	Tag           string
	RegisterFn    valid.RegisterTranslationsFunc
	TranslationFn valid.TranslationFunc
}

type Validation struct {
	Validator  *valid.Validate
	Translator ut.Translator
}

type Option func(*Validation) error

func New(opts ...Option) (*Validation, error) {
	validator := valid.New()
	enTranslator := en.New()
	universalTranslator := ut.New(enTranslator, enTranslator)
	translator, found := universalTranslator.GetTranslator("en")
	if !found {
		return nil, ErrTranslatorNotFound
	}

	validation := &Validation{
		Validator:  validator,
		Translator: translator,
	}

	if err := en_translations.RegisterDefaultTranslations(validator, translator); err != nil {
		return nil, fmt.Errorf("couldn't register English translations: %w", err)
	}

	for _, opt := range opts {
		if err := opt(validation); err != nil {
			return nil, err
		}
	}

	return validation, nil
}

func WithJSONNamesForStructFields() Option {
	return func(v *Validation) error {
		v.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.Split(fld.Tag.Get("json"), ",")[0]

			return name
		})

		return nil
	}
}

func (v *Validation) ValidateStruct(s interface{}) error {
	var errStr string
	if err := v.Validator.Struct(s); err != nil {
		var invalValErr *valid.InvalidValidationError
		if errors.As(err, &invalValErr) {
			//nolint: wrapcheck
			return err
		}
		var valErrs valid.ValidationErrors
		if errors.As(err, &valErrs) {
			var errs []string
			for _, e := range valErrs {
				errs = append(errs, e.Translate(v.Translator))
			}
			errStr = strings.Join(errs, "; ")
		}

		return fmt.Errorf("%w: %s", ErrValidation, errStr)
	}

	return nil
}

func (v *Validation) ValidateVar(field interface{}, tag string) error {
	var errStr string
	if err := v.Validator.Var(field, tag); err != nil {
		var invalValErr *valid.InvalidValidationError
		if errors.As(err, &invalValErr) {
			//nolint: wrapcheck
			return err
		}
		var valErrs valid.ValidationErrors
		if errors.As(err, &valErrs) {
			var errs []string
			for _, e := range valErrs {
				errs = append(errs, e.Translate(v.Translator))
			}
			errStr = strings.Join(errs, "; ")
		}

		return fmt.Errorf("%w: %s", ErrValidation, errStr)
	}

	return nil
}
