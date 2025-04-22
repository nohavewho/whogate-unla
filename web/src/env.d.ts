/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_MCP_GATEWAY_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
