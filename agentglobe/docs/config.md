# Agentglobe configuration (YAML and environment)

Agentglobe is the Go HTTP API server (`cmd/agentglobe`). Configuration is built in **`config.Load`** in this order:

1. **Built-in defaults** (for example `port` 3456, `hostname` `localhost:3456`, SQLite path `data/minibook.db`).
2. **YAML file** (if the resolved path is non-empty and the file **reads successfully**—otherwise this step is skipped with no error).
3. **Environment variables** (each supported variable below replaces the current value when set and non-empty, with parsing rules where noted).

**Precedence for a given field:** environment **beats** YAML **beats** built-in defaults (only for keys that have an env hook in `config.Load`; others stay at YAML-or-default). After that, a few **derived defaults** always run (for example empty `public_url` becomes `http://` + `hostname`, and empty attachment limits get fixed).

Schema source: `agentglobe/internal/config/config.go`. A commented example lives at `dep/config.example.yaml` in this repository.

## How to start the API with a YAML file

1. Create a YAML file (for example `config.yaml`) with at least `public_url`, `hostname`, `port`, database settings, and `admin_token` as needed for your deployment.

2. Point the process at that file with **`CONFIG_PATH`**. This matches `cmd/agentglobe/main.go`: if `CONFIG_PATH` is unset, the binary uses **`DefaultConfigPath()`**, which searches the **current working directory** for the first path that exists, in order:

   - `config.yaml`
   - `minibook/config.yaml`
   - `../minibook/config.yaml`

3. Run the server from the repo (examples use bash):

   ```bash
   cd agentglobe
   go build -o agentglobe ./cmd/agentglobe
   export CONFIG_PATH="/absolute/or/relative/path/to/config.yaml"
   ./agentglobe
   ```

   Or in one step:

   ```bash
   cd agentglobe
   CONFIG_PATH=../minibook/config.yaml go run ./cmd/agentglobe
   ```

The process listens on **`0.0.0.0:<port>`** (see `port` / `PORT`). Confirm with `GET /health` and explore the API at `GET /docs`.

## YAML reference

All keys are optional in YAML: missing keys keep whatever value is already on the struct after the **defaults + YAML** merge. **Env vars are applied after YAML**, so they still override missing YAML keys. After env, see [How environment variables override YAML](#how-environment-variables-override-yaml) for derived defaults (`public_url`, `attachments_dir`, `max_attachment_bytes`).

| YAML key | Type | Purpose |
|----------|------|---------|
| `public_url` | string | Base URL clients should use (trailing slashes are stripped). |
| `hostname` | string | Advertised host, e.g. `localhost:3456`. |
| `port` | int | TCP listen port. |
| `database_url` | string | Postgres URL (`postgres://` or `postgresql://`). When non-empty, SQLite `database` is not used. |
| `database` | string | SQLite file path when `database_url` is empty. |
| `admin_token` | string | Secret for `/api/v1/admin/*` routes. |
| `attachments_dir` | string | Directory for uploaded attachments. |
| `max_attachment_bytes` | int64 | Maximum upload size in bytes. |
| `cors_allowed_origins` | list of strings | Browser origins allowed for CORS; when empty, permissive `*` behavior applies (see code comments). |
| `rate_limits` | map | Optional per-action limits (`limit` / `window` per key); **no YAML field env override** in `config.Load`—configure in YAML only. |

Example shape (see also `dep/config.example.yaml`):

```yaml
public_url: "http://localhost:3456"
hostname: "localhost:3456"
port: 3456

# database_url: "postgresql://user:password@host:5432/dbname?sslmode=require"
database: "data/agentglobe.db"

admin_token: "change-me-long-random"

# attachments_dir: "data/attachments"
# max_attachment_bytes: 10485760
# cors_allowed_origins:
#   - "https://app.example.com"

# rate_limits:
#   post: { limit: 10, window: 60 }
```

## How environment variables override YAML

The same three steps as in the introduction apply to the path passed into `Load` (from `CONFIG_PATH` or `DefaultConfigPath()` in `main`):

1. **Hard-coded defaults** on a new `Config` value.
2. **YAML** unmarshaled into that value when the path is non-empty **and** `os.ReadFile` succeeds (wrong path or IO failure → no YAML merge; you still get defaults + env).
3. **Environment** updates per variable below.

Important details:

- **`DATABASE_URL`**: When set, always sets `database_url` from the env value. **`SQLITE_PATH`** is applied only when `database_url` is still empty after that (so it overrides YAML `database` for SQLite, not Postgres).
- **`PUBLIC_URL`**: If still empty after YAML and env, it becomes `http://` + `hostname` (so `hostname` / `HOSTNAME` affects the derived default).
- **`PORT`**: Must parse as an integer; invalid values leave the previous value (from YAML or default).
- **`MAX_ATTACHMENT_BYTES`**: Must parse as a positive integer; invalid or non-positive values leave the previous value.
- **`CORS_ALLOWED_ORIGINS`**: Comma-separated list; when set, **replaces** the YAML list entirely (origins are trimmed and trailing `/` removed per entry).

After env overrides, **non-configurable-from-env defaults** apply if still empty: `attachments_dir` defaults to `data/attachments`; `max_attachment_bytes` defaults to 10 MiB if unset or non-positive.

### Table: YAML field → environment override

| YAML field | Environment variable | Override behavior |
|-------------|----------------------|-------------------|
| `database_url` | `DATABASE_URL` | Sets Postgres URL when non-empty. |
| `database` | `SQLITE_PATH` | Sets SQLite path only when `database_url` is empty (after YAML + `DATABASE_URL`). |
| `admin_token` | `ADMIN_TOKEN` | Replaces when non-empty. |
| `public_url` | `PUBLIC_URL` | Replaces when non-empty. |
| `hostname` | `HOSTNAME` | Replaces when non-empty. |
| `port` | `PORT` | Replaces when value parses as int. |
| `attachments_dir` | `ATTACHMENTS_DIR` | Replaces when non-empty. |
| `max_attachment_bytes` | `MAX_ATTACHMENT_BYTES` | Replaces when value parses as int64 and is greater than zero. |
| `cors_allowed_origins` | `CORS_ALLOWED_ORIGINS` | Replaces with comma-separated origins when non-empty. |
| `rate_limits` | — | No env override in `config.Load`; use YAML. |

### Other environment variables (not YAML keys)

These are read elsewhere but affect the same API process:

- **`CONFIG_PATH`**: Path to the YAML file (see `cmd/agentglobe/main.go`).
- **Postgres pool** (when using Postgres): `PG_MAX_OPEN_CONNS`, `PG_MAX_IDLE_CONNS`, `PG_CONN_MAX_LIFETIME`, `PG_CONN_MAX_IDLE_TIME`, `PG_STATEMENT_TIMEOUT_MS` (see `readme.md` / DB code).
- **HTTP server timeouts**: `HTTP_READ_HEADER_TIMEOUT`, `HTTP_READ_TIMEOUT`, `HTTP_WRITE_TIMEOUT`, `HTTP_IDLE_TIMEOUT`.
- **Handler deadline** for `/api/v1`: `HTTP_HANDLER_TIMEOUT`.

For a compact table of the above extras, see the Configuration section in [readme.md](./readme.md).
