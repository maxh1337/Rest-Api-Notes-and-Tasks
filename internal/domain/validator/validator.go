package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"rest-api-notes/internal/domain/entities"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewValidator() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	cv.validator.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		matched, _ := regexp.MatchString(`^\+[0-9]{10,15}$`, phone)
		return matched
	})

	if err := cv.validator.Struct(i); err != nil {
		validationErrors := make(map[string]string)

		for _, err := range err.(validator.ValidationErrors) {
			field := getJSONFieldName(i, err.Field())
			validationErrors[field] = getValidationErrorMessage(err)
		}

		return entities.NewValidationError(validationErrors)
	}
	return nil
}

func getJSONFieldName(s interface{}, fieldName string) string {
	rt := reflect.TypeOf(s)
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	field, found := rt.FieldByName(fieldName)
	if !found {
		return strings.ToLower(fieldName)
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return strings.ToLower(fieldName)
	}

	if idx := strings.Index(jsonTag, ","); idx != -1 {
		jsonTag = jsonTag[:idx]
	}

	return jsonTag
}

func getValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Minimum length is %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Maximum length is %s characters", fe.Param())
	case "alphanum":
		return "Only alphanumeric characters are allowed"
	case "uuid":
		return "Invalid UUID format"
	case "len":
		return fmt.Sprintf("Length must be exactly %s characters", fe.Param())
	case "numeric":
		return "Only numeric characters are allowed"
	case "phone":
		return "Invalid phone number format"
	default:
		return fmt.Sprintf("Invalid value for %s", fe.Field())
	}
}
