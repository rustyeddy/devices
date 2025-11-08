# devices
Some sensor type devices written in Go

## Platform Support

This library supports development on multiple platforms:

- **Linux (Raspberry Pi)**: Full GPIO hardware support via go-gpiocdev
- **macOS**: Stub implementation for development (no hardware access)
- **Windows**: Stub implementation for development (no hardware access)

The repository uses Go build tags to conditionally compile platform-specific code. This allows you to develop and build on macOS or Windows without requiring the Linux-specific GPIO dependencies.

See [PLATFORM_SUPPORT.md](PLATFORM_SUPPORT.md) for detailed information about cross-platform development.

## Building

```bash
# Build on any platform
go build ./...

# Run tests
go test ./...

# Cross-compile for Raspberry Pi
GOOS=linux GOARCH=arm GOARM=7 go build ./...
```
