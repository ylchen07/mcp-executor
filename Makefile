GOCMD=go
GOCACHE_DIR?=$(CURDIR)/.cache/go-build
BIN_DIR?=$(CURDIR)/bin
COVERAGE_DIR?=$(CURDIR)/coverage
BINARY_NAME?=mcp-executor
GOTEST=CGO_ENABLED=0 GOCACHE=$(GOCACHE_DIR) $(GOCMD) test
GOTIDY=GOCACHE=$(GOCACHE_DIR) $(GOCMD) mod tidy
GOBUILD=CGO_ENABLED=0 GOCACHE=$(GOCACHE_DIR) $(GOCMD) build
GORUN=GOCACHE=$(GOCACHE_DIR) $(GOCMD) run main.go
GOLANGCI_LINT?=golangci-lint
LINT_ENV=CGO_ENABLED=0 XDG_CACHE_HOME=$(CURDIR)/.cache GOLANGCI_LINT_CACHE=$(CURDIR)/.cache/golangci

.PHONY: deps fmt lint test test-verbose test-coverage build run clean help

help:
	@echo "Available targets:"
	@echo "  make deps           - Tidy Go dependencies"
	@echo "  make fmt            - Format Go code"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make test           - Run tests with verbose output (no cache)"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make build          - Build binary to bin/$(BINARY_NAME)"
	@echo "  make run            - Run the application"
	@echo "  make clean          - Remove build artifacts and cache"

deps:
	$(GOTIDY)

fmt:
	$(GOCMD) fmt ./...

lint:
	$(LINT_ENV) $(GOLANGCI_LINT) run ./...

test:
	$(GOTEST) -v -count=1 ./...

test-coverage: | $(COVERAGE_DIR)
	$(GOTEST) -v -count=1 -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/coverage.html"

build: | $(BIN_DIR)
	$(GOBUILD) -o $(BIN_DIR)/$(BINARY_NAME) .

run:
	$(GORUN)

clean:
	rm -rf $(BIN_DIR) $(CURDIR)/.cache $(COVERAGE_DIR)

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

$(COVERAGE_DIR):
	mkdir -p $(COVERAGE_DIR)
