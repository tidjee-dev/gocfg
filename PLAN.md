# gocfg — Execution Plan

Source: `ROADMAP.md` (v0.1 §§35–36, v0.2 M8–M11). Original: `docs/ROADMAP.orig.md`.
Branch: `tidjee` → `origin/tidjee`. Go 1.27.1, stdlib-only core.

Decisions (MCQ 2026-09-29 + follow-ups): error-returning getters `(T, error)`,
godotenv-compat behavior (own table tests; godotenv fixtures not vendored yet),
append-only `.env` merge, heuristic-only secret redaction in v0.1
(`Secret` flag arrives with `Var[T]` in v0.2), thin v0.1
(`dotenv + getters + LoadEnv + init + validate + version`), no-flag
`validate` checks all `.env.example` keys. `gocfg env`, `check`, `Var[T]` → v0.2.

## Gate (every step)

```bash
gofmt -l .
go vet ./...
go test ./... -count=1 -race
```

If all green: flip checkbox, append `Done:` line, commit + `git push origin tidjee`.

## v0.1 — done (accurate record)

- [x] S0 PLAN.md + hygiene
  - Created this file, fixed `.gitignore`, tracked `internal/dotenv` draft.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, dotenv tests pass).

- [x] S1 dotenv harden (ROADMAP M1)
  - `internal/dotenv`: quotes/multiline/escapes, `export`, `${VAR}` OS-first
    expansion, whitespace, last-wins, malformed → `*ParseError` with line no.
  - `loader.go`: default `.env` missing = nil, explicit missing = error,
    never overrides OS, multi-path.
  - Tests: compat table tests + `loader_test.go`. NOTE: godotenv fixtures
    were never vendored — tracked in S12.
  - Fixed: export double-trim → `cutExportPrefix`; quote closing is now
    first-unescaped-quote (single-line trailing comments work, multiline
    `LastIndex` mis-close fixed); removed `countUnescaped`.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S2 env getters + errors (ROADMAP M2)
  - `env/*.go`: `String, Bool, Int, Int64, Float64, Duration, Required`,
    all `(T, error)`. Bool sets `1,true,yes,y,on` / `0,false,no,n,off`.
  - `Duration` via `time.ParseDuration`. Errors `invalid value for KEY...`,
    redacted via key heuristic (`PASSWORD|SECRET|KEY|TOKEN`). NOTE: no
    `Secret` flag in v0.1 — it arrives with `Var[T]` in S9.
  - Root `LoadEnv` wrapper. Tests incl. `OS > .env > default`.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S3 CLI foundation (ROADMAP M3)
  - `cobra` + `lipgloss` CLI-only, `cmd/gocfg/main.go`,
    `internal/cli/{root,version}.go`, ldflags version.
  - Done: 2026-09-30 (gate: `--help` + `version` verified).

- [x] S4 `gocfg init` (ROADMAP M4)
  - `internal/scaffold/` + `embed` templates, error-returning
    `config/app.go` + `config/config.go`, `.env 0600`, `.env.example 0644`.
  - `--force/--dry-run/--name`, skip-existing, atomic writes, `t.TempDir()` tests.
  - Done: 2026-09-30 (gate: `--dry-run` verified live).

- [x] S5 `gocfg validate` specs (ROADMAP M5)
  - Typed flag specs (`--int/--bool/--duration/...`, `--required`, `--env-file`).
    NOTE: does not call app `Config.Validate()` — impossible without an app
    import in v0.1; domain validation stays in user code (see `examples/basic`).
  - Exit codes verified live: 0 valid, 1 invalid, 2 usage.
  - Done: 2026-09-30 (gate: tests -race pass).

- [x] S6 docs + example + audit (ROADMAP M6)
  - README (install, quickstart, precedence, CLI reference, security),
    `examples/basic/` with `Load() (Config, error)` flow, verified end-to-end.
  - Audit: `os.Setenv` only in guarded loader, cobra/lipgloss confined to
    `internal/cli` (checked via `go list -deps`), secrets redacted (tested),
    `.env` gitignored + `0600` + atomic writes (tested).
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S7 validate-all
  - No-flag `validate` checks every key declared in `.env.example`
    (`Required` semantics), warns on extras, fails on malformed/missing.
    Typed flags unchanged. README updated. Verified live in /tmp/gocfg-demo.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

## v0.1 release — pending

- [x] S8 release
  - MIT LICENSE, CHANGELOG.md, CONTRIBUTING.md, CI (`gofmt`/`vet`/`test -race`).
  - M7 review: public surface frozen (`env` getters, `LoadEnv`, CLI flags);
    no dead code; added empty-file test; §30 boxes checked off.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

## v0.2 — planned (ROADMAP M8–M11)

- [x] S9 `Var[T]` definitions (M8)
  - `Var[T]{Key, Default, Kind, Secret, Required, Parse}` + key-aware
    `Parser[T]`, `Resolve() (T, error)` (`OS > .env > Default`, empty
    counts as unset), `Secret()` / `Require()` opts (non-generic `VarOpt`;
    `Require` because `Required` is taken), extended `Any`
    (`AnyKey/AnyDefault/AnyKind/IsSecret/IsRequired`).
  - Getters refactored onto shared `parseBool/parseInt/...`; behavior
    unchanged. Custom `Parse` errors: `*Error` kept with redaction
    enforced, others pass through (secret Vars redact by replacement).
  - README definitions section, CHANGELOG Unreleased entry.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S10 `gocfg env` (M9, reshaped: file-driven, not Definitions discovery)
  - `.env.example` is the schema (same model as validate-all): append missing
    keys with example values (empty for secrets, with warning), create
    missing `.env` (`0600`), never touch values/comments/order, atomic
    write preserving existing mode, extras warn only.
  - `--env-file/--example-file/--check` (`--check` exits 1 when out of sync).
  - `internal/envsync` core + cobra wiring, README + CHANGELOG.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S11 `gocfg check` (M10)
  - Static only: no env mutation, `config/` exists, files parse, schema
    keys present in `.env` or OS (empty counts — validate owns values);
    missing `.env` warns with OS fallback; extras/secret-values warn.
  - `internal/cli/check.go` + tests, README split documented.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

