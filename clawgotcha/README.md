# Clawgotcha

HTTP API for **Miroclaw** (and similar) swarm metadata: agents (“Hands”), cron jobs, swarm-wide defaults, runtime registration, and change notifications (SSE + signed webhooks). Backed by **PostgreSQL** via GORM.

## Quick start (Docker)

From the **repository root**:

```bash
docker compose -f clawgotcha/docker-compose.yml up --build
```

- API base URL: `http://localhost:3477`
- Health: `GET /healthz`
- OpenAPI document: `GET /openapi.json`

Set `DATABASE_URL` (or `database_url` in YAML) to a Postgres DSN. The compose file wires Postgres and Clawgotcha together.

### Build the image locally

```bash
cd clawgotcha
docker build -f Dockerfile -t clawgotcha:local .
```

## Configuration

Priority: **environment variables** override **YAML**, which overrides **defaults**.

| Variable | Purpose |
|----------|---------|
| `DATABASE_URL` | Postgres connection string (required for server) |
| `PORT` | Listen port (default `3477`) |
| `HTTP_ADDR` | Full bind address (e.g. `:3477`); overrides `PORT` when set |
| `PUBLIC_URL` | Public base URL of this API (optional) |
| `CLAWGOTCHA_API_KEY` | When set, all `/api/v1/*` routes require this key (`Authorization: Bearer …` or `X-API-Key: …`). `/healthz`, `/openapi.json`, and `/metrics` stay unauthenticated. |
| `CLAWGOTCHA_INTERNAL_TOKEN` | When set, `POST /api/v1/events/publish` requires `Bearer` or `X-Internal-Token`. |
| `CLAWGOTCHA_RATE_LIMIT_RPS` | Max sustained requests per second per client IP on `/api/v1/*` (default `0` = disabled). |
| `CLAWGOTCHA_MAX_REQUEST_BODY_BYTES` | Max JSON body size for `/api/v1/*` (default `1048576`). |
| `CLAWGOTCHA_CREDENTIALS_ENCRYPTION_KEY` | **32-byte** AES-256 key for the per-agent credential vault: raw 32 characters, **base64** (std encoding), or **64 hex** digits. When unset (in both env and YAML), `GET /api/v1/agents/{id}/credentials` still lists bindings; `POST` create and `POST …/rotate` return **503**. Invalid format prevents the server from starting. Overrides YAML `credentials_encryption_key` when set in the environment. |

YAML config (e.g. `-c /data/config.yaml`) supports `port`, `hostname`, `public_url`, `database_url`, and optional `credentials_encryption_key` (same formats as the env var) — see [`config.yaml`](config.yaml).

### Agent credentials (vault)

Bindings live under an agent UUID. Each binding has `provider_slug`, `label`, optional `mcp_server_name`, and `metadata` (JSON, non-secret). Encrypted material is stored in versioned rows; the API **never** returns `ciphertext`, nonces, or decrypted secrets.

**`material_kind`** (allowlisted on the server): `api_key`, `bearer_token`, `github_pat`, `oauth_client`, `oauth_tokens`, `oauth_authorization_pending`, `totp_seed`, `recovery_code_hashes`.

| Method | Path |
|--------|------|
| GET | `/api/v1/agents/{id}/credentials` |
| POST | `/api/v1/agents/{id}/credentials` |
| DELETE | `/api/v1/agents/{id}/credentials/{bindingId}` |
| POST | `/api/v1/agents/{id}/credentials/{bindingId}/rotate` |

## Integration with Miroclaw

1. **Register the runtime** so Clawgotcha knows how to reach it and can deliver webhooks:

   `POST /api/v1/instances/register` with JSON including `instance_name`, `callback_url`, `hostname`, `version`, and optional `public_url`, `capabilities`, `metadata`.

   On first registration, Clawgotcha creates a [`SwarmWebhookSubscription`](internal/db/models.go) row with a random **secret** (used only for HMAC; it is not the same as `CLAWGOTCHA_API_KEY`).

2. **Heartbeats** — `POST /api/v1/instances/{instance_name}/heartbeat` regularly (e.g. every 30s). Instances with no heartbeat for longer than **90 seconds** are marked **offline** and their webhook subscriptions are disabled until the next successful heartbeat.

3. **Poll or push for config changes**
   - **Poll**: `GET /api/v1/agents` and `GET /api/v1/cron-jobs` with `?since_revision=…` or `?updated_after=…` (RFC3339). Responses include `revision_summary` (`config_revision`, `agents_max_revision`, `cron_jobs_max_revision`).
   - **SSE**: `GET /api/v1/events` (Server-Sent Events) for real-time `change` events after an initial `revision_summary` event.
   - **Webhooks**: On agent/cron/config changes, Clawgotcha POSTs the **same JSON body** as SSE (`event_type`, `affected_entity_type`, `affected_ids`, `new_revision`, `ts`) to each runtime’s `callback_url` for enabled subscriptions.

### Webhook verification (HMAC)

The JSON body is signed with **HMAC-SHA256** using the subscription **secret** (created at registration time). The receiver **must** verify using the **exact raw request body bytes**.

- **Signature header**: `X-Clawgotcha-Signature: sha256=<hex>`  
  Compute `HMAC-SHA256(secret, rawBody)` and compare **constant-time** to the hex after `sha256=`.

The signature is **not** duplicated inside the JSON payload; clients should rely on the header (and TLS to the callback URL).

Optional headers for debugging:

- `X-Clawgotcha-Event-Type` — same as `event_type` in JSON.

## API overview

| Method | Path | Notes |
|--------|------|--------|
| GET | `/healthz` | Liveness; includes DB check when configured |
| GET | `/openapi.json` | OpenAPI 3.0 spec |
| GET | `/metrics` | Prometheus metrics (Go runtime + `clawgotcha_http_*`) |
| | `/api/v1/config` | Swarm defaults |
| | `/api/v1/agents`, `/api/v1/agents/{id}/credentials`, `/api/v1/cron-jobs` | CRUD + delta query params; nested credentials on agents |
| | `/api/v1/instances/…` | Register, heartbeat, list, detail, delete |
| GET | `/api/v1/events` | SSE stream |
| POST | `/api/v1/events/publish` | Internal publish (token-gated) |

Full detail: **OpenAPI** at `/openapi.json` once the server is running.

## Production notes

- Terminate TLS at a reverse proxy (e.g. nginx, Envoy) in front of Clawgotcha.
- Set `CLAWGOTCHA_API_KEY` and restrict network access to Postgres.
- Use the **callback URL** only over HTTPS for Miroclaw webhook receivers.

## Development

```bash
cd clawgotcha
export DATABASE_URL='postgresql://user:pass@localhost:5432/dbname?sslmode=disable'
go run ./cmd/clawgotcha -c config.yaml
```

See [`Makefile`](Makefile) for `make build`, `make test`, and `make docker-build`.

Prompt helpers: `clawgotcha prompt compose|decompose` — see `clawgotcha compose --help`.
