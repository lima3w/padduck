.PHONY: ci-local test vet build frontend-build frontend-install check-migrations help

BACKEND_DIR := backend
FRONTEND_DIR := frontend
MIGRATIONS_DIR := backend/migrations

## ci-local: run all checks that must pass before pushing (mirrors Gitea CI)
ci-local: check-migrations vet test frontend-build
	@echo "✓ ci-local passed"

## test: run backend unit tests with race detector
test:
	@echo "→ backend tests"
	cd $(BACKEND_DIR) && go test -race -count=1 ./...

## check-migrations: verify every migration file has the required sql-migrate annotation
check-migrations:
	@echo "→ migration annotations"
	@bad_up=$$(grep -rL '^\-\- +migrate Up' $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null); \
	bad_down=$$(grep -rL '^\-\- +migrate Down' $(MIGRATIONS_DIR)/*.down.sql 2>/dev/null); \
	if [ -n "$$bad_up" ] || [ -n "$$bad_down" ]; then \
		echo "ERROR: missing sql-migrate annotation in:"; \
		echo "$$bad_up $$bad_down" | tr ' ' '\n' | grep -v '^$$'; \
		exit 1; \
	fi

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
