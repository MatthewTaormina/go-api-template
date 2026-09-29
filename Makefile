# ── go-api-template Makefile ──────────────────────────────────────────────
# Targets designed to be composable: "make test" is the primary CI gate.
# All commands use the full go binary path if needed; adjust GOBIN for your env.

BINARY     := bin/api
MAIN       := ./cmd/api
GO         := go

.PHONY: run build test lint fmt vet clean help

## run: Start the API server locally (uses env vars or defaults)
run:
	$(GO) run $(MAIN)

## build: Compile and output to bin/api
build:
	$(GO) build -o $(BINARY) $(MAIN)

## test: Run all tests with the race detector
test:
	$(GO) test ./... -v -race -count=1

## vet: Run go vet (catches common correctness issues)
vet:
	$(GO) vet ./...

## fmt: Format all source files with gofmt
fmt:
	gofmt -w .

## lint: Run vet + format check (useful in CI)
lint: vet
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not gofmt'd:"; \
		gofmt -l .; \
		exit 1; \
	fi

## clean: Remove build artefacts
clean:
	rm -rf bin/

## help: Show this help message
help:
	@grep -E '^## ' Makefile | sed 's/## //'
