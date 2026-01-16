# AGENTS.md — devices

> 📘 **Full instructions**: See [.github/copilot-instructions.md](.github/copilot-instructions.md) for detailed guidelines, examples, and best practices.

## Quick Reference

### Purpose
This repo contains device interfaces, drivers, mocks/fakes, and shared utilities for Go-based hardware abstractions. Keep orchestration/runtime logic out of this repo.

### Essential Commands
```bash
go test ./...           # Run all tests
gofmt -w .             # Format code (required)
make test              # Run tests with coverage
make build             # Build project
```

### Key Principles
- **Hermetic tests only**: No real hardware, network, or external dependencies
- **Always run `gofmt`** before committing
- **Keep interfaces minimal** and stable
- **Drivers must be deterministic** - use fakes/mocks for testing

## Testing Requirements

### Framework: testify
- Use `require` for must-pass checks (stops on failure)
- Use `assert` for additional checks (continues on failure)
- Avoid `testify/mock` - prefer simple fakes

### Critical Rules
- ❌ No `time.Sleep` - use channels/contexts instead
- ❌ No real hardware/GPIO/network access
- ✅ Use `t.TempDir()` for filesystem tests
- ✅ Use `t.Parallel()` for independent tests
- ✅ Ensure goroutines shut down cleanly (context/channels)
- ✅ Table-driven tests with edge cases

### Test Priority
1. Interface behavior (Get/Set, errors, invariants)
2. Driver logic (conversions, bounds, state)
3. Serialization/encoding
4. Concurrency safety (races, shutdown)

## Non-Goals
- ❌ No integration tests requiring real hardware
- ❌ No breaking API changes (unless requested)
- ❌ No coupling to OttO runtime

## Quick Checklist
Before committing:
- [ ] Tests pass: `go test ./...`
- [ ] Code formatted: `gofmt -w .`
- [ ] Tests are hermetic (no hardware/network)
- [ ] No `time.Sleep` in tests
- [ ] Goroutines shut down cleanly
- [ ] Mocks/fakes are minimal

## Commit Format
```
test: <pkg> baseline
feat: add <device> driver
fix: correct <issue> in <component>
docs: update godoc for <pkg>
```
