# gocfg

A small, type-safe configuration library for Go applications.

Configuration is written in **Go**. Environment-specific values come from
`.env` and OS environment variables. No YAML/JSON/TOML required.

```go
port, err := env.Int("APP_PORT", 9000)
if err != nil {
    return err // invalid value for APP_PORT: "hello": expected integer
}
```

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
gocfg validate --int APP_PORT --bool APP_DEBUG --required DATABASE_URL [--env-file .env]
gocfg version
```

Exit codes: `0` success, `1` configuration/validation failure,
`2` CLI usage error.

`gocfg env` (sync) and `gocfg check` arrive in v0.2 with the explicit
`Var[T]` definitions API.

## Security

- Secret values (keys containing `PASSWORD`, `SECRET`, `KEY`, `TOKEN`)
  are redacted from error messages and never written to `.env.example`.
- Generated `.env` uses `0600` permissions; writes are atomic.
- Commit `.env.example`, never commit `.env`:

```gitignore
.env
```

## Status

v0.1: `.env` loader, typed getters, `init`, `validate`, `version`.
See `ROADMAP.md` and `PLAN.md`.
