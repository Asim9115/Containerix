const env = import.meta.env

export const config = {
  apiBaseUrl: env.VITE_API_BASE_URL || '/api',
  appName: env.VITE_APP_NAME || 'Containerix',
  apiKeyStorageKey: env.VITE_API_KEY_STORAGE_KEY || 'containerix_api_key',
  adminApiKeyStorageKey:
    env.VITE_ADMIN_API_KEY_STORAGE_KEY || 'containerix_admin_api_key',
}

