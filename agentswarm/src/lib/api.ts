const DEFAULT_API_ORIGIN = "http://127.0.0.1:3477";

/**
 * In dev, prefer same-origin + Vite proxy so the browser can call the API without CORS.
 * Set `VITE_API_URL` for preview/production when the static app and API are on different hosts.
 */
export function apiOrigin(): string {
  if (import.meta.env.DEV) {
    const raw = (import.meta.env.VITE_API_URL as string | undefined)?.trim();
    if (raw) return raw.replace(/\/$/, "");
    return "";
  }
  const raw = (import.meta.env.VITE_API_URL as string | undefined)?.trim();
  if (raw) return raw.replace(/\/$/, "");
  return DEFAULT_API_ORIGIN;
}

export function apiUrl(path: string): string {
  const p = path.startsWith("/") ? path : `/${path}`;
  const base = apiOrigin();
  return base ? `${base}${p}` : p;
}

export type SwarmAgent = {
  ID: string;
  Name: string;
  SystemPrompt: string;
  /** MiroClaw IDENTITY / SOUL / USER (also combined with markers in `SystemPrompt` in the DB). */
  identity?: string;
  soul?: string;
  user_context?: string;
  modular_prompt?: boolean;
  Tools: string[];
  Provider: string;
  Model: string;
  TimeoutSeconds: number;
  AutonomyLevel: string;
  CreatedAt: string;
  UpdatedAt: string;
};

export type AgentListResponse = { agents: SwarmAgent[] };

/** Single-agent GET/POST/PUT responses wrap the hand in `agent` plus `revision_summary`. */
type AgentSingleResponse = { agent: SwarmAgent };

export type PutAgentRequest = {
  name: string;
  identity: string;
  soul: string;
  user_context: string;
  tools: string[];
  provider: string;
  model: string;
  timeout_seconds: number;
  autonomy_level: string;
};

/** Body for `POST /api/v1/agents` (CreateOrReplaceAgentRequest). */
export type CreateAgentRequest = {
  name: string;
  /** If any of `identity` / `soul` / `user_context` is present, the server assembles `system_prompt`. */
  identity?: string;
  soul?: string;
  user_context?: string;
  system_prompt?: string;
  tools?: string[];
  provider?: string;
  model?: string;
  timeout_seconds?: number;
  autonomy_level: string;
};

export type ApiError = { error: { code: string; detail?: string } };

function parseErrorBody(text: string): string {
  try {
    const j = JSON.parse(text) as ApiError;
    if (j.error) {
      const d = j.error.detail?.trim();
      return d || j.error.code;
    }
  } catch {
    // ignore
  }
  return text || "Request failed";
}

/**
 * Central fetch for Clawgotcha JSON API. When `VITE_CLAWGOTCHA_API_KEY` is set at build time,
 * sends `Authorization: Bearer …` (unless headers already include auth). The value ships in the
 * static bundle — for production, prefer a BFF or reverse proxy that adds server-side auth.
 */
function apiFetch(path: string, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers);
  if (!headers.has("Accept")) {
    headers.set("Accept", "application/json");
  }
  const k = (import.meta.env.VITE_CLAWGOTCHA_API_KEY as string | undefined)?.trim();
  if (k && !headers.has("Authorization") && !headers.has("X-API-Key")) {
    headers.set("Authorization", `Bearer ${k}`);
  }
  return fetch(apiUrl(path), { ...init, headers });
}

