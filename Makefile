# Makefile for kcalc-add

SHELL    := /bin/bash

APP      ?= kcalc-add
CMDDIR   ?= ./cmd/$(APP)

BINARY ?= kcalc-add
BUILD_DIR ?= bin

PORT     ?= 8080
RUN_PID  ?= .run_server.pid
RUN_LOG  ?= .run_server.log

VERSION    ?= $(shell git describe --tags --dirty 2>/dev/null || echo 0.0.0-local)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE       ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BUILT_BY   ?= $(shell whoami)

# IMPORTANT: use your module path here. Currently it's 'kcalc-add'.
# If you later change module to 'github.com/YOURUSER/kcalc-add',
# update the package path prefix below accordingly.
LDFLAGS := -X 'main.Version=$(VERSION)' \
           -X 'main.Commit=$(COMMIT)'   \
           -X 'main.Date=$(DATE)'       \
           -X 'main.BuiltBy=$(BUILT_BY)'

GO         ?= go
PKG        ?= ./...
# Packages to instrument (comma-separated), excluding cmd/ and itest/
COVERPKG := $(shell $(GO) list ./... | grep -v '/cmd/' | grep -v '/itest$$' | tr '\n' ',' | sed 's/,$$//')
# Packages to run for coverage (space-separated), excluding cmd/ and itest/
PKG_COVER := $(shell $(GO) list ./... | grep -v '/cmd/' | grep -v '/itest$$')

# Coverage artifacts
COV_OUT    ?= coverage.out
COV_HTML   ?= coverage.html

# Integration test settings
IT_PORT    ?= 18080
IT_URL     ?= http://127.0.0.1:$(IT_PORT)
IT_PIDFILE ?= .itest_server.pid
IT_LOG     ?= .itest_server.log
IT_WAIT_MS ?= 5000       # total wait ~5s
IT_STEP_MS ?= 100

# Run Info

# Default target
.DEFAULT_GOAL := test

.PHONY: build test testv race cover cover-html bench fuzz fmt vet tidy lint ci ci-full clean help \
        run stop integration-test integration-clean build build-linux build-darwin print-version

build: ## Build the service binary (with version info)
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) $(CMDDIR)
	@echo "Built $(BUILD_DIR)/$(BINARY) (version $(VERSION), commit $(COMMIT))"

build-linux: ## Cross-compile linux/amd64
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-linux-amd64 $(CMDDIR)

build-darwin: ## Cross-compile darwin/arm64 (Apple Silicon)
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY)-darwin-arm64 $(CMDDIR)

print-version: ## Show resolved version metadata
	@echo VERSION=$(VERSION)
	@echo COMMIT=$(COMMIT)
	@echo DATE=$(DATE)
	@echo BUILT_BY=$(BUILT_BY)

test: ## Run unit tests
	$(GO) test -shuffle=on $(PKG)

testv: ## Run unit tests (verbose)
	$(GO) test -shuffle=on -v $(PKG)

race: ## Run tests with data race detector
	$(GO) test -shuffle=on -race $(PKG)

cover:
	$(GO) test -count=1 -covermode=atomic -coverpkg=$(COVERPKG) -coverprofile=$(COV_OUT) $(PKG_COVER)
	@$(GO) tool cover -func=$(COV_OUT)

cover-html: cover ## Create HTML coverage report
	@$(GO) tool cover -html=$(COV_OUT) -o $(COV_HTML)
	@echo "Wrote $(COV_HTML)"

run: build ## run service
	@set -euo pipefail; \
	if [ -f "$(RUN_PID)" ] && kill -0 "$$(cat "$(RUN_PID)")" 2>/dev/null; then \
	  echo "Already running (pid $$(cat "$(RUN_PID)"))"; exit 0; fi; \
	echo "Starting :$(PORT) ..."; \
	PORT="$(PORT)" ./$(BUILD_DIR)/$(BINARY) >"$(RUN_LOG)" 2>&1 & echo $$! >"$(RUN_PID)"; \
	echo "pid $$(cat "$(RUN_PID)") (logs: $(RUN_LOG))"

