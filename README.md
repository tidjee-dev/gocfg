# gocfg

[![Go Reference](https://pkg.go.dev/badge/github.com/tidjee-dev/gocfg.svg)](https://pkg.go.dev/github.com/tidjee-dev/gocfg)
[![CI](https://github.com/tidjee-dev/gocfg/actions/workflows/ci.yml/badge.svg)](https://github.com/tidjee-dev/gocfg/actions/workflows/ci.yml)

A small, type-safe configuration library for Go applications.

Configuration is written in **Go**. Environment-specific values come from
`.env` and OS environment variables. No YAML/JSON/TOML required.

```go
port, err := env.Int("APP_PORT", 9000)
if err != nil {
    return err // invalid value for APP_PORT: "hello": expected integer
}
```

Typed getters cover `String`, `Bool`, `Int`, `Int64`, `Float64`,
`Duration`, `URL` (scheme required), `IP`, `StringSlice` / `BoolSlice`
(comma-separated, quote-aware) — each with a matching `*Var`
definition constructor.

## Installation

Library:

```bash
go get github.com/tidjee-dev/gocfg
```

CLI:

```bash
go install github.com/tidjee-dev/gocfg/cmd/gocfg@latest
```

## Quick start

```bash
mkdir my-app && cd my-app
go mod init example.com/my-app
gocfg init
```

This creates `.env`, `.env.example`, and `config/` with error-returning
getters. Then:

```bash
gocfg validate --int APP_PORT --bool APP_DEBUG
go run .
```

See `examples/basic/` for a complete application.

## Configuration

The application owns its configuration in `config/`:

```go
func Load() (Config, error) {
    app, err := App()
    if err != nil {
        return Config{}, err
    }
    return Config{App: app}, nil
}
```

`gocfg` never discovers or executes `config/*.go` on its own.

## Definitions

Explicit `Var[T]` definitions are the source of truth the CLI consumes:

```go
var (
    AppName    = env.StringVar("APP_NAME", "My App")
    AppPort    = env.IntVar("APP_PORT", 9000)
    DBPassword = env.StringVar("DB_PASSWORD", "", env.Secret(), env.Require())
)

func Definitions() []env.Any {
    return []env.Any{AppName, AppPort, DBPassword}
}
```

`Resolve()` follows `OS > .env > Default`; empty counts as unset
(optional → default, required → error). `Secret()` redacts the value
from errors and generated examples even when the key looks innocent.

## Manifest bridge

Apps can publish their definitions for the CLI (versioned envelope):

```go
// tools/gocfg-gen/main.go
package main

import (
    "fmt"
    "os"

    "example.com/my-app/config"
    "github.com/tidjee-dev/gocfg/env"
)

func main() {
    raw, err := env.MarshalDefinitions(config.Definitions())
    if err != nil {
        panic(err)
    }
    fmt.Println(string(raw))
}
```

```bash
go run ./tools/gocfg-gen > .gocfg.json
gocfg env --defs .gocfg.json        # generate .env in manifest order
gocfg validate --defs .gocfg.json  # type-check per Kind
```

## Environment precedence

```text
OS environment > .env > Go default
```

Load explicitly (no `init()` magic, OS values are never overridden):

```go
if err := gocfg.LoadEnv(); err != nil {
    log.Fatal(err)
}
```

## CLI reference

```bash
gocfg init [--force] [--dry-run] [--name my-app]
gocfg env [--env-file .env --example-file .env.example --check]
gocfg check [--env-file .env --example-file .env.example]
gocfg export [--format shell|dotenv|json]
gocfg validate [--env-file .env]
gocfg validate --int APP_PORT --bool APP_DEBUG --required DATABASE_URL
gocfg version
```

`env` ensures `.env` contains every `.env.example` key (appended with
example values, empty for secrets); existing content is never touched.
`--check` reports without writing (exit 1 when out of sync).

`check` is the static health inspection: `config/` exists, files parse,
every schema key present in `.env` or OS (empty counts as present).
It never touches the environment and performs no type coercion —
that is `validate`'s job.

With no flags, `validate` checks every key declared in `.env.example`
(OS > `.env`), warns about extra `.env` keys, and fails on missing or
malformed entries.

`export` prints schema keys resolved as `OS > .env` for sourcing
(`eval "$(gocfg export)"`) or piping (`--format json`). Only declared
keys are emitted, never the whole environment — but values print
as-is on your explicit request, so keep the output out of logs.

Exit codes: `0` success, `1` configuration/validation failure,
`2` CLI usage error.

## Security

- Secret values (`Secret()` option, or keys containing `PASSWORD`,
  `SECRET`, `KEY`, `TOKEN`) are redacted from error messages and never
  written to `.env.example`.
- Generated `.env` uses `0600` permissions; writes are atomic.
- Commit `.env.example`, never commit `.env`:

```gitignore
.env
```

## Status

v0.1: `.env` loader, typed getters, `init`, `validate`, `version`.
See `ROADMAP.md` and `PLAN.md`.
