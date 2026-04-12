/**
 * Resolve API URLs for agentglobe. The UI always talks to the API origin directly
 * (no reverse proxy or Vite dev proxy). Set `VITE_API_URL` for non-default hosts.
 */

const DEFAULT_API_ORIGIN = "http://localhost:3456";

export function apiOrigin(): string {
  const raw = (import.meta.env.VITE_API_URL as string | undefined)?.trim();
  if (raw) return raw.replace(/\/$/, "");
  return DEFAULT_API_ORIGIN;
}

export function apiUrl(path: string): string {
  const p = path.startsWith("/") ? path : `/${path}`;
  return `${apiOrigin()}${p}`;
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
