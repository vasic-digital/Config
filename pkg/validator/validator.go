// Package validator provides configuration validation utilities.
package validator

import (
	"fmt"
	"reflect"
	"strings"
)

// Rule is a validation rule function.
type Rule func(value interface{}) error

// Required returns a rule that checks the value is non-zero.
func Required(fieldName string) Rule {
	return func(value interface{}) error {
		if reflect.ValueOf(value).IsZero() {
			return fmt.Errorf("%s is required", fieldName)
		}
		return nil
	}
}

// MinLength returns a rule that checks string minimum length.
func MinLength(fieldName string, min int) Rule {
	return func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string", fieldName)
		}
		if len(s) < min {
			return fmt.Errorf("%s must be at least %d characters", fieldName, min)
		}
		return nil
	}
}

// Range returns a rule that checks an int is within range.
func Range(fieldName string, min, max int) Rule {
	return func(value interface{}) error {
		v := reflect.ValueOf(value)
		var i int64
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i = v.Int()
		default:
			return fmt.Errorf("%s must be an integer", fieldName)
		}
		if i < int64(min) || i > int64(max) {
			return fmt.Errorf("%s must be between %d and %d", fieldName, min, max)
		}
		return nil
	}
}

// OneOf returns a rule that checks the value is one of the allowed values.
func OneOf(fieldName string, allowed ...string) Rule {
	return func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			return fmt.Errorf("%s must be a string", fieldName)
		}
		for _, a := range allowed {
			if s == a {
				return nil
			}
		}
		return fmt.Errorf("%s must be one of: %s", fieldName, strings.Join(allowed, ", "))
	}
}

// ValidationField pairs a value with its rules.
type ValidationField struct {
	Value interface{}
	Rules []Rule
}

// Validate runs all rules and collects errors.
func Validate(fields ...ValidationField) []error {
	var errs []error
	for _, f := range fields {
		for _, rule := range f.Rules {
			if err := rule(f.Value); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errs
}
