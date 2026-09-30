# gocfg — Detailed Roadmap

> A small, type-safe configuration library for Go applications.

Repository:

```text
github.com/tidjee-dev/gocfg
```

---

## 1. Vision

`gocfg` provides a simple configuration convention for Go applications.

The core idea:

- Configuration is written in **Go**
- Environment-specific values come from `.env` and OS environment variables
- Configuration is **type-safe**
- `config/` is the application's configuration source
- `.env` and `.env.example` can be generated from configuration definitions
- A small CLI manages initialization, environment files, and validation
- No YAML/JSON/TOML configuration is required

Example application:

```text
my-app/
├── config/
│   ├── app.go
│   ├── database.go
│   └── config.go
├── .env
├── .env.example
├── go.mod
└── main.go
```

---

# 2. Goals

## Core goals

- [ ] Simple Go-first configuration
- [ ] Type-safe environment values
- [ ] Laravel-inspired `config/` organization
- [ ] `.env` support
- [ ] OS environment support
- [ ] Explicit configuration composition
- [ ] Clear configuration errors
- [ ] No global mutable configuration
- [ ] Minimal dependencies
- [ ] Good testability
- [ ] Useful CLI
- [ ] Safe `.env` generation
- [ ] `.env.example` generation
- [ ] Configuration validation

## CLI goals

The CLI should provide:

```bash
gocfg init
gocfg env
gocfg check
gocfg validate
gocfg version
```

The CLI uses:

- **Cobra** for commands and flags
- **Lipgloss** for terminal presentation

Cobra and Lipgloss must remain isolated from the core configuration library.

---

# 3. Non-Goals

`gocfg` should not become a generic configuration framework.

The following are explicitly out of scope initially:

- [ ] YAML
- [ ] JSON
- [ ] TOML
- [ ] XML
- [ ] Automatic discovery of arbitrary configuration files
- [ ] Remote configuration
- [ ] Configuration servers
- [ ] Kubernetes-specific configuration
- [ ] Secret-management services
- [ ] Dependency injection
- [ ] Global mutable configuration
- [ ] Dynamic `config.Get("foo.bar")`
- [ ] Runtime Go code evaluation
- [ ] Automatic reflection over arbitrary application structs
- [ ] PostgreSQL-specific integrations
- [ ] Framework-specific integrations

---

# 4. Configuration Philosophy

The application owns its configuration.

`gocfg` provides the primitives and conventions.

Example:

```go
package config

import "github.com/tidjee-dev/gocfg/env"

type AppConfig struct {
    Name  string
    Env   string
    Debug bool
    URL   string
}

func App() AppConfig {
    return AppConfig{
        Name:  env.String("APP_NAME", "My App"),
        Env:   env.String("APP_ENV", "dev"),
        Debug: env.Bool("APP_DEBUG", true),
        URL:   env.String("APP_URL", "http://localhost:9000"),
    }
}
```

The application composes configuration explicitly:

```go
package config

type Config struct {
    App      AppConfig
    Database DatabaseConfig
}

func Load() Config {
    return Config{
        App:      App(),
        Database: Database(),
    }
}
```

`gocfg` should **not** dynamically discover and execute `config/*.go`.

---

# 5. Environment Precedence

Configuration values follow this precedence:

```text
Go default
     ↓
.env
     ↓
OS environment
```

Therefore:

```text
OS environment > .env > Go default
```

Example:

```go
env.String("APP_PORT", "9000")
```

with:

```dotenv
APP_PORT=8000
```

produces:

```text
8000
```

If the operating system contains:

```bash
APP_PORT=7000
```

the result is:

```text
7000
```

---

# 6. Project Structure

Initial library structure:

```text
gocfg/
├── cmd/
│   └── gocfg/
│       └── main.go
│
├── env/
│   ├── string.go
│   ├── bool.go
│   ├── int.go
│   ├── int64.go
│   ├── float64.go
│   └── duration.go
│
├── internal/
│   ├── cli/
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── env.go
│   │   ├── check.go
│   │   ├── validate.go
│   │   └── version.go
│   │
│   ├── dotenv/
│   │   ├── parser.go
│   │   └── loader.go
│   │
│   └── scaffold/
│       ├── scaffold.go
│       └── templates/
│           ├── env.tmpl
│           ├── env.example.tmpl
│           ├── app.go.tmpl
│           └── config.go.tmpl
│
├── config.go
├── errors.go
├── go.mod
├── go.sum
├── README.md
├── CHANGELOG.md
├── CONTRIBUTING.md
├── LICENSE
└── examples/
    └── basic/
        ├── config/
        │   ├── app.go
        │   ├── database.go
        │   └── config.go
        ├── .env.example
        └── main.go
```

