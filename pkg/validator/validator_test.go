package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequired_NonZero(t *testing.T) {
	rule := Required("name")
	err := rule("hello")
	assert.NoError(t, err)
}

func TestRequired_Zero(t *testing.T) {
	rule := Required("name")
	err := rule("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestRequired_ZeroInt(t *testing.T) {
	rule := Required("count")
	err := rule(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "count is required")
}

func TestRequired_NonZeroInt(t *testing.T) {
	rule := Required("count")
	err := rule(42)
	assert.NoError(t, err)
}

func TestMinLength_Valid(t *testing.T) {
	rule := MinLength("password", 8)
	err := rule("longpassword")
	assert.NoError(t, err)
}

func TestMinLength_TooShort(t *testing.T) {
	rule := MinLength("password", 8)
	err := rule("short")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 8 characters")
}

func TestMinLength_ExactLength(t *testing.T) {
	rule := MinLength("password", 5)
	err := rule("hello")
	assert.NoError(t, err)
}

func TestMinLength_NotString(t *testing.T) {
	rule := MinLength("password", 8)
	err := rule(12345)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

func TestRange_Valid(t *testing.T) {
	rule := Range("port", 1, 65535)
	err := rule(8080)
	assert.NoError(t, err)
}

func TestRange_TooLow(t *testing.T) {
	rule := Range("port", 1, 65535)
	err := rule(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 65535")
}

func TestRange_TooHigh(t *testing.T) {
	rule := Range("port", 1, 65535)
	err := rule(70000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be between 1 and 65535")
}

func TestRange_AtMin(t *testing.T) {
	rule := Range("port", 1, 65535)
	err := rule(1)
	assert.NoError(t, err)
}

func TestRange_AtMax(t *testing.T) {
	rule := Range("port", 1, 65535)
	err := rule(65535)
	assert.NoError(t, err)
}

func TestRange_NotInt(t *testing.T) {
	rule := Range("port", 1, 65535)
	err := rule("not_an_int")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be an integer")
}

func TestOneOf_Valid(t *testing.T) {
	rule := OneOf("mode", "debug", "release", "test")
	err := rule("debug")
	assert.NoError(t, err)
}

func TestOneOf_Invalid(t *testing.T) {
	rule := OneOf("mode", "debug", "release", "test")
	err := rule("production")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be one of: debug, release, test")
}

func TestOneOf_NotString(t *testing.T) {
	rule := OneOf("mode", "debug", "release")
	err := rule(42)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

func TestValidate_AllPass(t *testing.T) {
	errs := Validate(
		ValidationField{
			Value: "localhost",
			Rules: []Rule{Required("host")},
		},
		ValidationField{
			Value: 8080,
			Rules: []Rule{Range("port", 1, 65535)},
		},
	)
	assert.Empty(t, errs)
}

func TestValidate_SomeFailures(t *testing.T) {
	errs := Validate(
		ValidationField{
			Value: "",
			Rules: []Rule{Required("host")},
		},
		ValidationField{
			Value: 8080,
			Rules: []Rule{Range("port", 1, 65535)},
		},
		ValidationField{
			Value: "invalid",
			Rules: []Rule{OneOf("mode", "debug", "release")},
		},
	)
	assert.Len(t, errs, 2)
}

func TestValidate_MultipleRulesPerField(t *testing.T) {
	errs := Validate(
		ValidationField{
			Value: "ab",
			Rules: []Rule{
				Required("password"),
				MinLength("password", 8),
			},
		},
	)
	// Required passes (non-zero), MinLength fails
	assert.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), "at least 8 characters")
}

func TestValidate_Empty(t *testing.T) {
	errs := Validate()
	assert.Empty(t, errs)
}

func TestValidate_AllFail(t *testing.T) {
	errs := Validate(
		ValidationField{
			Value: "",
			Rules: []Rule{Required("host"), MinLength("host", 1)},
		},
		ValidationField{
			Value: 0,
			Rules: []Rule{Required("port"), Range("port", 1, 65535)},
		},
	)
	assert.Len(t, errs, 4)
}
