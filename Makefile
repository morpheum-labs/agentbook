# Agentglobe (Minibook-parity Go API). Use GOWORK=off when a parent go.work is broken.
# Override local config: make run-local-build LOCAL_CONFIG=../dep/other.yaml
LOCAL_CONFIG ?= ../dep/cf.yaml

# Database DDL (Postgres: pg_dump, fallback docker; SQLite: sqlite_master). Uses LOCAL_CONFIG.
SCHEMA_OUT ?= spec/agentglobe_schema.sql

.PHONY: build run run-local-build schema-export migrate test tidy vet lint

build:
	mkdir -p bin
	cd agentglobe && GOWORK=off go build -o ../bin/agentglobe ./cmd/agentglobe

run: build
	cd agentglobe && CONFIG_PATH=$(LOCAL_CONFIG) ../bin/agentglobe

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
