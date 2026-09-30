# Contributing

Keep it small: Go-first, explicit over magic, stdlib-only core.

## Workflow

Work happens in steps tracked in `PLAN.md`. After each step:

```bash
gofmt -l .
go vet ./...
go test ./... -count=1 -race
```

If all green: check the box in `PLAN.md`, commit, push.

## Rules

- Core packages (`env`, `internal/dotenv`, `internal/scaffold`, root)
  must stay stdlib-only. `cobra`/`lipgloss` live in `internal/cli` only —
  verify with `go list -deps ./env/... ./internal/dotenv/... ./internal/scaffold/... .`
- Typed getters return `(T, error)`; never silently fall back.
- Never log secret values; never override OS env in `LoadEnv`.
- Filesystem tests use `t.TempDir()`; env keys in tests must be unique
  per test (`OS > .env` precedence leaks across tests via process env).
- Commit messages: `feat(scope): ...`, `fix(scope): ...`, `docs: ...`,
  `chore: ...`.
