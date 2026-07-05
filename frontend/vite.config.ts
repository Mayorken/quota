import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// Proxy /api to the Go backend. Used for both the dev server and `vite preview`.
const proxy = {
  "/api": "http://localhost:8080",
};

export default defineConfig({
  plugins: [react()],
  server: {
    port: Number(process.env.PORT) || 5173,
    proxy,
  },
  preview: {
    port: Number(process.env.PORT) || 4173,
    proxy,
  },
});
