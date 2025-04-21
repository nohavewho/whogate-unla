import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import pkg from './package.json';

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(pkg.version),
  },
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
