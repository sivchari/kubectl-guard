GOLANGCI_LINT = go tool -modfile tools/go.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint

.PHONY: build
build: ## Build the binary
	go build -o kubectl-guard ./cmd/kubectl-guard

.PHONY: install
install: ## Install the binary to GOPATH/bin
	go install ./cmd/kubectl-guard

.PHONY: test
test: ## Run tests
	go test -race -shuffle=on -v ./...

.PHONY: lint
lint: ## Run linter
	${GOLANGCI_LINT} run ./... --timeout 5m

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	${GOLANGCI_LINT} run ./... --timeout 5m --fix

.PHONY: fmt
fmt: ## Format code
	${GOLANGCI_LINT} fmt ./...

.PHONY: clean
clean: ## Clean build artifacts
	rm -f kubectl-guard
	rm -rf dist/

.PHONY: release-dry-run
release-dry-run: ## Run goreleaser in dry-run mode
	goreleaser release --snapshot --clean

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
