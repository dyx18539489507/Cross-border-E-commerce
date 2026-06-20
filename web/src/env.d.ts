/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly BASE_URL: string
  readonly VITE_API_BASE_URL?: string
  readonly VITE_NETEASE_API_URL?: string
  readonly VITE_DEMO_DEVICE_ID?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
