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
  server: {
    port: 3459,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:3458",
        changeOrigin: true,
      },
    },
  },
});
