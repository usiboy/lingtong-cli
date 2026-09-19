# Copyright (c) 2026 Lingtong
# SPDX-License-Identifier: MIT

BINARY   := lingtong-cli
MODULE   := github.com/lingtong/cli
VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATE     := $(shell date +%Y-%m-%d)
LDFLAGS  := -s -w -X $(MODULE)/internal/build.Version=$(VERSION) -X $(MODULE)/internal/build.Date=$(DATE)
PREFIX   ?= /usr/local

.PHONY: build vet test unit-test integration-test install install-skills uninstall clean gen-docs check-docs

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

vet:
	go vet ./...

unit-test:
	go test -race -count=1 ./cmd/... ./internal/... ./shortcuts/...

integration-test: build
	./$(BINARY) --help >/dev/null
	./$(BINARY) --version >/dev/null

test: vet unit-test integration-test

# Generate markdown documentation for all CLI commands
gen-docs:
	go run cmd/gen-docs/main.go ./doc/commands
	@echo "Documentation generated in ./doc/commands"

# Check if documentation is up to date
check-docs: build
	@echo "Checking documentation generation..."
	@rm -rf /tmp/lingtong-cli-docs-check
	@mkdir -p /tmp/lingtong-cli-docs-check
	@./$(BINARY) gen-docs /tmp/lingtong-cli-docs-check 2>/dev/null || \
		go run cmd/gen-docs/main.go /tmp/lingtong-cli-docs-check
	@test -s /tmp/lingtong-cli-docs-check/lingtong-cli.md
	@echo "OK: Documentation generation succeeded"
	@rm -rf /tmp/lingtong-cli-docs-check

install: build
	install -d $(PREFIX)/bin
	install -m755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	@echo "OK: $(PREFIX)/bin/$(BINARY) ($(VERSION))"

# Install the embedded skills into detected AI editors (Claude Code, OpenCode,
# Qoder, Cursor, Trae, Codex). Forwards extra args, e.g. ARGS="--scope global".
install-skills: build
	./$(BINARY) skills install $(ARGS)

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)

clean:
	rm -f $(BINARY)
