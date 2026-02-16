// Package config provides a generic configuration management library
// with support for file-based, environment variable, and programmatic
// configuration with validation.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Loader loads configuration from various sources.
type Loader interface {
	Load(target interface{}) error
}

// Validator validates configuration values.
type Validator interface {
	Validate() error
}

// Config provides generic configuration management.
type Config struct {
	filePath  string
	envPrefix string
	data      map[string]interface{}
}

// New creates a new Config instance.
func New(opts ...Option) *Config {
	c := &Config{
		data: make(map[string]interface{}),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option configures a Config instance.
type Option func(*Config)

// WithFile sets the configuration file path.
func WithFile(path string) Option {
	return func(c *Config) {
		c.filePath = path
	}
}

// WithEnvPrefix sets the environment variable prefix.
func WithEnvPrefix(prefix string) Option {
	return func(c *Config) {
		c.envPrefix = prefix
	}
}

// LoadFile loads configuration from a JSON file into target struct.
func LoadFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	return nil
}

// SaveFile saves configuration to a JSON file.
func SaveFile(path string, config interface{}) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

// LoadOrCreate loads config from file, or creates with defaults if missing.
func LoadOrCreate(path string, target interface{}, defaults interface{}) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Copy defaults to target via JSON round-trip
		data, err := json.Marshal(defaults)
		if err != nil {
			return fmt.Errorf("failed to marshal defaults: %w", err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to apply defaults: %w", err)
		}
		return SaveFile(path, target)
	}
	return LoadFile(path, target)
}
