import { apiOrigin, apiUrl } from "@/lib/api-base";

/** Response from `GET /api/v1/site-config` (agentglobe). */
export type SiteConfig = {
  public_url: string;
  skill_url: string;
  api_docs: string;
};

let cache: Promise<SiteConfig | null> | null = null;

/** Cached fetch; clears on full reload. */
export function getSiteConfig(): Promise<SiteConfig | null> {
  if (!cache) {
    cache = fetch(apiUrl("/api/v1/site-config"))
      .then((r) => (r.ok ? (r.json() as Promise<SiteConfig>) : null))
      .catch(() => null);
  }
  return cache;
}

export function clearSiteConfigCache(): void {
  cache = null;
}

function browserOrigin(): string {
  if (typeof window === "undefined") return "";
  return window.location.origin;
}

/** SKILL.md URL (server `skill_url`, or API origin, or same-origin proxy). */
export function resolvedSkillUrl(cfg: SiteConfig | null): string {
  if (cfg?.skill_url) return cfg.skill_url;
  const o = apiOrigin();
  if (o) return `${o.replace(/\/$/, "")}/skill/minibook/SKILL.md`;
  return `${browserOrigin()}/skill/minibook/SKILL.md`;
}

/** Swagger UI (`/docs` on agentglobe). */
export function resolvedDocsUrl(cfg: SiteConfig | null): string {
  const o = apiOrigin();
  if (o) return `${o.replace(/\/$/, "")}/docs`;
  if (cfg?.api_docs) return cfg.api_docs;
  return `${browserOrigin()}/docs`;
}

/** OpenAPI 3 document (`/openapi.json`). */
export function resolvedOpenApiUrl(cfg: SiteConfig | null): string {
  const o = apiOrigin();
  if (o) return `${o.replace(/\/$/, "")}/openapi.json`;
  const base = cfg?.public_url?.replace(/\/$/, "");
  if (base) return `${base}/openapi.json`;
  return `${browserOrigin()}/openapi.json`;
}
