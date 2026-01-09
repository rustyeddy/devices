# AGENTS.md — devices

## Purpose
This repo contains device interfaces, drivers, mocks/fakes, and shared utilities.
Keep orchestration/runtime logic out of this repo.

## Non-goals
- No integration tests that require real hardware or external services.
- No breaking API changes unless explicitly requested.

## Go conventions
- Run `gofmt` on all changed files.
- Keep interfaces minimal and stable.
- Drivers should be deterministic; use fakes/mocks where needed.

## Testing (required)
Use **testify**:
- Prefer `require` for preconditions and must-pass checks.
- Use `assert` for additional verification.
- Use `testify/mock` sparingly; prefer simple fakes when possible.

Test rules:
- Hermetic tests only (no network, no serial/GPIO).
- No `time.Sleep` synchronization.
- Use `t.TempDir()` for any filesystem needs.
- If a goroutine is started in a test, ensure deterministic shutdown (context/close channels).
- Prefer table-driven tests and explicit edge cases (zero values, error paths).

Commands:
- Run: `go test ./...`

## What to test first (priority)
1. Interface behavior: Get/Set semantics, error handling, and invariants.
2. Driver logic: conversions, bounds, state transitions.
3. Serialization/encoding if present.
4. Concurrency safety (if any): races, shutdown behavior.

## Review checklist
- Are tests deterministic and fast?
- Any hidden dependencies on system state or hardware? Remove them.
- Are mocks/fakes minimal and readable?
- No accidental coupling to OttO runtime concepts.

## Commit guidance (if committing)
Small commits:
- `test: <pkg> baseline`
- `test: add fake <device>`
- `docs: godoc for <pkg>`