---

# 7. Dependencies

## Core

The core library should remain as dependency-light as possible.

Potentially:

```text
stdlib only
```

## CLI

The CLI uses:

```text
github.com/spf13/cobra
github.com/charmbracelet/lipgloss
```

Dependency boundary:

```text
cmd/gocfg
    │
    ├── Cobra
    ├── Lipgloss
    │
    ▼
internal/cli
    │
    ▼
gocfg core
    │
    ├── env
    ├── dotenv
    └── configuration
```

The core packages must not import Cobra or Lipgloss.

---

# 8. Environment API

Initial API:

```go
env.String(key, fallback)
env.Bool(key, fallback)
env.Int(key, fallback)
env.Int64(key, fallback)
env.Float64(key, fallback)
env.Duration(key, fallback)
```

Examples:

```go
port := env.Int("APP_PORT", 9000)

debug := env.Bool("APP_DEBUG", false)

timeout := env.Duration(
    "APP_TIMEOUT",
    30*time.Second,
)
```

---

# 9. Environment Parsing

## String

```go
env.String("APP_NAME", "My App")
```

Returns the environment value if present, otherwise the fallback.

## Boolean

Support conventional values such as:

```text
true
false
1
0
yes
no
```

The exact accepted values should be documented and tested.

## Integer

```go
env.Int("APP_PORT", 9000)
```

Invalid values should produce an error rather than silently using the fallback.

Example:

```text
invalid value for APP_PORT: "hello": expected integer
```

## Int64

```go
env.Int64("APP_MAX_SIZE", 100000)
```

## Float64

```go
env.Float64("APP_RATE", 1.5)
```

## Duration

```go
env.Duration("APP_TIMEOUT", 30*time.Second)
```

Example:

```dotenv
APP_TIMEOUT=30s
```

---

# 10. Error Handling

Configuration errors must be explicit.

Avoid:

```go
port, _ := strconv.Atoi(value)
```

Prefer contextual errors:

```text
invalid value for APP_PORT: "hello": expected integer
```

Errors should identify:

- environment variable
- supplied value where safe
- expected type/format

Secrets must not be included in error messages.

Example:

```text
invalid value for DB_PASSWORD
```

rather than:

```text
invalid value for DB_PASSWORD: "super-secret-password"
```

---

# 11. Required Values

Future API:

```go
env.Required("APP_KEY")
```

If missing:

```text
required environment variable APP_KEY is not set
```

Required values should be useful for:

- production secrets
- API keys
- database credentials
- signing keys

---

# 12. `.env` Support

`.env` should be optional.

Typical:

```dotenv
APP_NAME=My App
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

The loader should support:

- comments
- empty lines
- quoted values
- basic escaping
- whitespace handling
- duplicate-key rules
- environment precedence

---

# 13. `.env` Loading

Avoid implicit package-level side effects.

Do not use:

```go
func init() {
    loadDotEnv()
}
```

Instead, provide explicit loading.

Conceptually:

```go
err := gocfg.LoadEnv()
```

or:

```go
err := gocfg.LoadEnv(".env")
```

The exact API should be finalized during implementation.

---

# 14. Explicit `.env` Paths

Support custom paths:

```go
gocfg.LoadEnv(".env")
```

Potential future support:

```go
gocfg.LoadEnv("config/.env")
```

Do not make `.env` location assumptions part of the core configuration API.

---

# 15. Configuration Definitions

A major design goal is making `config/` the source of truth for environment configuration.

The CLI should eventually know which environment variables are used by the application.

Avoid trying to achieve this by parsing arbitrary Go AST/source code.

Instead, introduce explicit configuration metadata/definitions.

Possible direction:

```go
env.Define(...)
```

or:

```go
env.Group(...)
```

Example conceptual API:

```go
var App = env.Group("app", map[string]env.Definition{
    "APP_NAME":  env.String("My App"),
    "APP_ENV":   env.String("dev"),
    "APP_DEBUG": env.Bool(true),
    "APP_URL":   env.String("http://localhost:9000"),
})
```

The exact API is intentionally left open until the implementation phase.

Requirements:

- type-safe
- explicit
- easy to use from normal Go configuration
- usable by the CLI
- no arbitrary Go source parsing
- no reflection-heavy magic

---

# 16. Config-Driven Environment Generation

The CLI should be able to inspect configuration definitions and generate:

```text
.env
.env.example
```

Command:

```bash
gocfg env
```

Example configuration:

```go
env.String("APP_NAME", "My App")
env.String("APP_ENV", "dev")
env.Bool("APP_DEBUG", true)
env.String("APP_URL", "http://localhost:9000")
```

Produces:

```dotenv
APP_NAME=My App
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

