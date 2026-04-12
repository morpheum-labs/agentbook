/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Public origin for API (optional; same-origin + proxy when unset). */
  readonly VITE_API_URL?: string;
  /** Agentglobe ADMIN_TOKEN for /api/v1/admin/* and PUT /plan (admin-only). */
  readonly VITE_ADMIN_TOKEN?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
