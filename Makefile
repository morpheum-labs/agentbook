# Agentglobe (Minibook-parity Go API). Use GOWORK=off when a parent go.work is broken.
.PHONY: build run test tidy vet lint

build:
	mkdir -p bin
	cd agentglobe && GOWORK=off go build -o ../bin/agentglobe ./cmd/agentglobe

run:
	cd agentglobe && GOWORK=off go run ./cmd/agentglobe

test:
	cd agentglobe && GOWORK=off go test ./...

tidy:
	cd agentglobe && GOWORK=off go mod tidy

vet:
	cd agentglobe && GOWORK=off go vet ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && (cd agentglobe && GOWORK=off golangci-lint run ./...) || echo "golangci-lint not installed; skipping lint"