---

# 17. `.env` Preservation

`gocfg env` must **not destroy existing environment values**.

Existing:

```dotenv
APP_NAME=My Production App
APP_ENV=prod
APP_DEBUG=false
```

Configuration later adds:

```go
env.String("APP_TIMEZONE", "Europe/Brussels")
```

Running:

```bash
gocfg env
```

should produce:

```dotenv
APP_NAME=My Production App
APP_ENV=prod
APP_DEBUG=false
APP_TIMEZONE=Europe/Brussels
```

Existing values remain untouched.

---

# 18. `.env.example` Generation

`.env.example` should represent the configuration schema.

Example:

```dotenv
APP_NAME=My App
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

Secrets should never be populated with real secret values.

For example:

```dotenv
DB_PASSWORD=
APP_SECRET=
API_KEY=
```

or another safe placeholder strategy.

---

# 19. `gocfg init`

`init` creates the initial project structure.

Command:

```bash
gocfg init
```

Creates:

```text
.env
.env.example
config/
├── app.go
└── config.go
```

Possible initial generated files:

### `.env`

```dotenv
APP_NAME=my-app
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

### `.env.example`

```dotenv
APP_NAME=my-app
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

### `config/app.go`

```go
package config

import "github.com/tidjee-dev/gocfg/env"

type AppConfig struct {
    Name  string
    Env   string
    Debug bool
    URL   string
}

func App() AppConfig {
    return AppConfig{
        Name:  env.String("APP_NAME", "My App"),
        Env:   env.String("APP_ENV", "dev"),
        Debug: env.Bool("APP_DEBUG", true),
        URL:   env.String("APP_URL", "http://localhost:9000"),
    }
}
```

### `config/config.go`

```go
package config

type Config struct {
    App AppConfig
}

func Load() Config {
    return Config{
        App: App(),
    }
}
```

The exact generated API will follow the final configuration-definition design.

---

# 20. `gocfg init` Behavior

Default behavior should be conservative.

If a file already exists:

```text
.env already exists — skipped
```

Do not overwrite it automatically.

Support:

```bash
gocfg init --force
```

to explicitly overwrite generated files.

Potential option:

```bash
gocfg init --dry-run
```

to show what would be created without modifying the filesystem.

Potential option:

```bash
gocfg init --name my-app
```

to define the initial application name.

---

# 21. Scaffold Templates

Use Go's `embed` package for CLI templates.

Structure:

```text
internal/scaffold/
├── scaffold.go
└── templates/
    ├── env.tmpl
    ├── env.example.tmpl
    ├── app.go.tmpl
    └── config.go.tmpl
