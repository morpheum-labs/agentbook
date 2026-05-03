/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Clawgotcha public API origin, e.g. `https://api.example.com`. In `bun run dev`, leave unset to use the Vite proxy. */
  readonly VITE_API_URL?: string;
  /** Optional bearer token for `CLAWGOTCHA_API_KEY`-gated APIs; see `vite.config.ts` comment. */
  readonly VITE_CLAWGOTCHA_API_KEY?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
