.PHONY: ci-local test vet build frontend-build frontend-install help

BACKEND_DIR := backend
FRONTEND_DIR := frontend

## ci-local: run all checks that must pass before pushing (mirrors Gitea CI)
ci-local: vet test frontend-build
	@echo "✓ ci-local passed"

## test: run backend unit tests with race detector
test:
	@echo "→ backend tests"
	cd $(BACKEND_DIR) && go test -race -count=1 ./...

## vet: run go vet on the backend
vet:
	@echo "→ go vet"
	cd $(BACKEND_DIR) && go vet ./...

## build: compile the backend binary
build:
	@echo "→ backend build"
	cd $(BACKEND_DIR) && go build -o /dev/null ./...

## frontend-install: install frontend dependencies
frontend-install:
	cd $(FRONTEND_DIR) && npm install

## frontend-build: verify the frontend compiles cleanly
frontend-build:
	@echo "→ frontend build"
	cd $(FRONTEND_DIR) && npm run build --silent

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
