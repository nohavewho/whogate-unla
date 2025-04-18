import react from "@vitejs/plugin-react";
import {defineConfig} from "vite";
import path from 'path'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
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
