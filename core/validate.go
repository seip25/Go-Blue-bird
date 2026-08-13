package core

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ErrorMap map[string]string

func FormatValidationError(err error, obj interface{}) ErrorMap {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		out := make(ErrorMap, len(ve))
		for _, fe := range ve {
			fieldName := getJSONFieldName(obj, fe.StructField())
			out[fieldName] = formatFieldError(fe)
		}
		return out
	}
	return ErrorMap{"error": err.Error()}
}

func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email address format"
	case "min":
		return fmt.Sprintf("Must be at least %s characters or value", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters or value", fe.Param())
	case "eqfield":
		return fmt.Sprintf("Must match field %s", fe.Param())
	default:
		return fmt.Sprintf("Failed validation rule '%s'", fe.Tag())
	}
}

func getJSONFieldName(obj interface{}, fieldName string) string {
	t := reflect.TypeOf(obj)
	if t == nil {
		return strings.ToLower(fieldName)
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return strings.ToLower(fieldName)
	}

	field, found := t.FieldByName(fieldName)
	if !found {
		return strings.ToLower(fieldName)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return strings.ToLower(fieldName)
	}

	parts := strings.Split(jsonTag, ",")
	return parts[0]
}

func BindAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBind(obj); err != nil {
		formattedErrors := FormatValidationError(err, obj)
		RespondValidationError(c, formattedErrors)
		return false
	}
	return true
}

func BindJSONAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		formattedErrors := FormatValidationError(err, obj)
		RespondValidationError(c, formattedErrors)
		return false
	}
	return true
}
