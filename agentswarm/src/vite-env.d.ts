/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Clawlaundry public API origin, e.g. `https://api.example.com`. In `bun run dev`, leave unset to use the Vite proxy. */
  readonly VITE_API_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
