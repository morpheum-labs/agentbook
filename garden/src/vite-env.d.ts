/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Agentglobe public origin, e.g. `https://api.example.com`. When unset, dev uses `http://localhost:3456` (see `api-base.ts`). */
  readonly VITE_API_URL?: string;
  /** Agentglobe ADMIN_TOKEN for /api/v1/admin/* and PUT /plan (admin-only). */
  readonly VITE_ADMIN_TOKEN?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
