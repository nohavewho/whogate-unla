import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': '/src',
    },
  },
  server: {
    allowedHosts: true,
    proxy: {
      '/api': {
        target: 'http://localhost:5234',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:5234',
        ws: true,
      },
    },
  },
});
