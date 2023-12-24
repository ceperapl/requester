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

// Define some variables for the custom errors that might occur when using the validator package.
var (
	ErrTranslatorNotFound = errors.New("translator for en locale is not found")
	ErrValidation         = errors.New("validation error")
)

// Validator is an interface that defines the methods for validating data.
// It can validate structs or variables using validation tags.
type Validator interface {
	ValidateStruct(s interface{}) error
	ValidateVar(field interface{}, tag string) error
}

// Validation is a type that encapsulates a validator.
type Validation struct {
	Validator  *valid.Validate
	Translator ut.Translator
}

// Option is a type that represents a function that modifies the validation settings.
type Option func(*Validation) error

// New is a function that creates and returns a new validation instance with the given options.
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

// WithJSONNamesForStructFields is an option function that
// sets the tag name function for the validator to use the JSON names of the struct fields.
func WithJSONNamesForStructFields() Option {
	return func(v *Validation) error {
		v.Validator.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.Split(fld.Tag.Get("json"), ",")[0]

			return name
		})

		return nil
	}
}

// ValidateStruct validates a struct using the validator and the translator.
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

// ValidateVar validates a variable using the validator and the translator.
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
