package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type basicConfig struct {
	Host    string `env:"HOST" default:"localhost"`
	Port    int    `env:"PORT" default:"8080"`
	Debug   bool   `env:"DEBUG" default:"false"`
	Workers uint   `env:"WORKERS" default:"4"`
	Rate    float64 `env:"RATE" default:"1.5"`
}

type configWithDuration struct {
	Timeout time.Duration `env:"TIMEOUT" default:"30s"`
	Retry   time.Duration `env:"RETRY" default:"5m"`
}

type configWithSlice struct {
	Tags []string `env:"TAGS" default:"a, b, c"`
}

type nestedConfig struct {
	App string `env:"APP_NAME" default:"myapp"`
	DB  dbConfig `env_prefix:"DB"`
}

type dbConfig struct {
	Host string `env:"HOST" default:"localhost"`
	Port int    `env:"PORT" default:"5432"`
}

type configNoTags struct {
	Name    string
	Ignored string
}

func TestLoad_Defaults(t *testing.T) {
	var cfg basicConfig
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, false, cfg.Debug)
	assert.Equal(t, uint(4), cfg.Workers)
	assert.Equal(t, 1.5, cfg.Rate)
}

func TestLoad_EnvOverrides(t *testing.T) {
	os.Setenv("HOST", "0.0.0.0")
	os.Setenv("PORT", "9090")
	os.Setenv("DEBUG", "true")
	os.Setenv("WORKERS", "8")
	os.Setenv("RATE", "2.5")
	defer func() {
		os.Unsetenv("HOST")
		os.Unsetenv("PORT")
		os.Unsetenv("DEBUG")
		os.Unsetenv("WORKERS")
		os.Unsetenv("RATE")
	}()

	var cfg basicConfig
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, true, cfg.Debug)
	assert.Equal(t, uint(8), cfg.Workers)
	assert.Equal(t, 2.5, cfg.Rate)
}

func TestLoad_Duration(t *testing.T) {
	var cfg configWithDuration
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, 30*time.Second, cfg.Timeout)
	assert.Equal(t, 5*time.Minute, cfg.Retry)
}

func TestLoad_DurationFromEnv(t *testing.T) {
	os.Setenv("TIMEOUT", "10s")
	os.Setenv("RETRY", "1m")
	defer func() {
		os.Unsetenv("TIMEOUT")
		os.Unsetenv("RETRY")
	}()

	var cfg configWithDuration
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, cfg.Timeout)
	assert.Equal(t, 1*time.Minute, cfg.Retry)
}

func TestLoad_Slice(t *testing.T) {
	var cfg configWithSlice
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, cfg.Tags)
}

func TestLoad_SliceFromEnv(t *testing.T) {
	os.Setenv("TAGS", "x, y, z")
	defer os.Unsetenv("TAGS")

	var cfg configWithSlice
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, []string{"x", "y", "z"}, cfg.Tags)
}

func TestLoad_NestedStruct(t *testing.T) {
	var cfg nestedConfig
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "myapp", cfg.App)
	assert.Equal(t, "localhost", cfg.DB.Host)
	assert.Equal(t, 5432, cfg.DB.Port)
}

func TestLoad_NestedStructFromEnv(t *testing.T) {
	os.Setenv("DB_HOST", "db.example.com")
	os.Setenv("DB_PORT", "3306")
	defer func() {
		os.Unsetenv("DB_HOST")
		os.Unsetenv("DB_PORT")
	}()

	var cfg nestedConfig
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "db.example.com", cfg.DB.Host)
	assert.Equal(t, 3306, cfg.DB.Port)
}

func TestLoadWithPrefix(t *testing.T) {
	os.Setenv("MYAPP_HOST", "prefixed.host")
	os.Setenv("MYAPP_PORT", "3000")
	defer func() {
		os.Unsetenv("MYAPP_HOST")
		os.Unsetenv("MYAPP_PORT")
	}()

	var cfg basicConfig
	err := LoadWithPrefix("MYAPP_", &cfg)
	require.NoError(t, err)
	assert.Equal(t, "prefixed.host", cfg.Host)
	assert.Equal(t, 3000, cfg.Port)
}

func TestLoad_NonPointerError(t *testing.T) {
	var cfg basicConfig
	err := Load(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pointer to struct")
}

func TestLoad_NonStructPointerError(t *testing.T) {
	s := "hello"
	err := Load(&s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pointer to struct")
}

func TestLoad_InvalidIntValue(t *testing.T) {
	os.Setenv("PORT", "not_a_number")
	defer os.Unsetenv("PORT")

	var cfg basicConfig
	err := Load(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
}

func TestLoad_InvalidBoolValue(t *testing.T) {
	os.Setenv("DEBUG", "not_a_bool")
	defer os.Unsetenv("DEBUG")

	var cfg basicConfig
	err := Load(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
}

func TestLoad_NoTags(t *testing.T) {
	var cfg configNoTags
	err := Load(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "", cfg.Name)
}

func TestLoad_InvalidFloatValue(t *testing.T) {
	os.Setenv("RATE", "not_a_float")
	defer os.Unsetenv("RATE")

	var cfg basicConfig
	err := Load(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
}

func TestLoad_InvalidUintValue(t *testing.T) {
	os.Setenv("WORKERS", "-1")
	defer os.Unsetenv("WORKERS")

	var cfg basicConfig
	err := Load(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
}

func TestLoad_InvalidDurationValue(t *testing.T) {
	os.Setenv("TIMEOUT", "not_a_duration")
	defer os.Unsetenv("TIMEOUT")

	var cfg configWithDuration
	err := Load(&cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set field")
}
