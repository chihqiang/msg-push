/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_PROXY_TARGET?: string
  readonly VITE_AUTH_DOMAIN?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
