# code-minions Makefile — local dev workflow
# Run 'make help' to see available targets.

BINARY     := code-minions
CMD_PKG    := ./cmd/code-minions
COVER_OUT  := coverage.out

# Cross-platform delete:
# - On Windows (OS=Windows_NT), use cmd so that 'del' is available even under POSIX shells.
# - On other systems, use standard rm -f.
ifeq ($(OS),Windows_NT)
    RM = cmd /C del /q 2>nul
else
    RM = rm -f
endif

.DEFAULT_GOAL := build

.PHONY: build test lint fmt snapshot install coverage check-coverage clean help

build: ## Compile the binary
	go build $(CMD_PKG)

test: ## Run all tests with coverage summary
	go test ./... -cover

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format all Go source files
	gofmt -w .

snapshot: ## Run GoReleaser in snapshot mode (no publish)
	goreleaser release --snapshot --clean --skip=publish

install: ## Install the binary to $GOPATH/bin
	go install $(CMD_PKG)

coverage: ## Generate coverage profile and open HTML report
	go test -coverprofile=$(COVER_OUT) ./...
	go tool cover -html=$(COVER_OUT)

check-coverage: ## Run the coverage threshold check (default: 70%)
	./scripts/check-coverage.sh

clean: ## Remove build artifacts and coverage files
	-$(RM) $(BINARY) $(BINARY).exe $(COVER_OUT)

help: ## Show this help message
	@echo Available targets:
	@echo   build             Compile the binary
	@echo   test              Run all tests with coverage summary
	@echo   lint              Run golangci-lint
	@echo   fmt               Format all Go source files
	@echo   snapshot          Run GoReleaser in snapshot mode (no publish)
	@echo   install           Install the binary to GOPATH/bin
	@echo   coverage          Generate coverage profile and open HTML report
	@echo   check-coverage    Run the coverage threshold check (default 70%%)
	@echo   clean             Remove build artifacts and coverage files
	@echo   help              Show this help message
