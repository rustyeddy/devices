# Copilot Instructions for devices repository

## Repository Overview
This repository contains device interfaces, drivers, mocks/fakes, and shared utilities for Go-based hardware device abstractions. Keep orchestration/runtime logic out of this repo.

## Tech Stack
- **Language**: Go 1.24.5
- **Testing Framework**: testify (github.com/stretchr/testify v1.11.1)
- **Key Dependencies**: 
  - periph.io/x/conn/v3, periph.io/x/devices/v3, periph.io/x/host/v3 (hardware I/O)
  - github.com/warthog618/go-gpiocdev (GPIO)
  - github.com/maciej/bme280 (BME280 sensor)

## Essential Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=cover.out -cover ./...

# Run verbose tests
go test -v ./...

# View coverage report
go tool cover -func=cover.out
```

### Building
```bash
# Build the project
go build -v .

# Format code (required before committing)
gofmt -w .
```

### Makefile targets
```bash
make test      # Run tests with coverage
make build     # Build project
make fmt       # Format all Go files
make verbose   # Run tests with verbose output
make coverage  # Show coverage statistics
```

## Code Conventions

### Go Style
- **Always run `gofmt`** on all changed files before committing
- Follow standard Go naming conventions (MixedCaps for exported, mixedCaps for unexported)
- Keep interfaces minimal and stable
- Document all exported types, functions, and methods with godoc comments
- Use descriptive variable names; avoid single-letter names except for standard idioms (i, j for loops; r for io.Reader)

### Architecture Principles
- Drivers should be **deterministic** and testable without hardware
- Use fakes/mocks for testing instead of requiring real hardware
- Interfaces should be minimal and focused
- Keep device-specific logic separate from common utilities
- No orchestration or runtime logic belongs in this repository

## Testing Requirements

### Testing Framework: testify
- **Use `require`** for preconditions and must-pass checks (test stops on failure)
- **Use `assert`** for additional verification (test continues on failure)
- **Use `testify/mock`** sparingly; prefer simple fakes when possible

### Test Rules (CRITICAL)
- **Hermetic tests only**: No network, no serial/GPIO, no real hardware
- **No `time.Sleep`** for synchronization; use channels, contexts, or deterministic mocks
- **Use `t.TempDir()`** for any filesystem needs (auto-cleanup)
- **Goroutine safety**: If a goroutine is started, ensure deterministic shutdown via context/close channels
- **Table-driven tests** preferred for multiple test cases
- **Test edge cases explicitly**: zero values, nil inputs, error paths, boundary conditions
- **Run `t.Parallel()`** when tests are independent

### Test Priority (what to test first)
1. **Interface behavior**: Get/Set semantics, error handling, invariants
2. **Driver logic**: conversions, bounds checking, state transitions
3. **Serialization/encoding** if present
4. **Concurrency safety**: race conditions, shutdown behavior

### Example Test Structure
```go
func TestDeviceOperation(t *testing.T) {
    t.Parallel()
    
    tests := []struct {
        name    string
        input   int
        want    int
        wantErr bool
    }{
        {"normal case", 5, 10, false},
        {"zero value", 0, 0, false},
        {"error case", -1, 0, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := DoOperation(tt.input)
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            require.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

## Non-Goals
- **No integration tests** that require real hardware or external services
- **No breaking API changes** unless explicitly requested
- **No coupling to OttO runtime** or other orchestration systems

## Review Checklist
Before finalizing changes, verify:
- [ ] Are tests deterministic and fast?
- [ ] Any hidden dependencies on system state or hardware? (Remove them)
- [ ] Are mocks/fakes minimal and readable?
- [ ] No accidental coupling to OttO runtime concepts?
- [ ] All code formatted with `gofmt`?
- [ ] Tests pass: `go test ./...`
- [ ] No time.Sleep in tests?
- [ ] Goroutines shut down cleanly?

## Project Structure
```
/
├── base.go, device.go, descriptor.go, event.go    # Core interfaces
├── bme280/          # BME280 sensor driver
├── gpio/            # GPIO devices (LED, button, relay)
├── sensors/         # Sensor abstractions and polling
├── drivers/         # Hardware driver implementations
├── mock/            # Mock implementations for testing
├── examples/        # Example usage code
└── _archive/        # Archived/legacy code (do not modify)
```

## Boundaries and Restrictions
- **Do NOT modify** files in `_archive/` directory
- **Do NOT add** integration tests requiring hardware
- **Do NOT break** existing public APIs without explicit approval
- **Do NOT introduce** dependencies on orchestration/runtime systems
- **Do NOT use** real hardware in tests (use mocks/fakes)
- **Always maintain** backward compatibility unless breaking changes are requested

## Git Commit Guidance
Use small, focused commits with conventional commit prefixes:
- `test: <package> baseline` - Add or update tests
- `feat: add <device> driver` - New device support
- `fix: correct <issue> in <component>` - Bug fixes
- `docs: update godoc for <package>` - Documentation
- `refactor: simplify <component>` - Code refactoring
- `chore: update dependencies` - Maintenance

## Error Handling Patterns
- Return errors explicitly; don't panic in library code
- Wrap errors with context: `fmt.Errorf("operation failed: %w", err)`
- Check all error returns; don't ignore errors silently
- Use sentinel errors for expected error conditions

## Documentation Standards
- All exported identifiers must have godoc comments
- Start godoc comments with the name of the identifier
- Provide usage examples in doc comments for complex APIs
- Keep comments concise but informative

Example:
```go
// Device represents a generic hardware device with lifecycle management.
// Implementations must be safe for concurrent use.
type Device interface {
    // Name returns the unique identifier for this device.
    Name() string
    
    // Run starts the device operation. It blocks until ctx is canceled
    // or an unrecoverable error occurs.
    Run(ctx context.Context) error
}
```
