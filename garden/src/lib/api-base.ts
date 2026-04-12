/**
 * Resolve API URLs for the Agentbook HTTP server (agentglobe).
 * Dev: leave VITE_API_URL unset so requests stay same-origin; Vite proxies /api, /skill, /docs,
 * /openapi.json, and /health to agentglobe (see vite.config.ts).
 * Prod: set VITE_API_URL to the public origin that reverse-proxies to agentglobe.
 */

export function apiOrigin(): string {
  const raw = import.meta.env.VITE_API_URL as string | undefined;
  if (!raw?.trim()) return "";
  return raw.replace(/\/$/, "");
}

export function apiUrl(path: string): string {
  const p = path.startsWith("/") ? path : `/${path}`;
  const o = apiOrigin();
  return o ? `${o}${p}` : p;
}

/** Bearer ADMIN_TOKEN for agentglobe admin routes. */
export function adminAuthHeaders(jsonBody = false): HeadersInit {
  const h: Record<string, string> = {};
  if (jsonBody) {
    h["Content-Type"] = "application/json";
  }
  const t = (import.meta.env.VITE_ADMIN_TOKEN as string | undefined)?.trim();
  if (t) {
    h.Authorization = `Bearer ${t}`;
  }
  return h;
}
