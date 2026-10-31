package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Error struct {
	Fields map[string]any
}

func (e *Error) Error() string {
	return "validation failed"
}

var validate *validator.Validate

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

func Validate(structure any) *Error {
	err := validate.Struct(structure)
	if err == nil {
		return nil
	}

	var valErrors validator.ValidationErrors
	if errors.As(err, &valErrors) {
		fields := make(map[string]any, len(valErrors))
		for _, fe := range valErrors {
			fields[fe.Field()] = msgForTag(fe)
		}
		return &Error{Fields: fields}
	}

	return &Error{
		Fields: map[string]any{
			"error": err.Error(),
		},
	}
}

func msgForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address format"
	case "gte", "min":
		return fmt.Sprintf("Must be at least %s characters long", fe.Param())
	case "lte", "max":
		return fmt.Sprintf("Must be at most %s characters long", fe.Param())
	case "len":
		return fmt.Sprintf("Must be exactly %s characters long", fe.Param())
	case "uuid":
		return "Must be a valid UUID"
	default:
		return fmt.Sprintf("Failed on validation rule '%s'", fe.Tag())
	}
}
