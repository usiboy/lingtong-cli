# Copyright (c) 2026 Lingtong
# SPDX-License-Identifier: MIT

BINARY   := lingtong-cli
MODULE   := github.com/lingtong/cli
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE     := $(shell date +%Y-%m-%d)
LDFLAGS  := -s -w -X $(MODULE)/internal/build.Version=$(VERSION) -X $(MODULE)/internal/build.Date=$(DATE)
PREFIX   ?= /usr/local

.PHONY: build vet test unit-test integration-test install uninstall clean gen-docs check-docs

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

vet:
	go vet ./...

unit-test:
	go test -race -gcflags="all=-N -l" -count=1 ./cmd/... ./internal/... ./shortcuts/...

integration-test: build
	go test -v -count=1 ./tests/...

test: vet unit-test integration-test

# Generate markdown documentation for all CLI commands
gen-docs:
	go run cmd/gen-docs/main.go ./doc/commands
	@echo "Documentation generated in ./doc/commands"

# Check if documentation is up to date
check-docs: build
	@echo "Checking documentation..."
	@rm -rf /tmp/lingtong-cli-docs-check
	@mkdir -p /tmp/lingtong-cli-docs-check
	@./$(BINARY) gen-docs /tmp/lingtong-cli-docs-check 2>/dev/null || \
		go run cmd/gen-docs/main.go /tmp/lingtong-cli-docs-check
	@if diff -r doc/commands /tmp/lingtong-cli-docs-check > /dev/null 2>&1; then \
		echo "OK: Documentation is up to date"; \
	else \
		echo "WARNING: Documentation is out of date. Run 'make gen-docs' to update."; \
		diff -r doc/commands /tmp/lingtong-cli-docs-check || true; \
		rm -rf /tmp/lingtong-cli-docs-check; \
		exit 1; \
	fi
	@rm -rf /tmp/lingtong-cli-docs-check

install: build
	install -d $(PREFIX)/bin
	install -m755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "OK: $(PREFIX)/bin/$(BINARY) ($(VERSION))"

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)
