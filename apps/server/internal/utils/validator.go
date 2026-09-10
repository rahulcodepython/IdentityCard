package utils

import (
    "errors"
    "reflect"
    "strings"

    "github.com/go-playground/locales/en"
    ut "github.com/go-playground/universal-translator"
    "github.com/go-playground/validator/v10"
    entranslations "github.com/go-playground/validator/v10/translations/en"
)

var (
    validate  = validator.New()
    validateT ut.Translator
)

func init() {
    locale := en.New()
    translator := ut.New(locale, locale)
    validateT, _ = translator.GetTranslator("en")
    // Humanizes struct-tag validation errors ("Title is a required field")
    // instead of the raw "Field: tag" pairing.
    _ = entranslations.RegisterDefaultTranslations(validate, validateT)

    // Report validation errors using each field's `json` tag so they line
    // up with the zod schema field names on the frontend, not the Go
    // struct field names.
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
        if name == "-" || name == "" {
            return fld.Name
        }
        return name
    })
}

// ValidateStruct validates a struct according to its struct tags and translates error messages.
func ValidateStruct(s interface{}) error {
    if err := validate.Struct(s); err != nil {
        var errs []string
        if valErrors, ok := err.(validator.ValidationErrors); ok {
            for _, fe := range valErrors {
                errs = append(errs, fe.Translate(validateT))
            }
            return errors.New(strings.Join(errs, "; "))
        }
        return err
    }
    return nil
}
