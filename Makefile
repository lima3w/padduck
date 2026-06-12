.PHONY: ci-local test test-integration vet staticcheck gosec govulncheck go-analysis-tools build frontend-build frontend-install frontend-lint frontend-test frontend-audit check-migrations sbom help

BACKEND_DIR := backend
FRONTEND_DIR := frontend
MIGRATIONS_DIR := backend/migrations
GO_BIN := $(shell go env GOPATH)/bin
STATICCHECK := $(GO_BIN)/staticcheck
GOSEC := $(GO_BIN)/gosec
GOVULNCHECK := $(GO_BIN)/govulncheck
STATICCHECK_VERSION ?= v0.7.0
GOSEC_VERSION ?= v2.26.1
GOVULNCHECK_VERSION ?= v1.3.0
GOVULNCHECK_MIN_GO ?= go1.26.4
STATICCHECK_CHECKS := all,-U1000,-ST1000,-ST1003,-ST1020,-SA1019

## ci-local: run all checks that must pass before pushing (mirrors GitHub CI)
ci-local: check-migrations vet staticcheck gosec govulncheck test frontend-install frontend-audit frontend-lint frontend-test frontend-build
	@echo "✓ ci-local passed"

## test: run backend unit tests with race detector
test:
	@echo "→ backend tests"
	cd $(BACKEND_DIR) && go test -mod=vendor -race -count=1 ./...

## test-integration: run DB-backed tests against a throwaway Postgres container
test-integration:
	@echo "→ backend DB integration tests (throwaway Postgres)"
	docker run -d --rm --name padduck-test-pg -e POSTGRES_PASSWORD=test -p 127.0.0.1:55432:5432 postgres:18 >/dev/null
	@trap 'docker stop padduck-test-pg >/dev/null' EXIT; \
	timeout 60 sh -c 'until docker exec padduck-test-pg pg_isready -U postgres -q; do sleep 1; done'; \
	cd $(BACKEND_DIR) && TEST_DATABASE_URL="postgres://postgres:test@127.0.0.1:55432/postgres?sslmode=disable" \
		go test -mod=vendor -race -count=1 ./...

## check-migrations: verify migration files use paired, single-direction sql-migrate annotations
check-migrations:
	@echo "→ migration annotations"
	@bad_up=$$(grep -rL '^\-\- +migrate Up' $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null); \
	bad_down=$$(grep -rL '^\-\- +migrate Down' $(MIGRATIONS_DIR)/*.down.sql 2>/dev/null); \
	if [ -n "$$bad_up" ] || [ -n "$$bad_down" ]; then \
		echo "ERROR: missing sql-migrate annotation in:"; \
		echo "$$bad_up $$bad_down" | tr ' ' '\n' | grep -v '^$$'; \
		exit 1; \
	fi; \
	mixed_up=$$(grep -l '^\-\- +migrate Down' $(MIGRATIONS_DIR)/*.up.sql 2>/dev/null); \
	mixed_down=$$(grep -l '^\-\- +migrate Up' $(MIGRATIONS_DIR)/*.down.sql 2>/dev/null); \
	if [ -n "$$mixed_up" ] || [ -n "$$mixed_down" ]; then \
		echo "ERROR: migration files must not mix up/down sections:"; \
		echo "$$mixed_up $$mixed_down" | tr ' ' '\n' | grep -v '^$$'; \
		exit 1; \
	fi

## vet: run go vet on the backend
vet:
	@echo "→ go vet"
	cd $(BACKEND_DIR) && go vet -mod=vendor ./...

## go-analysis-tools: install Go static analysis tools used by CI
go-analysis-tools:
	@echo "→ Go analysis tools"
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)
	go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

## staticcheck: run staticcheck on Go modules
staticcheck:
	@echo "→ staticcheck backend"
	cd $(BACKEND_DIR) && GOFLAGS=-mod=vendor $(STATICCHECK) -checks=$(STATICCHECK_CHECKS) ./...
	@echo "→ staticcheck agent"
	cd agent && $(STATICCHECK) ./...

## gosec: run gosec on Go modules
gosec:
	@echo "→ gosec backend"
	cd $(BACKEND_DIR) && GOFLAGS=-mod=vendor $(GOSEC) ./...
	@echo "→ gosec agent"
	cd agent && $(GOSEC) ./...

## govulncheck: run govulncheck on Go modules
govulncheck:
	@v=$$(go env GOVERSION); \
	if [ "$$(printf '%s\n%s\n' "$(GOVULNCHECK_MIN_GO)" "$$v" | sort -V | head -n1)" != "$(GOVULNCHECK_MIN_GO)" ]; then \
		echo "ERROR: govulncheck requires Go $(GOVULNCHECK_MIN_GO) or newer; found $$v"; \
		exit 1; \
	fi
	@echo "→ govulncheck backend"
	cd $(BACKEND_DIR) && GOFLAGS=-mod=vendor $(GOVULNCHECK) ./...
	@echo "→ govulncheck agent"
	cd agent && $(GOVULNCHECK) ./...

## build: compile the backend binary
build:
	@echo "→ backend build"
	cd $(BACKEND_DIR) && go build -mod=vendor -o /dev/null ./...

## frontend-install: install frontend dependencies
frontend-install:
	cd $(FRONTEND_DIR) && npm ci

## frontend-lint: run ESLint on the frontend source
frontend-lint:
	@echo "→ frontend lint"
	cd $(FRONTEND_DIR) && npm run lint

## frontend-test: run frontend unit tests (non-watch mode)
frontend-test:
	@echo "→ frontend tests"
	cd $(FRONTEND_DIR) && npm run test:coverage

## frontend-build: verify the frontend compiles cleanly
frontend-build:
	@echo "→ frontend build"
	cd $(FRONTEND_DIR) && npm run build --silent

## frontend-audit: scan frontend production dependencies for known vulnerabilities
frontend-audit:
	@echo "→ npm audit"
	cd $(FRONTEND_DIR) && npm audit --omit=dev --audit-level=high

## sbom: generate dependency SBOM from vendored Go modules and npm lockfile
sbom:
	@echo "→ dependency SBOM"
	node tools/generate-sbom.mjs

## help: list available targets
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