stop: ## stop service
	@set -euo pipefail; \
	if [ -f "$(RUN_PID)" ]; then \
	  echo "Stopping pid $$(cat "$(RUN_PID)") ..."; \
	  kill -TERM "$$(cat "$(RUN_PID)")" 2>/dev/null || true; \
	  wait "$$(cat "$(RUN_PID)")" 2>/dev/null || true; \
	  rm -f "$(RUN_PID)"; \
	else \
	  echo "No pid file."; \
	fi

integration-test: ## Run integration tests against a live server
	@set -euo pipefail; \
	command -v curl >/dev/null 2>&1 || { echo "curl is required for integration tests"; exit 1; } ; \
	trap 'rc=$$?; echo "Stopping server..."; \
	      if [ -f $(IT_PIDFILE) ]; then kill -TERM $$(cat $(IT_PIDFILE)) 2>/dev/null || true; \
	      wait $$(cat $(IT_PIDFILE)) 2>/dev/null || true; fi; \
	      exit $$rc' EXIT INT TERM; \
	echo "Starting server on :$(IT_PORT) ..."; \
	PORT=$(IT_PORT) $(GO) run $(CMDDIR) >"$(IT_LOG)" 2>&1 & echo $$! > "$(IT_PIDFILE)"; \
	echo "Waiting for readiness at $(IT_URL)/add ..."; \
	limit=$$(( $(IT_WAIT_MS) / $(IT_STEP_MS) )); \
	for i in $$(seq 1 $$limit); do \
		if curl -sSf -X POST "$(IT_URL)/add" -H 'Content-Type: application/json' -d '{"a":0,"b":0}' >/dev/null 2>&1; then \
			echo "Server is up."; break; \
		fi; \
		sleep $$(( $(IT_STEP_MS) ))e-3; \
	done; \
	# Fail fast if not up
	if ! curl -sSf -X POST "$(IT_URL)/add" -H 'Content-Type: application/json' -d '{"a":1,"b":2}' >/dev/null 2>&1; then \
		echo "Server failed to start. Last 50 log lines:"; \
		tail -n 50 "$(IT_LOG)" || true; \
		exit 1; \
	fi; \
	echo "Running integration tests..."; \
	BASE_URL="$(IT_URL)" $(GO) test -v -count=1 ./itest

integration-clean: ## Clean integration artifacts
	@rm -f "$(IT_PIDFILE)"

bench: ## Run benchmarks (if any)
	$(GO) test -bench=. -benchmem $(PKG)

## Limit to ~10s so it doesn't run forever in CI
fuzz: ## Run fuzzing for functions named Fuzz* (example: FuzzAdd)
	$(GO) test -run=^$$ -fuzz=Fuzz -fuzztime=10s $(PKG)

fmt: ## Format code
	$(GO) fmt $(PKG)

vet: ## Static analysis
	$(GO) vet $(PKG)

tidy: ## Sync go.mod/go.sum
	$(GO) mod tidy

lint: ## Lint if golangci-lint is installed; otherwise print a hint
	@set -e; \
	if command -v golangci-lint >/dev/null 2>&1; then \
	  golangci-lint run; \
	else \
	  echo "golangci-lint not found. Install: https://golangci-lint.run/"; \
	fi

ci: tidy fmt vet lint race cover cover-html ## CI bundle: tidy, fmt, vet, lint, race tests, cover, cover-html

ci-full: ## CI bundle: tidy, fmt, vet, lint, race, cover, cover-html, integration tests (with cleanup)
	@set -euo pipefail; \
	$(MAKE) ci; \
	rc=0; \
	$(MAKE) integration-test || rc=$$?; \
	$(MAKE) integration-clean; \
	exit $$rc

clean: ## Clean generated artifacts
	@rm -fr "$(BUILD_DIR)"
	@rm -f "$(COV_OUT)" "$(COV_HTML)" "$(IT_PIDFILE)" "$(IT_LOG)" "$(RUN_LOG)"

help: ## Show this help
	@awk 'BEGIN{FS=":.*##"; printf "Targets:\n"} /^[a-zA-Z0-9_.-]+:.*##/{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
