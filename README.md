# digital.vasic.config

A Go configuration management library with support for JSON file-based configuration, environment variable binding, and validation.

## Packages

### pkg/config

Core configuration file management with JSON serialization.

```go
import "digital.vasic.config/pkg/config"

// Load from file
var cfg MyConfig
err := config.LoadFile("config.json", &cfg)

// Save to file
err := config.SaveFile("config.json", cfg)

// Load or create with defaults
defaults := MyConfig{Host: "localhost", Port: 8080}
var cfg MyConfig
err := config.LoadOrCreate("config.json", &cfg, defaults)
```

### pkg/env

Environment variable loading with struct tag binding and type conversion.

```go
import "digital.vasic.config/pkg/env"

type ServerConfig struct {
    Host    string        `env:"HOST" default:"localhost"`
    Port    int           `env:"PORT" default:"8080"`
    Debug   bool          `env:"DEBUG" default:"false"`
    Timeout time.Duration `env:"TIMEOUT" default:"30s"`
    Tags    []string      `env:"TAGS" default:"a,b,c"`
}

var cfg ServerConfig
err := env.Load(&cfg)

// With prefix (reads MYAPP_HOST, MYAPP_PORT, etc.)
err := env.LoadWithPrefix("MYAPP_", &cfg)
```

Supported types: `string`, `int*`, `uint*`, `float*`, `bool`, `time.Duration`, `[]string`.

Nested structs are supported via the `env_prefix` tag:

```go
type Config struct {
    DB DatabaseConfig `env_prefix:"DB"`
}

type DatabaseConfig struct {
    Host string `env:"HOST" default:"localhost"`
    Port int    `env:"PORT" default:"5432"`
}
// Reads DB_HOST and DB_PORT
```

### pkg/validator

Composable validation rules with multi-error collection.

```go
import "digital.vasic.config/pkg/validator"

errs := validator.Validate(
    validator.ValidationField{
        Value: cfg.Host,
        Rules: []validator.Rule{
            validator.Required("host"),
            validator.MinLength("host", 1),
        },
    },
    validator.ValidationField{
        Value: cfg.Port,
        Rules: []validator.Rule{
            validator.Required("port"),
            validator.Range("port", 1, 65535),
        },
    },
    validator.ValidationField{
        Value: cfg.Mode,
        Rules: []validator.Rule{
            validator.OneOf("mode", "debug", "release", "test"),
        },
    },
)
if len(errs) > 0 {
    // Handle validation errors
}
```

## Installation

```bash
go get digital.vasic.config
```

## Testing

```bash
go test ./... -count=1
```

## License

Copyright (c) Milos Vasic. All rights reserved.
