# Go-Calc developer commands.
#
#   make build   compile the binary into bin/go-calc
#   make dev     run the app with live reload (mygo dev)
#   make test    run all unit and UI tests
#   make help    list the targets

APP      := go-calc
BIN_DIR  := bin
MYGO     := go tool mygo

.DEFAULT_GOAL := help
.PHONY: help build dev test vet clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z_-]+:.*## / {printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Compile the app into bin/go-calc
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP) .

dev: ## Run the app with live reload (mygo dev)
	$(MYGO) dev

test: ## Run all tests
	go test ./...

vet: ## Run go vet on all packages
	go vet ./...

clean: ## Remove build outputs
	rm -rf $(BIN_DIR)
