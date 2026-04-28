# Run from clawgotcha/ (or pass -C clawgotcha).
.PHONY: build test docker-build docker-push

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ./bin/clawgotcha ./cmd/clawgotcha

test:
	go test ./...

docker-build:
	docker build -f Dockerfile -t clawgotcha:local .

docker-push:
	@echo "Set IMAGE=ghcr.io/OWNER/clawgotcha:TAG and run: docker push \$$IMAGE"
