# Agentglobe (Minibook-parity Go API). Use GOWORK=off when a parent go.work is broken.
# Override local config: make run-local-build LOCAL_CONFIG=../dep/other.yaml
LOCAL_CONFIG ?= ../dep/cf.yaml

# Clawgotcha (agent metadata HTTP API). Config YAML: database_url, port, etc.
# Override: make run-clawgotcha CL_LOCAL_CONFIG=../dep/other.yaml
CL_LOCAL_CONFIG ?= ../dep/cl.yaml

# Database DDL (Postgres: pg_dump, fallback docker; SQLite: sqlite_master). Uses LOCAL_CONFIG.
SCHEMA_OUT ?= spec/agentglobe_schema.sql

.DEFAULT_GOAL := help

.PHONY: help build build-all run run-local-build \
	build-clawgotcha run-clawgotcha build-newsapi build-worldmon build-feed-digest build-af-local-mcp \
	schema-export migrate test tidy vet lint

help:
	@echo "agentbook Makefile — set LOCAL_CONFIG / CL_LOCAL_CONFIG to point at your YAML."
	@echo ""
	@echo "  make build              Build agentglobe → bin/agentglobe"
	@echo "  make build-all          Build every app in bin/ (see targets below)"
	@echo "  make run                Build and run agentglobe (CONFIG via LOCAL_CONFIG, def. ../dep/cf.yaml)"
	@echo "  make build-clawgotcha  clawgotcha → bin/clawgotcha"
	@echo "  make build-newsapi      newsapi     → bin/newsapi"
	@echo "  make build-worldmon     worldmon    → bin/worldmon"
	@echo "  make build-feed-digest  worldmon    → bin/feed-digest"
	@echo "  make build-af-local-mcp  agentglobe → bin/af-local-mcp"
	@echo "  make run-clawgotcha    Build and run clawgotcha (CONFIG via CL_LOCAL_CONFIG, def. ../dep/cl.yaml)"
	@echo "  make schema-export      Write DB schema to SCHEMA_OUT (def. spec/agentglobe_schema.sql)"
	@echo "  make migrate            Apply spec/migrations (Postgres) via agentglobe migrate"
	@echo "  make test|tidy|vet|lint  Go test / mod tidy / vet / golangci-lint in agentglobe"
	@echo ""

build:
	mkdir -p bin
	cd agentglobe && GOWORK=off go build -o ../bin/agentglobe ./cmd/agentglobe

build-all: build build-clawgotcha build-newsapi build-worldmon build-feed-digest build-af-local-mcp

run: build
	cd agentglobe && CONFIG_PATH=$(LOCAL_CONFIG) ../bin/agentglobe

build-clawgotcha:
	mkdir -p bin
	cd clawgotcha && GOWORK=off go build -o ../bin/clawgotcha ./cmd/clawgotcha

build-newsapi:
	mkdir -p bin
	cd newsapi && GOWORK=off go build -o ../bin/newsapi ./cmd/server

build-worldmon:
	mkdir -p bin
	cd worldmon && GOWORK=off go build -o ../bin/worldmon ./cmd/server

build-feed-digest:
	mkdir -p bin
	cd worldmon && GOWORK=off go build -o ../bin/feed-digest ./cmd/feed-digest

build-af-local-mcp:
	mkdir -p bin
	cd agentglobe && GOWORK=off go build -o ../bin/af-local-mcp ./cmd/af-local-mcp

run-clawgotcha: build-clawgotcha
	cd clawgotcha && CONFIG_PATH=$(CL_LOCAL_CONFIG) ../bin/clawgotcha

schema-export:
	cd agentglobe && GOWORK=off CONFIG_PATH=$(LOCAL_CONFIG) go run ./cmd/schemaexport -out ../$(SCHEMA_OUT)

# Postgres only: applies spec/migrations/*.sql in lexical order once each (tracks public.schema_migrations).
# Uses database_url from LOCAL_CONFIG (default ../dep/cf.yaml) via migrate -c; DATABASE_URL env still overrides YAML (config.Load).
MIGRATE_DIR ?= ../spec/migrations
migrate:
	cd agentglobe && GOWORK=off go run ./cmd/migrate -c $(LOCAL_CONFIG) -d $(MIGRATE_DIR)

test:
	cd agentglobe && GOWORK=off go test ./...

tidy:
	cd agentglobe && GOWORK=off go mod tidy

vet:
	cd agentglobe && GOWORK=off go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && (cd agentglobe && GOWORK=off golangci-lint run ./...) || echo "golangci-lint not installed; skipping lint"
