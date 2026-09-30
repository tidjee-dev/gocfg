# gocfg — Detailed Roadmap

> A small, type-safe configuration library for Go applications.
>
> Original version preserved in `ROADMAP.orig.md`.
> Decisions from MCQ review (2026-09-29): error-returning getters,
> godotenv-compat parser, append-only `.env` merge, `Secret flag + heuristic`,
> thin `v0.1` slice (`dotenv + getters + LoadEnv + init + validate`).

Repository:

```text
github.com/tidjee-dev/gocfg
```

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

func App() (AppConfig, error) {
    name, err := env.String("APP_NAME", "My App")
    if err != nil {
        return AppConfig{}, err
    }
    envVal, err := env.String("APP_ENV", "dev")
    if err != nil {
        return AppConfig{}, err
    }
    debug, err := env.Bool("APP_DEBUG", true)
    if err != nil {
        return AppConfig{}, err
    }
    url, err := env.String("APP_URL", "http://localhost:9000")
    if err != nil {
        return AppConfig{}, err
    }
    return AppConfig{
        Name:  name,
        Env:   envVal,
        Debug: debug,
        URL:   url,
    }, nil
}
```

The application composes configuration explicitly (decided: `Load() (Config, error)`):

```go
package config

type Config struct {
    App      AppConfig
    Database DatabaseConfig
}

