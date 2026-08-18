package httpx

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()
	// Report validation errors using each field's `json` tag so they line
	// up with the zod schema field names on the frontend, not the Go
	// struct field names.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	return v
}

// BindAndValidate parses the request body into dst (a typed DTO struct)
// and validates it against its `validate` struct tags. Every handler that
// accepts a body must go through this — no untyped/unvalidated binding.
func BindAndValidate(c *fiber.Ctx, dst any) error {
	if err := c.BodyParser(dst); err != nil {
		return NewError(fiber.StatusBadRequest, "bad_request", "malformed request body")
	}
	if err := validate.Struct(dst); err != nil {
		if fieldErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			fields := make(map[string]string, len(fieldErrs))
			for _, fe := range fieldErrs {
				fields[fe.Field()] = fe.Tag()
			}
			return ErrValidation(fields)
		}
		return ErrInternal()
	}
	return nil
}
