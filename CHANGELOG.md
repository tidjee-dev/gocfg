# Changelog

All notable changes to this project are documented here.
Versioning follows [SemVer](https://semver.org/); API stability is
promised from `v1.0.0` (see `STABILITY.md`).

## Unreleased

- `gocfg init --with-definitions`: scaffolds `config/vars.go` and
  `tools/gocfg-gen` for the manifest workflow (module path from
  `go.mod`, next-step hint).

## v1.1.0 — 2026-09-30

- `gocfg init` ensures `.env` is covered by `.gitignore` by default
  (`--gitignore=false` opts out).
- Scaffolded `config/` uses `URL` types with `Env` validation
  (fixes missing `net/url` import in generated code).
- `examples/basic` demonstrates `Var[T]` definitions plus a
  committed manifest (`.gocfg.json`, `tools/gocfg-gen`).

## v1.0.0 — 2026-09-30

First stable release. Everything below is frozen per `STABILITY.md`.

- `gocfg doctor`: read-only diagnostic narrative (toolchain, files,
  parse health, `.gitignore` coverage, duplicates, OS shadowing);
  always exits 0.
- `gocfg diff [--defs]`: file-level manifest/example/env comparison
  (missing keys and default drift fail, extras and secret values warn;
  values never print).
- `gocfg export [--format shell|dotenv|json]`: prints schema keys resolved
  as `OS > .env` (missing warn on stderr, exit 0); never dumps the whole
  environment.
- Extra types: `URL`, `IP`, `StringSlice`, `BoolSlice` getters plus
  `URLVar`, `IPVar`, `StringSliceVar`, `BoolSliceVar` (kinds `url`,
  `ip`, `stringslice`, `boolslice`, wired into `validate --defs`).
- JSON manifest bridge: `env.MarshalDefinitions` / `env.UnmarshalDefinitions`
  (versioned envelope), `gocfg env --defs`, `gocfg validate --defs`
  (typed per-Kind checks).
- `cmd/gocfg`: package overview documentation (install, commands, exit codes).
- `gocfg version`: falls back to the binary's build info when no ldflags
  version was injected, so `go install ...@latest` reports the real version
  instead of `dev`.

## v0.2.0 — 2026-09-30

- `env.Var[T]` definitions (`StringVar`, `BoolVar`, `IntVar`, `Int64Var`,
  `Float64Var`, `DurationVar`, `Secret()` / `Require()`, `Resolve()`,
  `Any` interface).
- `gocfg env [--check]`: completes `.env` from `.env.example`.
- `gocfg check`: static health inspection.
- Parser: single-quoted multiline values, `-`/`.` in keys (non-leading),
  vendored godotenv fixtures with documented deltas.

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
