# clawgotcha `internal` package

## Overview

**clawgotcha** is the control-plane API for agent-swarm **configuration**—not the runtime that executes model calls. It persists metadata that lines up with agentic_swarm / claw-style TOML concepts: a single set of **defaults** (provider and model), many **agents** (Hands: prompts, tools, model overrides, autonomy), and **cron jobs** that target an agent by name. Clients and tools can read and edit this data over HTTP while a separate scheduler or host runs the actual work.

**Stack:** PostgreSQL, [GORM](https://gorm.io), [chi](https://github.com/go-chi/chi) for routing, standard library `net/http`. The server binary lives at [cmd/clawgotcha](../../clawgotcha/cmd/clawgotcha/main.go); on startup it loads [config](../../clawgotcha/internal/config/config.go), opens the DB via [db.Open](../../clawgotcha/internal/db/open.go), and serves [api.NewRouter](../../clawgotcha/internal/api/server.go). Embedded [OpenAPI](../../clawgotcha/internal/api/openapi.json) is available at `GET /openapi.json`.

**Data flow (typical request):** HTTP handler in `internal/api` → GORM on `*gorm.DB` → `swarm_*` tables. Errors are normalized to JSON with [httperr](../../clawgotcha/internal/httperr/httperr.go) so callers always get a structured `error` object when something fails.

**Package map:** `config` = how to listen and connect; `db` = schema and migrations; `api` = REST surface; `httperr` = response shaping. The subsections below describe each area in more detail.

## `internal/config`

Loads process settings with merge order: **defaults → optional YAML file → environment variables** (highest).

- **YAML** (optional, path from `-config`/`-c`, `CONFIG_PATH`, or [DefaultConfigPath](../../clawgotcha/internal/config/config.go)) can set `database_url`, `port`, `hostname`, `public_url`, and optional `credentials_encryption_key` for the agent credential vault—aligned with shared deploy files such as `dep/cl.yaml`.
- **Environment** can override: `DATABASE_URL`, `HTTP_ADDR`, `HOSTNAME`, `PORT`, `PUBLIC_URL`, `CLAWGOTCHA_CREDENTIALS_ENCRYPTION_KEY` (overrides YAML when set).
- **Listen address** `HTTPAddr` is not read from YAML; it comes from `HTTP_ADDR` or is derived as `":" + Port` (default port **3477**), or `:8080` if port is invalid/zero.

## `internal/db`

PostgreSQL via GORM. [Open](../../clawgotcha/internal/db/open.go) runs **AutoMigrate** for swarm tables, then ensures a default row in `swarm_config` (id `1`).

| Model | Table | Role |
| --- | --- | --- |
| `SwarmConfig` | `swarm_config` | Single row (`id=1`): `default_provider`, `default_model` (mirrors agentic_swarm top-level defaults). |
| `SwarmAgent` | `swarm_agents` | One row per **Hand** / `[[agents]]` block: `name` (unique), `system_prompt`, `tools` (JSON), `provider`, `model`, `timeout_seconds`, `autonomy_level`. |
| `CredentialBinding` / `CredentialSecretVersion` | `credential_bindings`, `credential_secret_versions` | Per-agent encrypted credential vault (AES-GCM at rest; see YAML `credentials_encryption_key` or `CLAWGOTCHA_CREDENTIALS_ENCRYPTION_KEY` and OpenAPI **Agent credentials**). |
| `SwarmCronJob` | `swarm_cron_jobs` | One row per **cron** / `[[cron_jobs]]` block: `name` (unique), `agent_name`, `schedule`, `timeout_seconds`, `prompt`, `active`. |

Autonomy levels in code: `ReadOnly`, `Supervised`, `Full` (see [models.go](../../clawgotcha/internal/db/models.go)).

## `internal/httperr`

Small helper for **JSON error responses**: `PublicError` with `code` and `detail`, mapping `not_found` to HTTP 404, [gorm.ErrRecordNotFound](../../clawgotcha/internal/httperr/httperr.go) to 404, and everything else to 500 with `internal_error` and structured logging of the real error.

## `internal/api`

REST API on [chi](https://github.com/go-chi/chi), exposed by [NewRouter](../../clawgotcha/internal/api/server.go). **CORS** is applied globally (including `OPTIONS` preflight) so a static SPA on another origin can call the JSON API; see [server.go](../../clawgotcha/internal/api/server.go).

Authoritative request/response shapes, status codes, and query parameters live in the embedded [OpenAPI](../../clawgotcha/internal/api/openapi.json) (also at `GET /openapi.json`). Notable JSON conventions: many entity responses use **PascalCase** (Go `json` defaults); create/update bodies for agents and cron often use **snake_case** keys; errors return an `error` object with `code` and optional `detail` (see `ErrorResponse` in the spec).

- **`GET /healthz`** — liveness JSON (`status`, `ts`).
- **`GET /openapi.json`** — embedded OpenAPI spec (see `openapi.json` + `openapi_embed.go`).
- **`/api/v1/config`** — `GET` / `PUT` for `SwarmConfig` (id 1).
- **`/api/v1/agents`** — `GET` list, `POST` create; `GET/PUT/PATCH/DELETE` by path UUID; **`GET /api/v1/agents/by-name/{name}`** to resolve by unique hand name. Create/put can supply modular parts (`identity`, `soul`, `user_context`) that are assembled into `system_prompt` in the store, or a flat `system_prompt`; see the `SwarmAgent` and request schemas in OpenAPI.
- **`/api/v1/agents/{id}/credentials`** — list, create, soft-delete, and rotate encrypted per-agent credentials (see OpenAPI tag **Agent credentials**).
- **`/api/v1/cron-jobs`** — `GET` list, `POST` create; `GET/PUT/PATCH/DELETE` by path UUID. Create, replace, and patch (when `agent_name` changes) require an existing agent name (`agentExists` in [validate.go](../../clawgotcha/internal/api/validate.go)).
- **`GET /api/v1/cron-jobs/schedule-timeline`** — projected next run instants for each **active** job (standard five-field cron), with query params `horizon_hours` (default 168, max 8760) and `max_runs` (default 64, max 200). Response includes `as_of`, `horizon_ends`, and per-job `ProjectedRuns`; tick alignment uses each row’s `anchor_at` (`UpdatedAt`, or `CreatedAt` if `UpdatedAt` is zero). Inactive or unparsable schedules yield empty `ProjectedRuns` (see OpenAPI for `CronScheduleTimelineRow`).

Support code:

- **[validate.go](../../clawgotcha/internal/api/validate.go)** — `autonomy_level` must be one of the three constants; cron jobs must reference an existing agent name.
- **[pqerr.go](../../clawgotcha/internal/api/pqerr.go)** — detect Postgres unique violation (`23505`) for friendlier `duplicate` errors.
- **[jsonutil.go](../../clawgotcha/internal/api/jsonutil.go)** — `writeJSON` helper.
- **Tests** — `server_test.go` (handler/router tests as implemented).