```

The CLI binary should contain its templates and should not depend on files existing beside the executable.

---

# 22. File Permissions

Generated files should use appropriate permissions.

Recommended:

```text
.env            0600
.env.example    0644
config/*.go     0644
```

The exact behavior should account for existing files and platform differences.

---

# 23. CLI Architecture

Entry point:

```text
cmd/gocfg/main.go
```

Cobra root:

```text
internal/cli/root.go
```

Commands:

```text
internal/cli/
├── root.go
├── init.go
├── env.go
├── check.go
├── validate.go
└── version.go
```

The CLI should follow normal Cobra command composition.

Conceptually:

```text
gocfg
│
├── init
├── env
├── check
├── validate
└── version
```

---

# 24. Lipgloss Output

Lipgloss is responsible only for presentation.

Example output:

```text
gocfg

✓ created .env
✓ created .env.example
✓ created config/app.go
✓ created config/config.go

Project initialized.
```

For `check`:

```text
gocfg check

✓ .env found
✓ config/ found
✓ configuration definitions found
✓ environment is valid
```

Output should remain readable without relying on styling to understand the result.

---

# 25. `gocfg check`

Purpose:

```bash
gocfg check
```

Check the project configuration environment.

Potential checks:

- `.env` exists
- `.env.example` exists
- `config/` exists
- configuration definitions are valid
- required variables are present
- environment values can be parsed
- `.env` and configuration are consistent
- unexpected/missing variables can optionally be reported

Example:

```text
gocfg check

✓ configuration directory
✓ environment file
✓ environment example
✓ configuration definitions
✓ environment values

Configuration looks consistent.
```

`check` should primarily inspect and report.

---

# 26. `gocfg validate`

Purpose:

```bash
gocfg validate
```

Validation should evaluate actual configuration values.

Example:

```text
gocfg validate

✓ APP_ENV
✓ APP_DEBUG
✓ APP_PORT
✓ DATABASE_URL

Configuration is valid.
```

Invalid example:

```text
✗ APP_PORT: expected integer
✗ APP_ENV: unsupported value "productionn"

Configuration is invalid.
```

The command should return a non-zero exit status when validation fails.

---

# 27. `gocfg version`

Command:

```bash
gocfg version
```

Example:

```text
gocfg v0.1.0
```

The version should be injected during release builds where appropriate.

---

# 28. CLI Exit Codes

The CLI should use meaningful exit statuses.

Example:

```text
0 = success
1 = configuration/validation failure
2 = CLI usage error
```

Exact conventions should be finalized during implementation.

---

# 29. Configuration Validation

Separate configuration loading from application-specific validation.

Example:

```go
type Config struct {
    App      AppConfig
    Database DatabaseConfig
}

func (c Config) Validate() error {
    // application-specific validation
    return nil
}
```

This keeps `gocfg` responsible for parsing while allowing applications to define domain-specific rules.

---

# 30. Testing Strategy

Testing should be a major part of the project.

## Environment tests

- [ ] string fallback
- [ ] string environment value
- [ ] boolean parsing
- [ ] integer parsing
- [ ] int64 parsing
- [ ] float parsing
- [ ] duration parsing
- [ ] invalid values
- [ ] precedence

## `.env` tests

- [ ] empty file
- [ ] comments
- [ ] whitespace
- [ ] quoted values
- [ ] duplicate keys
- [ ] malformed entries
- [ ] precedence against OS environment
- [ ] custom path

## Configuration tests

- [ ] configuration definitions
- [ ] defaults
- [ ] required values
- [ ] validation
- [ ] invalid configuration

## CLI tests

Use `t.TempDir()` for filesystem isolation.

### `init`

- [ ] creates `.env`
- [ ] creates `.env.example`
- [ ] creates `config/`
- [ ] creates config files
- [ ] does not overwrite existing files
- [ ] `--force` overwrites
- [ ] `--dry-run` does not modify files

### `env`

- [ ] creates missing `.env`
- [ ] creates missing `.env.example`
- [ ] adds new configuration variables
- [ ] preserves existing `.env` values
- [ ] updates `.env.example`
- [ ] does not expose secrets
- [ ] handles malformed configuration

### `check`

- [ ] detects missing `.env`
- [ ] detects missing `config/`
- [ ] detects invalid definitions
- [ ] reports inconsistencies

### `validate`

- [ ] succeeds for valid configuration
- [ ] fails for invalid values
- [ ] returns correct exit code

### `version`

- [ ] outputs version

---

# 31. Security

Security requirements:

- [ ] Never log secret values
- [ ] Never put real secrets in `.env.example`
- [ ] Do not dump the complete environment
- [ ] Protect generated `.env` where possible
- [ ] Do not include passwords in parsing errors
- [ ] Avoid accidental `.env` commits
- [ ] Document `.gitignore` recommendations

Recommended:

```gitignore
.env
```

while keeping:

```text
.env.example
```

tracked.

`gocfg init` should initially avoid silently modifying `.gitignore`. A future explicit option can handle this.

---

# 32. Documentation

README should cover:

## Installation

```bash
go get github.com/tidjee-dev/gocfg
```

CLI:

```bash
go install github.com/tidjee-dev/gocfg/cmd/gocfg@latest
```

## Quick start

```bash
gocfg init
```

Then:

```bash
gocfg env
```

Then:

```bash
gocfg validate
```

## Configuration

Explain:

```text
config/
.env
.env.example
```

## Environment precedence

Document:

```text
OS environment
        ↓
      .env
        ↓
   Go defaults
```

## CLI reference

Document every command and option.

---

# 33. Example Application

Provide a complete example:

```text
examples/basic/
├── config/
│   ├── app.go
│   ├── database.go
│   └── config.go
├── .env.example
├── go.mod
└── main.go
```

Example:

```go
package main

import (
    "fmt"
    "log"

    "example.com/my-app/config"
)

func main() {
    cfg := config.Load()

    fmt.Println(cfg.App.Name)
}
```

The example should demonstrate the intended architecture without framework-specific code.

---

# 34. Versioning

Use semantic versioning.

Initial target:

```text
v0.1.0
```

Potential release progression:

```text
v0.1.0
v0.2.0
v0.3.0
v1.0.0
```

Avoid promising API stability before `v1.0.0`.

---

# 35. Milestones

## M0 — Repository Setup

- [ ] Create repository
- [ ] Initialize Go module
- [ ] Add license
- [ ] Add README
- [ ] Add `.gitignore`
- [ ] Establish package layout
- [ ] Add CI
- [ ] Add formatting/linting

---

## M1 — CLI Foundation

- [ ] Add Cobra
- [ ] Add Lipgloss
- [ ] Create `cmd/gocfg/main.go`
- [ ] Create Cobra root command
- [ ] Add version infrastructure
- [ ] Implement `gocfg --help`
- [ ] Keep CLI dependencies isolated

---

## M2 — `gocfg init`

- [ ] Create scaffold system
- [ ] Add embedded templates
- [ ] Generate `.env`
- [ ] Generate `.env.example`
- [ ] Generate `config/`
- [ ] Generate `config/app.go`
- [ ] Generate `config/config.go`
- [ ] Detect existing files
- [ ] Avoid accidental overwrites
- [ ] Add `--force`
- [ ] Add `--dry-run`
- [ ] Add Lipgloss output
- [ ] Test filesystem behavior

---

## M3 — `.env` Parser

- [ ] Implement parser
- [ ] Support comments
- [ ] Support whitespace
- [ ] Support quoted values
- [ ] Handle malformed entries
- [ ] Define duplicate-key behavior
- [ ] Add parser tests

---

## M4 — Environment Lookup

- [ ] Implement OS environment lookup
- [ ] Implement `.env` precedence
- [ ] Implement fallback values
- [ ] Define precedence rules
- [ ] Add tests

---

## M5 — Typed Environment Values

Implement:

- [ ] `String`
- [ ] `Bool`
- [ ] `Int`
- [ ] `Int64`
- [ ] `Float64`
- [ ] `Duration`

Add comprehensive tests.

---

## M6 — Error Handling

- [ ] Define configuration errors
- [ ] Add contextual errors
- [ ] Prevent secret leakage
- [ ] Handle malformed values
- [ ] Add error tests

---

## M7 — Required Values

- [ ] Implement required values
- [ ] Define missing-variable errors
- [ ] Add tests
- [ ] Integrate with validation

---

## M8 — Configuration Definitions

- [ ] Design configuration metadata API
- [ ] Define environment variables explicitly
- [ ] Support defaults
- [ ] Support types
- [ ] Support required values
- [ ] Avoid arbitrary Go source parsing
- [ ] Add configuration definition tests

---

## M9 — Config-Driven `.env`

Implement:

```bash
gocfg env
```

- [ ] Discover configuration definitions
- [ ] Generate missing `.env`
- [ ] Generate `.env.example`
- [ ] Preserve existing `.env`
- [ ] Add new variables
- [ ] Handle secrets safely
- [ ] Add tests
- [ ] Add `--force` only where appropriate

---

## M10 — `gocfg check`

- [ ] Implement check command
- [ ] Check project structure
- [ ] Check environment files
- [ ] Check configuration definitions
- [ ] Report inconsistencies
- [ ] Define exit codes
- [ ] Add tests

---

## M11 — `gocfg validate`

- [ ] Implement validation command
- [ ] Parse configuration values
- [ ] Check required values
- [ ] Validate types
- [ ] Execute configuration validation
- [ ] Return appropriate exit code
- [ ] Add tests

---

## M12 — Testing & Documentation

- [ ] Full unit-test suite
- [ ] CLI integration tests
- [ ] Temporary-directory tests
- [ ] README
- [ ] API documentation
- [ ] CLI documentation
- [ ] Configuration examples
- [ ] Security documentation

---

## M13 — Security & Hardening

- [ ] Audit secret handling
- [ ] Audit error messages
- [ ] Review `.env` permissions
- [ ] Review file overwrite behavior
- [ ] Review CLI exit codes
- [ ] Test malformed files
- [ ] Test unusual environment values
- [ ] Test concurrent/partial writes where relevant

---

## M14 — API Stabilization

- [ ] Review public API
- [ ] Review package names
- [ ] Review configuration-definition API
- [ ] Review CLI flags
- [ ] Remove unnecessary abstractions
- [ ] Improve documentation
- [ ] Freeze v0.1 API where appropriate

---

# 36. v0.1.0 Scope

The first useful release should contain:

### Core

- `.env` loader
- environment lookup
- precedence handling
- `String`
- `Bool`
- `Int`
- `Int64`
- `Float64`
- `Duration`
- required values
- configuration errors
- configuration definitions

### CLI

```bash
gocfg init
gocfg env
gocfg check
gocfg validate
gocfg version
```

### Scaffold

```text
.env
.env.example
config/
├── app.go
└── config.go
```

### Quality

- tests
- documentation
- security review
- CI

---

# 37. Future Features

These should remain outside the initial implementation.

## Additional environment sources

Potential future sources:

```text
.env
OS environment
CLI flags
JSON
YAML
TOML
remote configuration
secret managers
```

These should only be considered if real use cases emerge.

## Configuration snapshots

Potential API:

```go
snapshot := config.Snapshot()
```

Useful for:

- testing
- debugging
- deterministic configuration

## Custom types

Potential future API:

```go
env.URL(...)
env.IP(...)
env.BoolSlice(...)
env.StringSlice(...)
```

Only add these when there is a demonstrated need.

---

# 38. CLI Future Commands

Possible future commands:

```bash
gocfg export
gocfg diff
gocfg doctor
```

These should not be implemented until the core workflow proves useful.

Potential future behavior:

```bash
gocfg diff
```

could show differences between:

```text
configuration definitions
        vs
.env
        vs
.env.example
```

---

# 39. Design Principles

## 1. Go first

Configuration should feel like normal Go code.

## 2. Explicit over magic

Prefer:

```go
config.Load()
```

over automatic discovery.

## 3. Type safety

Prefer typed helpers:

```go
env.Int(...)
```

over:

```go
env.Get(...)
```

## 4. Configuration is code

Avoid introducing another configuration language unnecessarily.

## 5. CLI is separate

The library must remain useful without installing the CLI.

## 6. `config/` is the source of truth

Environment files are generated from explicit configuration definitions rather than arbitrary source-code parsing.

## 7. Preserve user data

`gocfg env` must not destroy existing `.env` values.

## 8. Minimal magic

The library should be understandable by a Go developer reading the source.

## 9. Secure defaults

Secrets should never accidentally appear in generated examples, logs, or errors.

## 10. Small API

Do not turn `gocfg` into a dependency-injection or application-framework system.

---

# 40. Target Developer Experience

The intended workflow should eventually be:

```bash
go install github.com/tidjee-dev/gocfg/cmd/gocfg@latest
```

Create a project:

```bash
mkdir my-app
cd my-app
go mod init example.com/my-app
```

Initialize configuration:

```bash
gocfg init
```

Result:

```text
my-app/
├── config/
│   ├── app.go
│   └── config.go
├── .env
├── .env.example
└── go.mod
```

Add configuration definitions.

Synchronize environment files:

```bash
gocfg env
```

Check the project:

```bash
gocfg check
```

Validate configuration:

```bash
gocfg validate
```

Run the application:

```bash
go run .
```

The resulting workflow should be simple enough that a developer can understand the entire configuration system without learning a separate configuration DSL.

---

# 41. Final Architecture

The intended architecture is:

```text
                         ┌───────────────────────┐
                         │     gocfg CLI         │
                         │                       │
                         │ Cobra + Lipgloss      │
                         └───────────┬───────────┘
                                     │
              ┌──────────────────────┼──────────────────────┐
              │                      │                      │
              ▼                      ▼                      ▼
           init                    env                   check
              │                      │                      │
              └──────────────────────┼──────────────────────┘
                                     │
                                     ▼
                              Configuration
                                 Engine
                                     │
                ┌────────────────────┼────────────────────┐
                │                    │                    │
                ▼                    ▼                    ▼
             env/*.go             .env              config/*
                │                    │                    │
                └────────────────────┼────────────────────┘
                                     │
                                     ▼
                               Application
                                 Config
```

The core principle remains:

```text
Go configuration definitions
           │
           ▼
       gocfg core
           │
      ┌────┴────┐
      ▼         ▼
    .env    OS environment
      │         │
      └────┬────┘
           ▼
      typed Config
```

`gocfg` should remain a **small Go configuration library with a useful project-management CLI**, rather than evolving into a full application framework.
