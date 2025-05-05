package validator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Define validation error types
var (
	ErrRequired   = errors.New("field is required")
	ErrMinLength  = errors.New("field is too short")
	ErrMaxLength  = errors.New("field is too long")
	ErrEmail      = errors.New("invalid email format")
	ErrInvalid    = errors.New("invalid value")
	ErrValidation = errors.New("validation error")
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

// Validator handles data validation
type Validator struct {
	Errors map[string]string
}

// New creates a new validator
func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}

// Valid returns true if there are no errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error for a field
func (v *Validator) AddError(field, message string) {
	if _, exists := v.Errors[field]; !exists {
		v.Errors[field] = message
	}
}

// Check adds an error if the condition is false
func (v *Validator) Check(ok bool, field, message string) {
	if !ok {
		v.AddError(field, message)
	}
}

// In returns true if a value is in a list of allowed values
func (v *Validator) In(value string, list ...string) bool {
	for _, item := range list {
		if value == item {
			return true
		}
	}
	return false
}

// Required checks if a value is not empty
func (v *Validator) Required(value string, field string) {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, fmt.Errorf("%w: %s", ErrRequired, field).Error())
	}
}

// MinLength checks if a value meets the minimum length
func (v *Validator) MinLength(value string, min int, field string) {
	if len(value) < min {
		v.AddError(field, fmt.Errorf("%w: %s must be at least %d characters", ErrMinLength, field, min).Error())
	}
}

// MaxLength checks if a value doesn't exceed the maximum length
func (v *Validator) MaxLength(value string, max int, field string) {
	if len(value) > max {
		v.AddError(field, fmt.Errorf("%w: %s must be at most %d characters", ErrMaxLength, field, max).Error())
	}
}

// Email checks if a value is a valid email
func (v *Validator) Email(value string, field string) {
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(pattern)
	if !regex.MatchString(value) {
		v.AddError(field, fmt.Errorf("%w: %s", ErrEmail, field).Error())
	}
}

// ValidateStruct validates a struct using validate tags
func (v *Validator) ValidateStruct(s interface{}) error {
	val := reflect.ValueOf(s)

	// If pointer, get the underlying value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Only handle structs
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validation requires a struct or a pointer to a struct")
	}

	typ := val.Type()

	// Iterate over struct fields
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		// Get validation tag
		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		// Get field name (use json tag if available)
		fieldName := fieldType.Name
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		// Only handle string fields for now
		if field.Kind() != reflect.String {
			continue
		}

		// Get field value
		value := field.String()

		// Process validation rules
		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			parts := strings.Split(rule, "=")
			ruleName := parts[0]

			switch ruleName {
			case "required":
				v.Required(value, fieldName)
			case "min":
				if len(parts) < 2 {
					continue
				}
				min := 0
				fmt.Sscanf(parts[1], "%d", &min)
				v.MinLength(value, min, fieldName)
			case "max":
				if len(parts) < 2 {
					continue
				}
				max := 0
				fmt.Sscanf(parts[1], "%d", &max)
				v.MaxLength(value, max, fieldName)
			case "email":
				v.Email(value, fieldName)
			}
		}
	}

	if !v.Valid() {
		return ErrValidation
	}

	return nil
}

// GetErrorMessages returns all error messages
func (v *Validator) GetErrorMessages() map[string]string {
	return v.Errors
}