# gocfg — Execution Plan (thin v0.1)

Source: `ROADMAP.md` §§35–36. Original roadmap: `docs/ROADMAP.orig.md`.
Branch: `tidjee` → `origin/tidjee`. Go 1.27.1, stdlib-only core.

Decisions (MCQ 2026-09-29): error-returning getters `(T, error)`,
godotenv-compat parser, append-only `.env` merge, `Secret flag + heuristic`,
thin v0.1 (`dotenv + getters + LoadEnv + init + validate + version`).
`gocfg env`, `check`, `Var[T]` definitions → v0.2.

## Gate (every step)

```bash
gofmt -l .
go vet ./...
go test ./... -count=1 -race
```

If all green: flip checkbox, append `Done:` line, commit + `git push origin tidjee`.

## Steps

- [x] S0 PLAN.md + hygiene
  - Create this file, fix `.gitignore`, track `internal/dotenv` draft.
  - Known issues logged for S1: `parser.go:71-74` export double-trim,
    `isClosedMultiline` over-counts quotes.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, dotenv tests pass).

- [x] S1 dotenv harden (ROADMAP M1)
  - `internal/dotenv`: quotes/multiline/escapes, `export`, `${VAR}` OS-first
    expansion, whitespace, last-wins, malformed → `*ParseError` with line no.
  - `loader.go`: default `.env` missing = nil, explicit missing = error,
    never override OS, multi-path.
  - Tests: table tests + godotenv parity fixtures + `loader_test.go`.
  - Fixed: export double-trim → `cutExportPrefix`; quote closing now
    first-unescaped-quote (single-line trailing comments work, multiline
    `LastIndex` mis-close fixed); removed `countUnescaped`.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S2 env getters + errors (ROADMAP M2)
  - `env/*.go`: `String, Bool, Int, Int64, Float64, Duration, Required`,
    all `(T, error)`. Bool sets `1,true,yes,y,on` / `0,false,no,n,off`.
  - `Duration` via `time.ParseDuration`. Errors `invalid value for KEY...`,
    secret redaction (`Secret:true` + `PASSWORD|SECRET|KEY|TOKEN`).
  - Root `LoadEnv` wrapper. Tests incl. `OS > .env > default`.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S3 CLI foundation (ROADMAP M3)
  - `cobra` + `lipgloss` CLI-only, `cmd/gocfg/main.go`,
    `internal/cli/{root,version}.go`, ldflags version.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass,
    `--help` + `version` verified).

- [x] S4 `gocfg init` (ROADMAP M4)
  - `internal/scaffold/` + `embed` templates, error-returning
    `config/app.go` + `config/config.go`, `.env 0600`, `.env.example 0644`.
  - `--force/--dry-run/--name`, skip-existing, `t.TempDir()` tests.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass,
    `--dry-run` verified).

- [x] S5 `gocfg validate` (ROADMAP M5)
  - `LoadEnv` + typed resolve + `Required` + `Config.Validate()`,
    redacted output, exit 1. Tests + exit-code.
  - v0.1 takes explicit specs via flags (`--int/--bool/--duration/...`,
    `--required`, `--env-file`); Var[T] definitions deferred to v0.2.
  - Exit codes verified: 0 valid, 1 invalid, 2 usage.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S6 docs + example + audit (ROADMAP M6/M7)
  - `examples/basic/` with `Load() (Config, error)` flow, README quickstart,
    CLI reference, security doc. Final audit. Tag proposal `v0.1.0`.
  - Example verified end-to-end (`go run .` prints app name).
  - Audit: `os.Setenv` only in guarded loader, cobra/lipgloss confined to
    `internal/cli` (boundary checked via `go list -deps`), secrets redacted
    (tested), `.env` gitignored + `0600` + atomic writes (tested).
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

## Log

- (append `Done: YYYY-MM-DD <sha>` per step)
