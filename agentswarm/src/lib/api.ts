const DEFAULT_API_ORIGIN = "http://127.0.0.1:3456";

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
  Tools: string[];
  Provider: string;
  Model: string;
  TimeoutSeconds: number;
  AutonomyLevel: string;
  CreatedAt: string;
  UpdatedAt: string;
};

export type AgentListResponse = { agents: SwarmAgent[] };

export type PutAgentRequest = {
  name: string;
  system_prompt: string;
  tools: string[];
  provider: string;
  model: string;
  timeout_seconds: number;
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

export async function fetchAgent(id: string): Promise<SwarmAgent> {
  const r = await fetch(apiUrl(`/api/v1/agents/${encodeURIComponent(id)}`), {
    headers: { Accept: "application/json" },
  });
  if (!r.ok) {
    const t = await r.text();
    throw new Error(parseErrorBody(t));
  }
  return (await r.json()) as SwarmAgent;
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
  return (await r.json()) as SwarmAgent;
}
