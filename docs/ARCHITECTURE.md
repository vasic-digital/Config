# Architecture

## Purpose

`digital.vasic.config` is a Go library for configuration management that supports JSON file loading/saving, struct-based environment variable binding with type conversion, and composable validation rules with multi-error collection.

## Package Overview

| Package | Path | Description |
|---------|------|-------------|
| `config` | `pkg/config/` | Core configuration management: JSON file I/O, load-or-create with defaults, functional options |
| `env` | `pkg/env/` | Struct tag-based environment variable loading with prefix support and automatic type conversion |
| `validator` | `pkg/validator/` | Composable validation rule functions with batch execution and multi-error aggregation |

## Design Patterns

| Pattern | Where | Description |
|---------|-------|-------------|
| Functional Options | `config.New()` | `Option` functions (`WithFile`, `WithEnvPrefix`) configure `Config` at construction time |
| Struct Tag Binding | `env.Load()` | `env` and `default` struct tags drive environment variable mapping and fallback values |
| Nested Prefix Composition | `env.loadStruct()` | `env_prefix` tag on struct fields composes hierarchical environment variable names (e.g., `DB_HOST`) |
| Reflection-Based Type Conversion | `env.setField()` | Automatically converts string env values to string, int, uint, float, bool, `time.Duration`, and `[]string` |
| Composable Rules | `validator.Rule` | Each rule is a standalone function; multiple rules compose per field via `ValidationField.Rules` |
| Multi-Error Collection | `validator.Validate()` | Runs all rules across all fields, collecting every error rather than failing on the first |
| Load-or-Create | `config.LoadOrCreate()` | Loads existing config file or initializes from defaults via JSON round-trip, then persists |

## Key Interfaces and Types

### config

```go
// Loader loads configuration from various sources.
type Loader interface {
    Load(target interface{}) error
}

// Validator validates configuration values.
type Validator interface {
    Validate() error
}

// Option configures a Config instance.
type Option func(*Config)

func New(opts ...Option) *Config
func LoadFile(path string, target interface{}) error
func SaveFile(path string, config interface{}) error
func LoadOrCreate(path string, target interface{}, defaults interface{}) error
```

### env

```go
func Load(target interface{}) error
func LoadWithPrefix(prefix string, target interface{}) error
```

Supported struct tags:
- `env:"VAR_NAME"` -- maps the field to an environment variable
- `default:"value"` -- fallback when the environment variable is unset
- `env_prefix:"PREFIX"` -- on nested struct fields, prepends prefix to child env var names

Supported types: `string`, `int*`, `uint*`, `float*`, `bool`, `time.Duration`, `[]string` (comma-separated).

### validator

```go
// Rule is a validation rule function.
type Rule func(value interface{}) error

// ValidationField pairs a value with its rules.
type ValidationField struct {
    Value interface{}
    Rules []Rule
}

func Required(fieldName string) Rule
func MinLength(fieldName string, min int) Rule
func Range(fieldName string, min, max int) Rule
func OneOf(fieldName string, allowed ...string) Rule
func Validate(fields ...ValidationField) []error
```

## Usage Example

```go
package main

import (
    "fmt"
    "log"
    "time"

    "digital.vasic.config/pkg/config"
    "digital.vasic.config/pkg/env"
    "digital.vasic.config/pkg/validator"
)

type AppConfig struct {
    Host    string        `json:"host"    env:"HOST"    default:"localhost"`
    Port    int           `json:"port"    env:"PORT"    default:"8080"`
    Mode    string        `json:"mode"    env:"MODE"    default:"debug"`
    Timeout time.Duration `json:"timeout" env:"TIMEOUT" default:"30s"`
    DB      DBConfig      `json:"db"      env_prefix:"DB"`
}

type DBConfig struct {
    Host   string `json:"host"   env:"HOST"   default:"localhost"`
    Port   int    `json:"port"   env:"PORT"   default:"5432"`
    Driver string `json:"driver" env:"DRIVER" default:"postgres"`
}

func main() {
    // 1. Load from file (or create with defaults)
    var cfg AppConfig
    defaults := AppConfig{Host: "0.0.0.0", Port: 8080, Mode: "release"}
    if err := config.LoadOrCreate("config.json", &cfg, defaults); err != nil {
        log.Fatal(err)
    }

    // 2. Override with environment variables
    if err := env.Load(&cfg); err != nil {
        log.Fatal(err)
    }

    // 3. Validate
    errs := validator.Validate(
        validator.ValidationField{
            Value: cfg.Host,
            Rules: []validator.Rule{validator.Required("host")},
        },
        validator.ValidationField{
            Value: cfg.Port,
            Rules: []validator.Rule{validator.Range("port", 1, 65535)},
        },
        validator.ValidationField{
            Value: cfg.Mode,
            Rules: []validator.Rule{validator.OneOf("mode", "debug", "release", "test")},
        },
    )
    if len(errs) > 0 {
        for _, e := range errs {
            fmt.Println("validation error:", e)
        }
        log.Fatal("invalid configuration")
    }

    fmt.Printf("Starting %s server on %s:%d\n", cfg.Mode, cfg.Host, cfg.Port)
}
```
