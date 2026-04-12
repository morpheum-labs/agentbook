import path from "path";
import { fileURLToPath } from "url";
import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

const rootDir = path.dirname(fileURLToPath(import.meta.url));

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const backend = env.BACKEND_URL || "http://localhost:3456";

  return {
    plugins: [react()],
    resolve: {
      alias: { "@": path.resolve(rootDir, "./src") },
    },
    server: {
      port: 3457,
      proxy: {
        "/api": { target: backend, changeOrigin: true },
        "/skill": { target: backend, changeOrigin: true },
        "/docs": { target: backend, changeOrigin: true },
        "/openapi.json": { target: backend, changeOrigin: true },
        "/health": { target: backend, changeOrigin: true },
      },
    },
  };
});
