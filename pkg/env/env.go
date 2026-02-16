// Package env provides environment variable configuration loading
// with struct tag-based binding and type conversion.
package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Load populates a struct from environment variables using `env` struct tags.
// Example: `env:"PORT" default:"8080"`
func Load(target interface{}) error {
	return load(target, "")
}

// LoadWithPrefix populates a struct with a prefix added to env var names.
func LoadWithPrefix(prefix string, target interface{}) error {
	return load(target, prefix)
}

func load(target interface{}, prefix string) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to struct")
	}
	return loadStruct(v.Elem(), prefix)
}

func loadStruct(v reflect.Value, prefix string) error {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Handle nested structs
		if field.Type.Kind() == reflect.Struct && field.Type != reflect.TypeOf(time.Time{}) {
			nestedPrefix := prefix
			if tag := field.Tag.Get("env_prefix"); tag != "" {
				nestedPrefix = prefix + tag + "_"
			}
			if err := loadStruct(fieldVal, nestedPrefix); err != nil {
				return err
			}
			continue
		}

		envTag := field.Tag.Get("env")
		if envTag == "" {
			continue
		}

		envKey := prefix + envTag
		defaultVal := field.Tag.Get("default")

		val := os.Getenv(envKey)
		if val == "" {
			val = defaultVal
		}
		if val == "" {
			continue
		}

		if err := setField(fieldVal, val); err != nil {
			return fmt.Errorf("failed to set field %s from env %s: %w", field.Name, envKey, err)
		}
	}
	return nil
}

func setField(v reflect.Value, val string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			v.Set(reflect.ValueOf(d))
		} else {
			i, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return err
			}
			v.SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(val, ",")
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			v.Set(reflect.ValueOf(parts))
		}
	default:
		return fmt.Errorf("unsupported type: %s", v.Kind())
	}
	return nil
}
