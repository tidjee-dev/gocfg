# Stability policy (v1.0.0)

SemVer applies from `v1.0.0`: breaking changes bump the major version.
Before that, anything could change; from here the following is frozen.

## Frozen: Go API

- `env` getters and their `(T, error)` signatures, fallback semantics
  (unset → fallback; set-but-empty is a value, except `Required`
  which errors), and error shapes (`*Error` with key/kind/quoted value,
  redacted for secrets; `*RequiredError`).
- `env.Var[T]` fields, `Parser[T]` signature, `VarOpt` options
  (`Secret()`, `Require()`), `Resolve()` semantics (empty counts as
  unset), `Any` interface methods, `Definition` JSON shape and the
  `{"version": 1, ...}` manifest envelope.
- `gocfg.LoadEnv(paths ...string)`: OS values never overridden, missing
  default `.env` is a no-op.
- Boolean sets, duration semantics (`time.ParseDuration`), slice
  splitting (comma, quote-aware, empties dropped), URL (scheme
  required), IP (`netip`).
- Naming note: `env.Required` is the required-value _getter_;
  `env.Require()` is the `Var` _option_. The pair is intentional and frozen.

## Frozen: CLI surface

- Commands and flags: `init`, `env`, `check`, `export`, `diff`,
  `doctor`, `validate`, `version` with their documented flags.
- Exit codes: `0` success, `1` configuration/validation failure,
  `2` CLI usage error (`doctor` always exits `0`).
- `.env` dialect: godotenv-compatible behavior with the deltas
  documented in `internal/dotenv` (ASCII keys, OS-first expansion,
  `\t` → tab, no `:` separator).
- File contracts: new `.env` files are `0600`; existing content,
  comments, order and permissions are preserved by `env`/`init`.

## Explicitly unstable

- `internal/...` packages: import at your own risk.
- Human output wording and styling (only exit codes are contractual).
- `examples/basic`: illustration, not API.
