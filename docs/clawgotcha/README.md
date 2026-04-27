# clawgotcha `internal` package

## Overview

**clawgotcha** is the control-plane API for agent-swarm **configuration**—not the runtime that executes model calls. It persists metadata that lines up with agentic_swarm / claw-style TOML concepts: a single set of **defaults** (provider and model), many **agents** (Hands: prompts, tools, model overrides, autonomy), and **cron jobs** that target an agent by name. Clients and tools can read and edit this data over HTTP while a separate scheduler or host runs the actual work.

**Stack:** PostgreSQL, [GORM](https://gorm.io), [chi](https://github.com/go-chi/chi) for routing, standard library `net/http`. The server binary lives at [cmd/clawgotcha](../../clawgotcha/cmd/clawgotcha/main.go); on startup it loads [config](../../clawgotcha/internal/config/config.go), opens the DB via [db.Open](../../clawgotcha/internal/db/open.go), and serves [api.NewRouter](../../clawgotcha/internal/api/server.go). Embedded [OpenAPI](../../clawgotcha/internal/api/openapi.json) is available at `GET /openapi.json`.

**Data flow (typical request):** HTTP handler in `internal/api` → GORM on `*gorm.DB` → `swarm_*` tables. Errors are normalized to JSON with [httperr](../../clawgotcha/internal/httperr/httperr.go) so callers always get a structured `error` object when something fails.

**Package map:** `config` = how to listen and connect; `db` = schema and migrations; `api` = REST surface; `httperr` = response shaping. The subsections below describe each area in more detail.

## `internal/config`

Loads process settings with merge order: **defaults → optional YAML file → environment variables** (highest).

- **YAML** (optional, path from `-config`/`-c`, `CONFIG_PATH`, or [DefaultConfigPath](../../clawgotcha/internal/config/config.go)) can set `database_url`, `port`, `hostname`, `public_url`—aligned with shared deploy files such as `dep/cl.yaml`.
- **Environment** can override: `DATABASE_URL`, `HTTP_ADDR`, `HOSTNAME`, `PORT`, `PUBLIC_URL`.
- **Listen address** `HTTPAddr` is not read from YAML; it comes from `HTTP_ADDR` or is derived as `":" + Port` (default port **3477**), or `:8080` if port is invalid/zero.

## `internal/db`

PostgreSQL via GORM. [Open](../../clawgotcha/internal/db/open.go) runs **AutoMigrate** for swarm tables, then ensures a default row in `swarm_config` (id `1`).

| Model | Table | Role |
| --- | --- | --- |
| `SwarmConfig` | `swarm_config` | Single row (`id=1`): `default_provider`, `default_model` (mirrors agentic_swarm top-level defaults). |
| `SwarmAgent` | `swarm_agents` | One row per **Hand** / `[[agents]]` block: `name` (unique), `system_prompt`, `tools` (JSON), `provider`, `model`, `timeout_seconds`, `autonomy_level`. |
| `SwarmCronJob` | `swarm_cron_jobs` | One row per **cron** / `[[cron_jobs]]` block: `name` (unique), `agent_name`, `schedule`, `timeout_seconds`, `prompt`. |

Autonomy levels in code: `ReadOnly`, `Supervised`, `Full` (see [models.go](../../clawgotcha/internal/db/models.go)).

## `internal/httperr`

Small helper for **JSON error responses**: `PublicError` with `code` and `detail`, mapping `not_found` to HTTP 404, [gorm.ErrRecordNotFound](../../clawgotcha/internal/httperr/httperr.go) to 404, and everything else to 500 with `internal_error` and structured logging of the real error.

## `internal/api`

REST API on [chi](https://github.com/go-chi/chi), exposed by [NewRouter](../../clawgotcha/internal/api/server.go).

- **`GET /healthz`** — liveness JSON (`status`, `ts`).
- **`GET /openapi.json`** — embedded OpenAPI spec (see `openapi.json` + `openapi_embed.go`).
- **`/api/v1/config`** — `GET` / `PUT` for `SwarmConfig` (id 1).
- **`/api/v1/agents`** — CRUD for `SwarmAgent` (`GET/POST` list and create, `GET/PUT/PATCH/DELETE` by UUID, `GET /agents/by-name/{name}`).
- **`/api/v1/cron-jobs`** — CRUD for `SwarmCronJob`; create/put require a matching agent **name** (`agentExists` in [validate.go](../../clawgotcha/internal/api/validate.go)).

Support code:

- **[validate.go](../../clawgotcha/internal/api/validate.go)** — `autonomy_level` must be one of the three constants; cron jobs must reference an existing agent name.
- **[pqerr.go](../../clawgotcha/internal/api/pqerr.go)** — detect Postgres unique violation (`23505`) for friendlier `duplicate` errors.
- **[jsonutil.go](../../clawgotcha/internal/api/jsonutil.go)** — `writeJSON` helper.
- **Tests** — `server_test.go` (handler/router tests as implemented).
