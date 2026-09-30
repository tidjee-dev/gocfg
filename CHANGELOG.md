# Changelog

All notable changes to this project are documented here.
Versioning follows [SemVer](https://semver.org/); no API stability is
promised before `v1.0.0`.

## Unreleased

- `env.Var[T]` definitions: `StringVar`, `BoolVar`, `IntVar`, `Int64Var`,
  `Float64Var`, `DurationVar` with `Secret()` / `Require()` options,
  `Resolve()` (`OS > .env > Default`, empty counts as unset), and the
  `Any` interface (`AnyKey`, `AnyDefault`, `AnyKind`, `IsSecret`,
  `IsRequired`) for CLI consumption. Getters now share the same parsers.

## v0.1.0 — 2026-09-30

First useful release (thin slice).

Core (stdlib-only):

- `.env` loader: godotenv-compatible behavior (quotes, multiline, `export`,
  `${VAR}` expansion), malformed lines error with line numbers,
  duplicate keys last-wins.
- Explicit `LoadEnv()` with `OS > .env > Go default` precedence;
  OS values are never overridden.
- Typed getters returning `(T, error)`: `String`, `Bool`, `Int`, `Int64`,
  `Float64`, `Duration`, `Required`. Invalid values error, never silently
  fall back. Secret keys redacted from error messages (heuristic).

CLI (`cobra`, isolated from core):

- `gocfg init [--force] [--dry-run] [--name]`: scaffolds `.env` (`0600`),
  `.env.example`, `config/` with error-returning getters; skips existing
  files; atomic writes.
- `gocfg validate [--env-file]`: no flags validates every key declared in
  `.env.example` (extras warn); flags validate typed specs. Exit codes
  `0` valid, `1` invalid, `2` usage.
- `gocfg version` (ldflags-injectable).

Docs: README, `examples/basic` end-to-end application.

Deferred to v0.2: `Var[T]` definitions, `gocfg env`, `gocfg check`.
