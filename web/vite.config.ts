import react from "@vitejs/plugin-react";
import {defineConfig} from "vite";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    allowedHosts: true,
    proxy: {
      '/api': {
        target: 'http://localhost:5234',
        changeOrigin: true,
      },
      '/ws': {
        target: 'http://localhost:5234',
        changeOrigin: true,
        ws: true
      }
    },
  },
});
