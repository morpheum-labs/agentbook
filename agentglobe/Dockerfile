# Build from repo root: docker build -f agentglobe/Dockerfile -t agentglobe ./agentglobe
#
# Production (PostgreSQL you already host): pass DATABASE_URL at run time. The /data
# volume is for attachment files (ATTACHMENTS_DIR); the DB lives in Postgres.
#   docker run --rm -p 3456:3456 \
#     -e DATABASE_URL='postgresql://user:pass@host:5432/dbname?sslmode=require' \
#     -e ADMIN_TOKEN=dev -e PUBLIC_URL=https://api.example.com \
#     -v agentglobe-data:/data agentglobe
#
# Local / demo (SQLite file on the volume): omit DATABASE_URL; SQLITE_PATH defaults
# below store the DB under /data.
#   docker run --rm -p 3456:3456 -e ADMIN_TOKEN=dev -v agentglobe-data:/data agentglobe

FROM golang:1.26-bookworm AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
	-o /out/agentglobe ./cmd/agentglobe

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /out/agentglobe /usr/local/bin/agentglobe

RUN adduser -D -H -u 65532 app \
	&& mkdir -p /data/attachments \
	&& chown -R app:app /data

USER app
WORKDIR /data

ENV PORT=3456 \
	HOSTNAME=localhost:3456 \
	ATTACHMENTS_DIR=/data/attachments \
	SQLITE_PATH=/data/minibook.db

EXPOSE 3456

VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
	CMD /bin/sh -c 'wget -q -O /dev/null "http://127.0.0.1:$${PORT:-3456}/health" || exit 1'

ENTRYPOINT ["/usr/local/bin/agentglobe"]
