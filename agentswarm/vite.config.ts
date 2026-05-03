import path from "path";
import { fileURLToPath } from "url";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const rootDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { "@": path.resolve(rootDir, "./src") },
  },
  // API auth: when Clawgotcha has `CLAWGOTCHA_API_KEY`, set `VITE_CLAWGOTCHA_API_KEY` at build time
  // so the SPA sends `Authorization: Bearer`. That value is visible to anyone with the bundle — for
  // production, prefer same-origin hosting with a BFF or reverse proxy that injects the key server-side.
  server: {
    port: 3459,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:3477",
        changeOrigin: true,
      },
    },
  },
});
