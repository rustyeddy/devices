# Platform Support

This document explains how the GPIO driver implementation works across different platforms.

## Overview

The devices library uses GPIO hardware on Linux (specifically Raspberry Pi), but also supports development on non-Linux platforms (macOS, Windows) through stub implementations.

## Build Tags

The codebase uses Go build tags to conditionally compile platform-specific code:

- **Linux**: Uses actual GPIO hardware via the `go-gpiocdev` library
- **Non-Linux** (macOS, Windows, etc.): Uses stub implementations that simulate GPIO functionality

## Architecture

### Drivers Package

The `drivers` package is split into three files:

1. **driver_gpio_common.go**: Platform-agnostic code including:
   - GPIO struct and GetGPIO() singleton
   - DigitalPin struct with common methods (Get, Set, Toggle, etc.)
   - Line interface definition

2. **driver_gpio_linux.go** (`//go:build linux`): Linux-specific implementation
   - Imports `github.com/warthog618/go-gpiocdev`
   - Implements actual GPIO hardware access
   - Provides lineWrapper to adapt gpiocdev.Line to our Line interface
   - Provides MockLine for testing on Linux

3. **driver_gpio_stub.go** (`//go:build !linux`): Stub implementation for non-Linux
   - Provides stub types (LineReqOption, EventHandler, etc.)
   - MockLine always used (logs warning about stub mode)
   - Provides compatibility functions (AsOutput, AsInput, WithPullUp, etc.)

### Device Packages

Several device packages (button, relay, led) also have platform-specific implementations:

#### Button Package
- **button_linux.go**: Uses gpiocdev types for event handling
- **button_stub.go**: Stub implementation for development on non-Linux platforms
- **button_test.go**: Test file with Linux build tag

#### Relay Package
- **relay_linux.go**: Linux implementation with gpiocdev
- **relay_stub.go**: Stub implementation for non-Linux platforms

#### LED Package
- **led_linux.go**: Linux implementation with gpiocdev
- **led_stub.go**: Stub implementation for non-Linux platforms

### Examples

Examples that use GPIO directly also have platform-specific versions:
- **examples/relay/main_linux.go**: Linux-only example
- **examples/switch/main_linux.go**: Linux-only example

## Building

### On Linux
```bash
go build ./...
go test ./...
```

### On macOS
```bash
GOOS=darwin go build ./...
```

### On Windows
```bash
GOOS=windows go build ./...
```

### Cross-compilation
Build for Linux from any platform:
```bash
GOOS=linux GOARCH=arm GOARM=7 go build ./...
```

## Development Workflow

1. **Develop on any platform**: The stub implementations allow you to build and test basic functionality on macOS or Windows
2. **Test on Linux**: Deploy to actual Raspberry Pi hardware for full GPIO testing
3. **Mock mode**: Even on Linux, you can enable mock mode by calling `devices.SetMock(true)` to test without actual hardware

## Testing

Tests run on Linux will use the actual GPIO implementation (or mock if enabled).
On non-Linux platforms, tests will use stub implementations automatically.

## Adding New GPIO Features

When adding features that use GPIO:

1. Add common functionality to `driver_gpio_common.go`
2. Add Linux-specific code to `driver_gpio_linux.go` with gpiocdev imports
3. Add stub code to `driver_gpio_stub.go` for compatibility
4. Consider using build tags for device packages if they use gpiocdev types directly

## Troubleshooting

### "undefined: gpiocdev" error
This means a file is importing gpiocdev but doesn't have the `//go:build linux` tag.

### Build fails on non-Linux
Check that all files importing `github.com/warthog618/go-gpiocdev` have the `//go:build linux` build tag.

### Tests fail on non-Linux
Ensure test files that use gpiocdev types also have the `//go:build linux` build tag.
