package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Debug    bool   `json:"debug"`
	Database dbCfg  `json:"database"`
}

type dbCfg struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
}

func TestNew(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
	assert.NotNil(t, c.data)
	assert.Empty(t, c.filePath)
	assert.Empty(t, c.envPrefix)
}

func TestNewWithOptions(t *testing.T) {
	c := New(WithFile("/etc/app.json"), WithEnvPrefix("MYAPP"))
	assert.Equal(t, "/etc/app.json", c.filePath)
	assert.Equal(t, "MYAPP", c.envPrefix)
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	content := `{
		"host": "example.com",
		"port": 9090,
		"debug": true,
		"database": {
			"driver": "postgres",
			"dsn": "postgres://localhost/test"
		}
	}`
	err := os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)

	var cfg testConfig
	err = LoadFile(path, &cfg)
	require.NoError(t, err)
	assert.Equal(t, "example.com", cfg.Host)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, true, cfg.Debug)
	assert.Equal(t, "postgres", cfg.Database.Driver)
	assert.Equal(t, "postgres://localhost/test", cfg.Database.DSN)
}

func TestLoadFile_NotFound(t *testing.T) {
	var cfg testConfig
	err := LoadFile("/nonexistent/path/config.json", &cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	err := os.WriteFile(path, []byte("not json"), 0644)
	require.NoError(t, err)

	var cfg testConfig
	err = LoadFile(path, &cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestSaveFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.json")

	cfg := testConfig{
		Host:  "localhost",
		Port:  8080,
		Debug: false,
		Database: dbCfg{
			Driver: "sqlite",
			DSN:    "file:test.db",
		},
	}

	err := SaveFile(path, cfg)
	require.NoError(t, err)

	// Verify file exists and has correct permissions
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

	// Verify content
	var loaded testConfig
	err = LoadFile(path, &loaded)
	require.NoError(t, err)
	assert.Equal(t, cfg, loaded)
}

func TestSaveFile_CreatesDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "config.json")

	cfg := testConfig{Host: "test"}
	err := SaveFile(path, cfg)
	require.NoError(t, err)

	var loaded testConfig
	err = LoadFile(path, &loaded)
	require.NoError(t, err)
	assert.Equal(t, "test", loaded.Host)
}

func TestLoadOrCreate_CreatesNew(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	defaults := testConfig{
		Host:  "default.host",
		Port:  3000,
		Debug: true,
		Database: dbCfg{
			Driver: "sqlite",
			DSN:    "file:default.db",
		},
	}

	var cfg testConfig
	err := LoadOrCreate(path, &cfg, defaults)
	require.NoError(t, err)
	assert.Equal(t, "default.host", cfg.Host)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, true, cfg.Debug)
	assert.Equal(t, "sqlite", cfg.Database.Driver)

	// Verify the file was created
	_, err = os.Stat(path)
	assert.NoError(t, err)

	// Verify loading the created file gives same values
	var loaded testConfig
	err = LoadFile(path, &loaded)
	require.NoError(t, err)
	assert.Equal(t, cfg, loaded)
}

func TestLoadOrCreate_LoadsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	existing := testConfig{
		Host: "existing.host",
		Port: 5000,
	}
	err := SaveFile(path, existing)
	require.NoError(t, err)

	defaults := testConfig{
		Host: "default.host",
		Port: 3000,
	}

	var cfg testConfig
	err = LoadOrCreate(path, &cfg, defaults)
	require.NoError(t, err)
	assert.Equal(t, "existing.host", cfg.Host)
	assert.Equal(t, 5000, cfg.Port)
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "roundtrip.json")

	original := testConfig{
		Host:  "roundtrip.test",
		Port:  7777,
		Debug: true,
		Database: dbCfg{
			Driver: "mysql",
			DSN:    "mysql://root@localhost/db",
		},
	}

	err := SaveFile(path, original)
	require.NoError(t, err)

	var loaded testConfig
	err = LoadFile(path, &loaded)
	require.NoError(t, err)
	assert.Equal(t, original, loaded)
}