func Load() (Config, error) {
    app, err := App()
    if err != nil {
        return Config{}, err
    }
    db, err := Database()
    if err != nil {
        return Config{}, err
    }
    return Config{
        App:      app,
        Database: db,
    }, nil
}
```

`gocfg` should **not** dynamically discover and execute `config/*.go`.

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

# 8. Environment API

Decided (MCQ): all typed getters return `(T, error)`.
Invalid values error instead of silently using the fallback.

```go
env.String(key, fallback string) (string, error)
env.Bool(key string, fallback bool) (bool, error)
env.Int(key string, fallback int) (int, error)
env.Int64(key string, fallback int64) (int64, error)
env.Float64(key string, fallback float64) (float64, error)
env.Duration(key string, fallback time.Duration) (time.Duration, error)
env.Required(key string) (string, error)
```

Examples:

```go
port, err := env.Int("APP_PORT", 9000)
if err != nil {
    return err
}

debug, err := env.Bool("APP_DEBUG", false)
if err != nil {
    return err
}

timeout, err := env.Duration(
    "APP_TIMEOUT",
    30*time.Second,
)
if err != nil {
    return err
}
```

# 9. Environment Parsing

## String

```go
env.String("APP_NAME", "My App")
```

Returns the environment value if present, otherwise the fallback.

## Boolean

Decided: case-insensitive, trimmed. Accepted true set:
`1, true, yes, y, on`. Accepted false set: `0, false, no, n, off`.
Anything else errors. Document and test exhaustively.

## Integer

```go
port, err := env.Int("APP_PORT", 9000)
```

Invalid values return an error rather than silently using the fallback.

Example:

```text
invalid value for APP_PORT: "hello": expected integer
```

## Int64

```go
v, err := env.Int64("APP_MAX_SIZE", 100000)
```

## Float64

```go
v, err := env.Float64("APP_RATE", 1.5)
```

## Duration

Uses `time.ParseDuration` semantics.

```go
v, err := env.Duration("APP_TIMEOUT", 30*time.Second)
```

Example:

```dotenv
APP_TIMEOUT=30s
```

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

Secrets must not be included in error messages. Decided (MCQ): `Flag + heuristic`.
A value is secret if its definition has `Secret: true` OR its key contains
`PASSWORD`, `SECRET`, `KEY`, or `TOKEN` (case-insensitive). Both trigger
redaction in errors and empty values in `.env.example`.

Example:

```text
invalid value for DB_PASSWORD
```

rather than:

```text
invalid value for DB_PASSWORD: "super-secret-password"
```

# 11. Required Values

Decided v0.1 API (part of thin slice):

```go
v, err := env.Required("APP_KEY")
```

Empty or unset returns an error:

```text
required environment variable APP_KEY is not set
```

Required values should be useful for:

- production secrets
- API keys
- database credentials
- signing keys

# 12. `.env` Support

`.env` is optional. Decided (MCQ): `godotenv compat` for v0.1.

The loader must support:

- comments (`#`)
- empty lines
- quoted values (single, double)
- multiline quoted values
- basic escaping (`\n`, `\"`, `\\` inside double quotes)
- `export KEY=value` prefix
- `${VAR}` / `$VAR` expansion against OS env + earlier keys
- whitespace trimming around key, `=`, unquoted values
- keys: ASCII `[A-Za-z_][A-Za-z0-9_.-]*` (no leading digit, no `:` separator)
- single-quoted multiline values (literal, like double-quoted)
- duplicate keys: last wins (documented, tested)
- malformed line: return error with line number, do not silently skip
- parity: `internal/dotenv/testdata/godotenv-*.env` snapshots with per-fixture
  expectations; documented deltas (OS-first expansion, `\t` → tab)

Typical:

```dotenv
APP_NAME=My App
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

The loader should support (see list above) plus environment precedence:

- `LoadEnv` populates only keys not already present in OS env
  (`OS > .env > Go default` is enforced by never overriding OS).
  Tests use `t.Setenv`; no override API is provided.

# 13. `.env` Loading

Avoid implicit package-level side effects.

Do not use:

```go
func init() {
    loadDotEnv()
}
```

Instead, provide explicit loading. Decided API:

```go
err := gocfg.LoadEnv()            // defaults to ".env", no override
err := gocfg.LoadEnv(".env")      // custom path(s)
err := gocfg.LoadEnv(".env", ".env.local")
```

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

# 15. Configuration Definitions

A major design goal is making `config/` the source of truth for environment configuration.

The CLI should eventually know which environment variables are used by the application.

Avoid trying to achieve this by parsing arbitrary Go AST/source code.
No global mutable registry. No reflection over arbitrary structs.

Decided direction (explicit list, deferred to v0.2 for CLI use):

```go
// core: order-preserving, type-safe definition
type Var[T any] struct {
    Key     string
    Default T
    Secret  bool
    Parse   func(string) (T, error)
}

func StringVar(key, def string, opts ...VarOpt) Var[string]
func BoolVar(key string, def bool, opts ...VarOpt) Var[bool]
// ... IntVar, Int64Var, Float64Var, DurationVar
func (v Var[T]) Resolve() (T, error) // OS > .env-loaded env > Default

type Any interface { AnyKey() string; AnyDefault() string; IsSecret() bool }

// app side: single source of truth the CLI can consume in v0.2
var AppName = env.StringVar("APP_NAME", "My App")
var AppDebug = env.BoolVar("APP_DEBUG", true)
var DBPassword = env.StringVar("DB_PASSWORD", "", env.Secret())

func Definitions() []env.Any {
    return []env.Any{AppName, AppDebug, DBPassword}
}
```

Requirements:

- type-safe
- explicit
- easy to use from normal Go configuration
- usable by the CLI
- no arbitrary Go source parsing
- no reflection-heavy magic

# 16. Config-Driven Environment Generation

`gocfg env` completes `.env` from `.env.example` (the file-level schema,
same model as validate-all). `Var[T]` definitions (§15) are the Go-side
source of truth for application code; a manifest bridge from
`Definitions()` to the CLI remains possible future work.

The CLI inspects the example file and generates/updates:

```text
.env
.env.example
```

Command:

```bash
gocfg env
```

Example definitions:

```go
var AppName = env.StringVar("APP_NAME", "My App")
var AppEnv = env.StringVar("APP_ENV", "dev")
var AppDebug = env.BoolVar("APP_DEBUG", true)
var AppURL = env.StringVar("APP_URL", "http://localhost:9000")
```

Produces:

```dotenv
APP_NAME=My App
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

# 17. `.env` Preservation

Decided (MCQ): `append-missing-only`. `gocfg env` must **not destroy
existing environment values, comments, order, or formatting**.

Rules:

- parse existing `.env` preserving raw lines
- append only missing keys at end as `KEY=default`
- never modify existing values, comments, blank lines, or order
- never delete keys (report extras via `check` instead)
- write atomically (tmp file + rename), new files `0600`, existing
  permissions preserved across the rename

Existing:

```dotenv
APP_NAME=My Production App
APP_ENV=prod
APP_DEBUG=false
```

Configuration later adds:

```go
var AppTimezone = env.StringVar("APP_TIMEZONE", "Europe/Brussels")
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

# 18. `.env.example` Generation

`.env.example` represents the configuration schema (defaults, not live values).

Example:

```dotenv
APP_NAME=My App
APP_ENV=dev
APP_DEBUG=true
APP_URL=http://localhost:9000
```

Secrets (decided: `Secret:true` flag + heuristic on
`PASSWORD|SECRET|KEY|TOKEN`) are always empty, never real values.

For example:

```dotenv
DB_PASSWORD=
APP_SECRET=
API_KEY=
```

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

func App() (AppConfig, error) {
    name, err := env.String("APP_NAME", "My App")
    if err != nil {
        return AppConfig{}, err
    }
    envVal, err := env.String("APP_ENV", "dev")
    if err != nil {
        return AppConfig{}, err
    }
    debug, err := env.Bool("APP_DEBUG", true)
    if err != nil {
        return AppConfig{}, err
    }
    url, err := env.String("APP_URL", "http://localhost:9000")
    if err != nil {
        return AppConfig{}, err
    }
    return AppConfig{Name: name, Env: envVal, Debug: debug, URL: url}, nil
}
```

### `config/config.go`

```go
package config

type Config struct {
    App AppConfig
}

func Load() (Config, error) {
    app, err := App()
    if err != nil {
        return Config{}, err
    }
    return Config{App: app}, nil
}
```

The exact generated API will follow the final configuration-definition design.

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

# 22. File Permissions

Generated files should use appropriate permissions.

Recommended:

```text
.env            0600
.env.example    0644
config/*.go     0644
```

The exact behavior should account for existing files and platform differences.

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

# 25. `gocfg check`

Decided (MCQ): `check` = shallow static check. Fast, no app import.

```bash
gocfg check
```

Checks:

- `.env` exists
- `.env.example` exists
- `config/` exists
- required keys present (string presence only, no type coercion)
- `.env` parses without error (line numbers on failure)
- missing / extra keys vs `.env.example` reported as warnings

Example:

```text
gocfg check

✓ configuration directory
✓ environment file
✓ environment example
✓ keys present
✓ .env parses

Configuration looks consistent.
```

`check` inspects and reports. Exit 1 on missing/unparseable.
Shipped in v0.2 alongside `validate` (values) — see §26 for the split.

# 26. `gocfg validate`

`validate` checks values (deep), `check` checks structure (shallow, §25).

With no flags, every key declared in `.env.example` must resolve to a
non-empty value (`Required` semantics, `OS > .env`); extras warn.
With flags, the given typed specs are resolved instead:

```bash
gocfg validate
gocfg validate --int APP_PORT --bool APP_DEBUG --required DATABASE_URL
```

Steps (no-flag mode):

1. `LoadEnv()` (no override; malformed `.env` fails)
2. `Required` check per schema key
3. extras reported as warnings

No app `Config` import: domain rules stay in user code
(`Config.Validate()` in `examples/basic`).

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

# 28. CLI Exit Codes

The CLI should use meaningful exit statuses.

Example:

```text
0 = success
1 = configuration/validation failure
2 = CLI usage error
```

Exact conventions should be finalized during implementation.

# 29. Configuration Validation

Separate parsing (`gocfg` responsibility) from domain rules (app responsibility).
Decided: `Load() (Config, error)` parses, then `Validate() error` checks domain.

Example:

```go
type Config struct {
    App      AppConfig
    Database DatabaseConfig
}

func Load() (Config, error) {
    // typed resolve, returns first parse/required error
}

func (c Config) Validate() error {
    // application-specific validation, e.g. Env in {dev,prod}
    return nil
}
```

`gocfg validate` runs `Load` + `Validate` and exits 1 with redacted errors.

# 30. Testing Strategy

Testing should be a major part of the project.

## Environment tests (v0.1: done, see `env/env_test.go`)

- [x] string fallback
- [x] string environment value
- [x] boolean parsing (all true/false sets + invalid)
- [x] integer parsing
- [x] int64 parsing
- [x] float parsing
- [x] duration parsing (`time.ParseDuration` cases)
- [x] invalid values return error (not fallback)
- [x] precedence `OS > .env > default`, `LoadEnv` never overrides OS
- [x] secret redaction in errors (heuristic; `Secret` flag arrives with `Var[T]` in v0.2)

## `.env` tests (v0.1: done, see `internal/dotenv/`)

- [x] empty file
- [x] comments
- [x] whitespace
- [x] quoted / multiline / escaped values (compat table tests; vendored fixtures in S12)

- [x] empty file
- [x] comments
- [x] whitespace
- [x] quoted / multiline / escaped values (godotenv parity)
- [x] `export` prefix
- [x] `${VAR}` expansion
- [x] duplicate keys last-wins
- [x] malformed entries error with line number
- [x] precedence against OS environment
- [x] custom path + multi-path

## Configuration tests

- [ ] configuration definitions (v0.2, S9)
- [x] defaults
- [x] required values
- [x] validation (`gocfg validate` + `Config.Validate` in example)
- [x] invalid configuration

## CLI tests

Use `t.TempDir()` for filesystem isolation.

### `init` (v0.1: done)

- [x] creates `.env`
- [x] creates `.env.example`
- [x] creates `config/`
- [x] creates config files
- [x] does not overwrite existing files
- [x] `--force` overwrites
- [x] `--dry-run` does not modify files

### `env` (v0.2, S10)

- [ ] creates missing `.env`
- [ ] creates missing `.env.example`
- [ ] adds new configuration variables
- [ ] preserves existing `.env` values
- [ ] updates `.env.example`
- [ ] does not expose secrets
- [ ] handles malformed configuration

### `check` (v0.2, S11)

- [ ] detects missing `.env`
- [ ] detects missing `config/`
- [ ] detects invalid definitions
- [ ] reports inconsistencies

### `validate` (v0.1: done)

- [x] succeeds for valid configuration
- [x] fails for invalid values
- [x] returns correct exit code

### `version` (v0.1: done)

- [x] outputs version

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
    if err := gocfg.LoadEnv(); err != nil {
        log.Fatal(err)
    }
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    if err := cfg.Validate(); err != nil {
        log.Fatal(err)
    }

    fmt.Println(cfg.App.Name)
}
```

The example should demonstrate the intended architecture without framework-specific code.

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

# 35. Milestones

Decided order: core first, then CLI. Thin v0.1 = M0–M7. M8–M10 deferred to v0.2.

## M0 — Repository Setup

- [ ] Create repository
- [ ] Initialize Go module
- [ ] Add license
- [ ] Add README
- [ ] Add `.gitignore`
- [ ] Establish package layout
- [ ] Add CI
- [ ] Add formatting/linting

## M1 — `.env` Parser + Lookup (core first)

- [ ] Implement godotenv-compat parser (quotes, multiline, `export`, expansion)
- [ ] Malformed lines error with line number
- [ ] Duplicate keys last-wins
- [ ] `LoadEnv` explicit, no `init()` side effects, never overrides OS env
- [ ] Precedence `OS > .env > default`, custom/multi-path
- [ ] Parser + precedence tests, godotenv parity fixtures

## M2 — Typed Values + Errors (core)

- [ ] `String, Bool, Int, Int64, Float64, Duration` returning `(T, error)`
- [ ] Bool sets: `1,true,yes,y,on` / `0,false,no,n,off`, case-insensitive
- [ ] `Duration` via `time.ParseDuration`
- [ ] `Required` returning `(string, error)`
- [ ] Contextual errors `invalid value for KEY: "v": expected X`
- [ ] Secret redaction: `Secret:true` flag + `PASSWORD|SECRET|KEY|TOKEN` heuristic
- [ ] Comprehensive tests

## M3 — CLI Foundation

- [ ] Add Cobra
- [ ] Add Lipgloss
- [ ] Create `cmd/gocfg/main.go`
- [ ] Create Cobra root command
- [ ] Add version infrastructure (`ldflags` injection)
- [ ] Implement `gocfg --help`
- [ ] Keep CLI dependencies isolated

## M4 — `gocfg init` (static, v0.1)

- [ ] Create scaffold system
- [ ] Add embedded templates
- [ ] Generate `.env` (`0600`), `.env.example` (`0644`), `config/` (`0644`)
- [ ] Generate error-returning `config/app.go`, `config/config.go`
- [ ] Detect existing files, skip with message, `--force` to overwrite
- [ ] Add `--dry-run`, `--name`
- [ ] Add Lipgloss output
- [ ] Test filesystem behavior with `t.TempDir()`

## M5 — `gocfg validate` (v0.1)

- [ ] `LoadEnv` + typed resolve + `Required` + `Config.Validate()`
- [ ] Per-var output, redacted secrets, exit 1 on failure
- [ ] Tests + exit-code tests

## M6 — Testing & Docs (v0.1 close)

- [ ] Full unit-test suite, `t.TempDir()` CLI tests
- [ ] README, API docs, CLI reference, security docs
- [ ] `examples/basic` with `Load() (Config, error)` flow
- [ ] Security audit (no secret logs, `.gitignore`, `0600`)

## M7 — API Stabilization (v0.1 freeze)

- [ ] Review public API, package names, CLI flags
- [ ] Remove unnecessary abstractions
- [ ] Freeze v0.1 API

## M8 — Configuration Definitions (v0.2: done)

- [x] `Var[T]{Key, Default, Kind, Secret, Required, Parse}` + `Resolve() (T, error)`
- [x] `Definitions() []Any` explicit list, no globals, no AST parsing
- [x] Defaults, types, required (`Require()` opt), `Secret()` opt
- [x] Definition tests (precedence, empty≡unset, redaction, `Any` conformance)

## M9 — Config-Driven `.env` (v0.2: done, file-driven)

`gocfg env` completes `.env` from `.env.example` (the file-level schema;
`Var[T]` remains the Go-side API — no app import, no manifest).

- [x] Read schema from `.env.example`
- [x] Generate missing `.env` (`0600`)
- [x] Append-missing-only merge, atomic write, preserve comments/order/mode
- [x] Add new variables (example values; secrets appended empty)
- [x] Handle secrets safely (redacted warnings, empty in `.env`)
- [x] Tests (incl. failed-write atomicity)
- [x] `--check` instead of `--force` (reports without writing)

## M10 — `gocfg check` (v0.2: done)

- [x] Shallow static check: files exist, `.env` parses, keys present
- [x] Missing keys fail, extras as warnings
- [x] No type coercion (that's `validate`)
- [x] Exit codes 0/1/2
- [x] Tests (incl. OS-satisfies-without-file, empty-counts-present)

## M11 — Hardening (v0.2: done)

- [x] Audit secret handling, error messages, permissions
- [x] Review overwrite behavior, exit codes
- [x] Test malformed files, unusual values, partial writes
- [x] Vendored godotenv fixtures for parser parity

# 36. Release Scope

v0.1.0 shipped the thin slice (loader, typed getters, `init`,
flag-spec `validate`, `version`). v0.2.0 adds:

### Core (v0.2)

- `Var[T]` definitions (`StringVar`, `BoolVar`, `IntVar`, `Int64Var`,
  `Float64Var`, `DurationVar`) with `Secret()` / `Require()` options
- `Resolve()` (`OS > .env > Default`, empty counts as unset)
- `Any` interface for CLI consumption
- single-quoted multiline values; `-`/`.` allowed in keys (non-leading)

### CLI (v0.2)

```bash
gocfg init
gocfg env        # file-driven sync from .env.example (+ --check)
gocfg check      # shallow static health
gocfg validate   # no-flag schema presence + typed flag specs
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
