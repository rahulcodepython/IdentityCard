package utils

import (
    "errors"
    "reflect"
    "strings"

    "github.com/go-playground/validator/v10"
    "github.com/gofiber/fiber/v2"

    "identitycard-server/internal/generic"
)

var validate = validator.New()

func init() {
    // Map validation error keys to JSON struct tag names instead of Go field names
    validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
        name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
        if name == "-" || name == "" {
            return fld.Name
        }
        return name
    })
}

// ValidateStruct validates a struct against its validation tags.
func ValidateStruct(s any) error {
    if err := validate.Struct(s); err != nil {
        var errs []string
        if valErrors, ok := err.(validator.ValidationErrors); ok {
            for _, fe := range valErrors {
                errs = append(errs, fe.Field()+": "+fe.Tag())
            }
            return errors.New(strings.Join(errs, "; "))
        }
        return err
    }
    return nil
}

// BindAndValidate parses the JSON request body into dst and validates struct tags.
func BindAndValidate(c *fiber.Ctx, dst any) error {
    if err := c.BodyParser(dst); err != nil {
        return ErrBadRequest(c, generic.ErrMsgInvalidRequestBody, err)
    }
    if err := ValidateStruct(dst); err != nil {
        return ErrValidation(c, err.Error(), err)
    }
    return nil
}
