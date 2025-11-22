package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func NormalizeValidationErrors(err error, req interface{}) []string {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []string{"validation failed"}
	}

	var errorsList []string
	reqType := reflect.TypeOf(req)
	if reqType.Kind() == reflect.Ptr {
		reqType = reqType.Elem()
	}

	for _, fieldErr := range validationErrors {
		jsonName := fieldErr.Field()

		if field, found := reqType.FieldByName(fieldErr.StructField()); found {
			if jsonTag := field.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
				if commaIndex := strings.Index(jsonTag, ","); commaIndex > 0 {
					jsonName = jsonTag[:commaIndex]
				} else {
					jsonName = jsonTag
				}
			}
		}

		var message string
		switch fieldErr.Tag() {
		case "required", "required_without":
			message = jsonName + " is required"
		case "min":
			message = jsonName + " is too short"
		case "max":
			message = jsonName + " is too long"
		default:
			message = jsonName + " validation failed"
		}
		errorsList = append(errorsList, message)
	}

	return errorsList
}