- [x] S12 hardening + roadmap sync + v0.2.0 (M11)
  - Vendored godotenv fixtures (`testdata/godotenv-*.env` + attribution) with
    per-fixture parity tests; closes S1 NOTE. Parser gaps closed:
    single-quoted multiline, `-`/`.` in keys. Deltas documented in code:
    OS-first expansion, `\t` → tab, no `:` separator.
  - Secret/write safety: warning-text redaction, failed-write atomicity
    (original intact, no stray tmps, mode preserved).
  - ROADMAP sync: §26 (actual validate), §13 (`Override` dropped),
    §16/§17/§25/§36 (shipped scope), M8–M11 + §30 checked.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).

## v1 additions (all selected 2026-09-30)

- [x] S13 JSON bridge
  - `env/bridge.go`: versioned envelope (`{"version":1,"vars":[...]}`),
    `MarshalDefinitions` / `UnmarshalDefinitions` (rejects bad envelope,
    empty keys; kinds validated at use).
  - `envsync.RunDefs`: manifest-ordered codegen (secrets + required
    appended empty with warning); `env --defs`, `validate --defs`
    (typed per-Kind, unknown kinds fail naming the key).
  - README manifest recipe. No app import, no manifest staleness invented:
    manifest is an explicit user artifact.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).
- [x] S14 extra types (URL, IP, StringSlice, BoolSlice)
  - Getters + `Var` constructors + `AnyKind` + `validate --defs` kinds.
    `splitList` groups both quote types (stripped), drops empties.
    `AnyDefault` renders Stringer/slices canonically, guards typed nils.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).
- [x] S15 `gocfg export`
  - Schema-bounded output (`OS > .env`, never whole environ), `shell`
    (POSIX-quoted) / `dotenv` (minimal quoting) / `json` formats, missing
    keys warn on stderr with exit 0, bad format exits 2.
  - `internal/envexport` core + cobra wiring, README + CHANGELOG.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).
- [x] S16 `gocfg diff`
  - File-level three-way (manifest optional): missing keys + default
    drift fail, extras/secret-values/example-only warn, values never
    print, OS ignored. `internal/envdiff` core + cobra wiring.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).
- [x] S17 `gocfg doctor`
  - Read-only narrative (toolchain, files + perms, parse health,
    `.gitignore` coverage, duplicate keys, OS shadowing, example
    secrets); always exits 0. `internal/cli/doctor.go` + tests.
  - Done: 2026-09-30 (gate: gofmt clean, vet clean, tests -race pass).
- [x] S18 v1 close-out (STABILITY.md, docs, gate, tag v1.0.0)
  - `STABILITY.md` frozen surface; `Required`/`Require()` documented as-is;
    README Status, ROADMAP §34/§36, CHANGELOG `v1.0.0`.
  - Full gate + dep-boundary + secret audit + live smoke of all commands.
  - Done: 2026-09-30.

## Post-v1 (on tidjee, unreleased)

- [x] `init --with-definitions` (opt-in manifest scaffolding)
  - `config/vars.go` + `tools/gocfg-gen` templates, module path from
    `go.mod` (example.com fallback + hint), skip-existing/dry-run
    aware; cold-scaffold loop verified live (`init` → build →
    manifest → `env`/`validate --defs`).
- [x] Scaffold `.gitignore` (default on, `--gitignore=false` opts out)
  - `scaffold.Options.Gitignore`, create/append/skip-if-covered, atomic
    `0644`, dry-run aware; CLI output lines; `t.TempDir()` tests.
- [x] Scaffold + example consistency (URL types, Env validation)
  - Fixed missing `net/url` import (generated code didn't compile);
    template + `examples/basic` use `env.URL`; template `Validate()`
    matches example. Generated project verified with `go build`.
- [x] Example depth (`Var[T]` showcase)
  - `examples/basic/config/vars.go` + `Definitions()`,
    `tools/gocfg-gen`, committed `.gocfg.json`; manifest loop
    (`env`/`validate`/`diff --defs`) verified live.
- [x] Docs: README flags, CHANGELOG Unreleased, ROADMAP §31 amendment.

## Post-v1: internal/config migration (breaking → v2.0.0)

- [ ] Scaffold `Options.ConfigDir` (default `internal/config`),
  `ConfigImport` for templates, `..`/absolute rejection.
- [ ] `--config-dir` on `init`/`check`/`doctor`; legacy shim
  (warn always, strict on explicit flag).
- [ ] `examples/basic`: `git mv` + imports + manifest.
- [ ] Tests: goldens, legacy/custom-dir/rejection/matrix cases.
- [ ] Docs: README, `main.go` doc, STABILITY, ROADMAP, CHANGELOG.
- [ ] Gate, merge, tag v2.0.0, proxy verify.

## Traceability (ROADMAP → PLAN)

```text
M0  → S0 + S8
M1  → S1
M2  → S2
M3  → S3
M4  → S4
M5  → S5 + S7
M6  → S6
M7  → S8
M8  → S9
M9  → S10
M10 → S11
M11 → S12
```
