/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Backend origin in production, e.g. https://quota-backend.onrender.com.
   *  Empty in dev so requests hit "/api" via the Vite proxy. */
  readonly VITE_API_BASE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
