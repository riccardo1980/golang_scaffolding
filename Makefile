BINARY := bin/agent

.PHONY: all build run test fmt vet check clean help

all: check build

build: ## Build the agent binary into bin/agent
	go build -o $(BINARY) ./cmd/agent

run: ## Run the agent CLI demo
	go run ./cmd/agent

test: ## Run all tests with the race detector
	go test -race -v ./...

fmt: ## Format Go source files
	gofmt -w .

vet: ## Run go vet across all packages
	go vet ./...

check: vet test ## Run formatting check, go vet, and tests
	@test -z "$$(gofmt -l .)" || (echo "Unformatted files:"; gofmt -l .; exit 1)

clean: ## Remove build artifacts
	rm -rf bin/

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'
