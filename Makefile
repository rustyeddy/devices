# PIENV	= env GOOS=linux GOARCH=arm GOARM=7

# all: test $(SUBDIRS) build

# init:
# 	git update --init 

# build:
# 	rm -f garden-station
# 	go build -v . 

# fmt:
# 	gofmt -w .

# test:
# 	rm -f cover.out
# 	go test -coverprofile=cover.out -cover ./...

# verbose:
# 	rm -f cover.out
# 	go test -v -coverprofile=cover.out -cover ./...

# coverage: test
# 	go tool cover -func=cover.out

# html: test
# 	rm -f coverage.html
# 	go tool cover -html=cover.out -o coverage.html

# .PHONY: all test build fmt $(SUBDIRS)


# Makefile for github.com/rustyeddy/devices
#
# Builds example binaries for:
#   - linux/amd64 (dev/CI, uses stubbed drivers where appropriate)
#   - linux/arm64 or linux/arm (Raspberry Pi, uses real periph-backed drivers)
#
# Usage:
#   make build-host
#   make build-pi PI_ARCH=arm64
#   make build-pi PI_ARCH=arm
#   make test
#   make clean

SHELL := /bin/bash

GO        ?= go
BIN_DIR   ?= bin

# Examples we build into binaries (expects ./examples/<name>/main.go)
EXAMPLES  ?= led relay button vh400

# Version stamping (optional)
GIT_SHA   := $(shell git rev-parse --short HEAD 2>/dev/null || echo "nogit")
BUILD_TS  := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Common build flags
GOFLAGS   ?=
LDFLAGS   ?= -s -w \
  -X 'main.buildSHA=$(GIT_SHA)' \
  -X 'main.buildTime=$(BUILD_TS)'

# Default Pi arch. Use PI_ARCH=arm for 32-bit Pi OS.
PI_ARCH   ?= arm64

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build-host           Build examples for linux/amd64 (dev/CI; uses stubs where needed)"
	@echo "  build-pi             Build examples for linux/$(PI_ARCH) (Raspberry Pi)"
	@echo "  test                 Run unit tests"
	@echo "  vet                  Run go vet"
	@echo "  tidy                 Run go mod tidy"
	@echo "  clean                Remove ./$(BIN_DIR)"
	@echo ""
	@echo "Vars:"
	@echo "  EXAMPLES=...          Space-separated example list (default: $(EXAMPLES))"
	@echo "  PI_ARCH=arm64|arm     Raspberry Pi target arch (default: $(PI_ARCH))"
	@echo "  BIN_DIR=...           Output directory (default: $(BIN_DIR))"
	@echo ""
	@echo "Examples:"
	@echo "  make build-host"
	@echo "  make build-pi PI_ARCH=arm64"
	@echo "  make build-pi PI_ARCH=arm"

# -------------------------
# Host build (linux/amd64)
# -------------------------
.PHONY: build-host
build-host: $(addprefix build-host-,$(EXAMPLES))

build-host-%:
	@mkdir -p "$(BIN_DIR)/host"
	@echo "==> build host: $*"
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
		-o "$(BIN_DIR)/host/$*" "./examples/$*"

# -------------------------
# Raspberry Pi build
# -------------------------
.PHONY: build-pi
build-pi: $(addprefix build-pi-,$(EXAMPLES))

build-pi-%:
	@mkdir -p "$(BIN_DIR)/pi-$(PI_ARCH)"
	@echo "==> build pi ($(PI_ARCH)): $*"
	GOOS=linux GOARCH=$(PI_ARCH) CGO_ENABLED=0 \
		$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
		-o "$(BIN_DIR)/pi-$(PI_ARCH)/$*" "./examples/$*"

# -------------------------
# Quality targets
# -------------------------
.PHONY: test
test:
	$(GO) test $(GOFLAGS) ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: clean
clean:
	rm -rf "$(BIN_DIR)"
