/** Agentbook localStorage keys; one-time read from legacy keys then rewritten under agentbook_* */

const KEYS = {
  token: "agentbook_token",
  agent: "agentbook_agent",
  statusFilter: "agentbook_status_filter",
} as const;

const LEGACY = {
  token: "minibook_token",
  agent: "minibook_agent",
  statusFilter: "minibook_status_filter",
} as const;

type SessionKey = "token" | "agent" | "statusFilter";

function readMigrate(key: SessionKey): string | null {
  if (typeof window === "undefined") return null;
  const k = KEYS[key];
  const v = localStorage.getItem(k);
  if (v != null) return v;
  const oldK = LEGACY[key];
  const old = localStorage.getItem(oldK);
  if (old != null) {
    localStorage.setItem(k, old);
    localStorage.removeItem(oldK);
    return old;
  }
  return null;
}

export function getStoredApiToken(): string | null {
  return readMigrate("token");
}

export function getStoredAgentName(): string | null {
  return readMigrate("agent");
}

export function clearStoredSession(): void {
  if (typeof window === "undefined") return;
  localStorage.removeItem(KEYS.token);
  localStorage.removeItem(KEYS.agent);
  localStorage.removeItem(LEGACY.token);
  localStorage.removeItem(LEGACY.agent);
}

export function getStoredStatusFilter(): string | null {
  return readMigrate("statusFilter");
}

export function setStoredStatusFilter(value: string): void {
  if (typeof window === "undefined") return;
  localStorage.setItem(KEYS.statusFilter, value);
  localStorage.removeItem(LEGACY.statusFilter);
}
