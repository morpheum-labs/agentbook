# Build from this module directory (context = clawgotcha/):
#   docker build -f Dockerfile -t clawgotcha:local .
#
# The image bakes in config.yaml at /data and starts with: clawgotcha -c /data/config.yaml
# database_url in YAML is empty by default: set -e DATABASE_URL=... (overrides) or fill YAML via bind mount.
# Optional env overrides: PORT, PUBLIC_URL, HTTP_ADDR, HOSTNAME (see internal/config).
#
# Example (Postgres via env, default listen port 3477):
#   docker run --rm -p 3477:3477 \
#     -e DATABASE_URL='postgresql://user:pass@host:5432/db?sslmode=require' \
#     -e PUBLIC_URL=https://api.example.com \
#     -v clawgotcha-data:/data clawgotcha:local
#
# Replace the baked file (optional):
#   -v $(pwd)/my.yaml:/data/config.yaml:ro

FROM golang:1.26-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
	-o /out/clawgotcha ./cmd/clawgotcha

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/clawgotcha /usr/local/bin/clawgotcha

RUN adduser -D -H -u 65532 app && mkdir -p /data

COPY config.yaml /data/config.yaml
RUN chown -R app:app /data

USER app
WORKDIR /data

ENV PORT=3477 \
	HOSTNAME=localhost:3477

EXPOSE 3477

VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD /bin/sh -c 'wget -q -O /dev/null "http://127.0.0.1:$${PORT:-3477}/healthz" || exit 1'

ENTRYPOINT ["/usr/local/bin/clawgotcha", "-c", "/data/config.yaml"]
