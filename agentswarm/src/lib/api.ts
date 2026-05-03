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

export async function fetchAgents(): Promise<SwarmAgent[]> {
  const r = await fetch(apiUrl("/api/v1/agents"), {
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
  const r = await fetch(apiUrl("/api/v1/agents"), {
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
  const r = await fetch(apiUrl(`/api/v1/agents/${encodeURIComponent(id)}`), {
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
  const r = await fetch(apiUrl(`/api/v1/agents/${encodeURIComponent(id)}`), {
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
  const r = await fetch(
    apiUrl(`/api/v1/cron-jobs/schedule-timeline${q ? `?${q}` : ""}`),
    { headers: { Accept: "application/json" } }
  );
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  return (await r.json()) as CronScheduleTimelineResponse;
}

export async function fetchCronJobs(): Promise<SwarmCronJob[]> {
  const r = await fetch(apiUrl("/api/v1/cron-jobs"), {
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
  const r = await fetch(apiUrl("/api/v1/cron-jobs"), {
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
  const r = await fetch(apiUrl(`/api/v1/cron-jobs/${encodeURIComponent(id)}`), {
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
  const r = await fetch(apiUrl(`/api/v1/cron-jobs/${encodeURIComponent(id)}`), {
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
  const r = await fetch(apiUrl(`/api/v1/cron-jobs/${encodeURIComponent(id)}`), {
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
  const r = await fetch(apiUrl(`/api/v1/cron-jobs/${encodeURIComponent(id)}`), {
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
  const r = await fetch(apiUrl(path), { headers: { Accept: "application/json" } });
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