export async function fetchAgents(): Promise<SwarmAgent[]> {
  const r = await apiFetch("/api/v1/agents", {
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const data = (await r.json()) as AgentListResponse;
  return data.agents ?? [];
}

export async function postAgent(body: CreateAgentRequest): Promise<SwarmAgent> {
  const r = await apiFetch("/api/v1/agents", {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const data = (await r.json()) as AgentSingleResponse;
  return data.agent;
}

export async function fetchAgent(id: string): Promise<SwarmAgent> {
  const r = await apiFetch(`/api/v1/agents/${encodeURIComponent(id)}`, {
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const data = (await r.json()) as AgentSingleResponse;
  return data.agent;
}

export async function putAgent(
  id: string,
  body: PutAgentRequest
): Promise<SwarmAgent> {
  const r = await apiFetch(`/api/v1/agents/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const data = (await r.json()) as AgentSingleResponse;
  return data.agent;
}

export type SwarmCronJob = {
  ID: string;
  Name: string;
  AgentName: string;
  Schedule: string;
  TimeoutSeconds: number;
  Prompt: string;
  Active: boolean;
  CreatedAt: string;
  UpdatedAt: string;
};

export type CronJobListResponse = { cron_jobs: SwarmCronJob[] };

export type CreateOrReplaceCronJobRequest = {
  name: string;
  agent_name: string;
  schedule?: string;
  timeout_seconds?: number;
  prompt?: string;
  active?: boolean;
};

export type PatchCronJobRequest = {
  name?: string;
  agent_name?: string;
  schedule?: string;
  timeout_seconds?: number;
  prompt?: string;
  active?: boolean;
};

export type CronScheduleTimelineRow = {
  ID: string;
  Name: string;
  AgentName: string;
  Schedule: string;
  Active: boolean;
  UpdatedAt: string;
  CreatedAt: string;
  anchor_at: string;
  ScheduleParsed: boolean;
  ParseError?: string;
  ProjectedRuns: string[];
};

export type CronScheduleTimelineResponse = {
  as_of: string;
  horizon_ends: string;
  anchored_by: string;
  horizon_hours: number;
  max_runs: number;
  rows: CronScheduleTimelineRow[];
};

export async function fetchCronScheduleTimeline(
  options?: { horizonHours?: number; maxRuns?: number }
): Promise<CronScheduleTimelineResponse> {
  const p = new URLSearchParams();
  if (options?.horizonHours != null) {
    p.set("horizon_hours", String(options.horizonHours));
  }
  if (options?.maxRuns != null) {
    p.set("max_runs", String(options.maxRuns));
  }
  const q = p.toString();
  const r = await apiFetch(`/api/v1/cron-jobs/schedule-timeline${q ? `?${q}` : ""}`, {
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  return (await r.json()) as CronScheduleTimelineResponse;
}

export async function fetchCronJobs(): Promise<SwarmCronJob[]> {
  const r = await apiFetch("/api/v1/cron-jobs", {
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const data = (await r.json()) as CronJobListResponse;
  return data.cron_jobs ?? [];
}

export async function postCronJob(
  body: CreateOrReplaceCronJobRequest
): Promise<SwarmCronJob> {
  const r = await apiFetch("/api/v1/cron-jobs", {
    method: "POST",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  return (await r.json()) as SwarmCronJob;
}

export async function fetchCronJob(id: string): Promise<SwarmCronJob> {
  const r = await apiFetch(`/api/v1/cron-jobs/${encodeURIComponent(id)}`, {
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  return (await r.json()) as SwarmCronJob;
}

export async function putCronJob(
  id: string,
  body: CreateOrReplaceCronJobRequest
): Promise<SwarmCronJob> {
  const r = await apiFetch(`/api/v1/cron-jobs/${encodeURIComponent(id)}`, {
    method: "PUT",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  return (await r.json()) as SwarmCronJob;
}

export async function patchCronJob(
  id: string,
  body: PatchCronJobRequest
): Promise<SwarmCronJob> {
  const r = await apiFetch(`/api/v1/cron-jobs/${encodeURIComponent(id)}`, {
    method: "PATCH",
    headers: {
      Accept: "application/json",
      "Content-Type": "application/json",
    },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  return (await r.json()) as SwarmCronJob;
}

export async function deleteCronJob(id: string): Promise<void> {
  const r = await apiFetch(`/api/v1/cron-jobs/${encodeURIComponent(id)}`, {
    method: "DELETE",
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
}

export type SwarmRuntimeInstance = {
  ID: string;
  InstanceName: string;
  InstanceType: string;
  Version: string;
  Hostname: string;
  PublicURL?: string | null;
  CallbackURL: string;
  Capabilities: string[];
  LastHeartbeatAt?: string | null;
  Status: string;
  StartedAt: string;
  Metadata?: Record<string, unknown> | null;
  CreatedAt: string;
  UpdatedAt: string;
};

export type RuntimeInstanceListResponse = {
  instances: SwarmRuntimeInstance[];
  revision_summary?: unknown;
};

export async function fetchInstances(options?: {
  status?: string;
}): Promise<SwarmRuntimeInstance[]> {
  const p = new URLSearchParams();
  if (options?.status?.trim()) {
    p.set("status", options.status.trim());
  }
  const q = p.toString();
  const path = `/api/v1/instances${q ? `?${q}` : ""}`;
  const r = await apiFetch(path, { headers: { Accept: "application/json" } });
  if (!r.ok) {
    const t = await r.text();
    let msg = parseErrorBody(t);
    if (r.status === 404) {
      msg +=
        " If this URL should exist, the control plane may be an older build without GET /api/v1/instances.";
    }
    throw new Error(msg);
  }
  const data = (await r.json()) as RuntimeInstanceListResponse;
  const list = data.instances;
  if (!Array.isArray(list)) {
    throw new Error(
      "Invalid instances response: expected JSON object with an `instances` array."
    );
  }
  return list;
}

/** Allowlisted `material_kind` values (must match Clawgotcha server). */
export const CREDENTIAL_MATERIAL_KINDS = [
  "api_key",
  "bearer_token",
  "github_pat",
  "oauth_client",
  "oauth_tokens",
  "oauth_authorization_pending",
  "totp_seed",
  "recovery_code_hashes",
] as const;

export type CredentialMaterialKind = (typeof CREDENTIAL_MATERIAL_KINDS)[number];

export type AgentCredentialBinding = {
  id: string;
  provider_slug: string;
  label: string;
  mcp_server_name?: string | null;
  metadata?: Record<string, unknown>;
  current_version: number;
  material_kind: CredentialMaterialKind | string | null;
  has_secret: boolean;
  expires_at?: string | null;
  secret_updated_at?: string | null;
  created_at: string;
  updated_at: string;
};

export type AgentCredentialListResponse = {
  credentials: AgentCredentialBinding[];
};

export type PostAgentCredentialRequest = {
  provider_slug: string;
  label: string;
  mcp_server_name?: string;
  metadata?: Record<string, unknown>;
  material_kind: CredentialMaterialKind | string;
  /** String (single secret) or structured object (e.g. OAuth). */
  plaintext: string | Record<string, unknown>;
};

export async function fetchAgentCredentials(agentId: string): Promise<AgentCredentialBinding[]> {
  const r = await apiFetch(`/api/v1/agents/${encodeURIComponent(agentId)}/credentials`, {
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const data = (await r.json()) as AgentCredentialListResponse;
  return data.credentials ?? [];
}

export async function postAgentCredential(
  agentId: string,
  body: PostAgentCredentialRequest
): Promise<AgentCredentialBinding> {
  const r = await apiFetch(`/api/v1/agents/${encodeURIComponent(agentId)}/credentials`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const wrap = (await r.json()) as { credential: AgentCredentialBinding };
  return wrap.credential;
}

export async function deleteAgentCredential(agentId: string, bindingId: string): Promise<void> {
  const r = await apiFetch(
    `/api/v1/agents/${encodeURIComponent(agentId)}/credentials/${encodeURIComponent(bindingId)}`,
    { method: "DELETE", headers: { Accept: "application/json" } }
  );
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
}

export async function rotateAgentCredential(
  agentId: string,
  bindingId: string,
  plaintext: string | Record<string, unknown>
): Promise<AgentCredentialBinding> {
  const r = await apiFetch(
    `/api/v1/agents/${encodeURIComponent(agentId)}/credentials/${encodeURIComponent(bindingId)}/rotate`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ plaintext }),
    }
  );
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  const wrap = (await r.json()) as { credential: AgentCredentialBinding };
  return wrap.credential;
}
