BINARY  := pgview
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"
OUTDIR  := bin

.PHONY: all build install clean test lint

all: build

## build: compile the binary into bin/
build:
	@mkdir -p $(OUTDIR)
	go build $(LDFLAGS) -o $(OUTDIR)/$(BINARY) .
	@echo "Built $(OUTDIR)/$(BINARY)"

## install: install binary to GOPATH/bin
install:
	go install $(LDFLAGS) .

## run: build and run interactively
run: build
	./$(OUTDIR)/$(BINARY)

## test: run unit tests
test:
	go test ./...

## lint: run golangci-lint (must be installed separately)
lint:
	golangci-lint run ./...

## clean: remove build artefacts
clean:
	rm -rf $(OUTDIR)

## tidy: tidy and verify go modules
tidy:
	go mod tidy
	go mod verify

help:
	@grep -E '^##' Makefile | sed 's/## //'

.DEFAULT_GOAL := build
