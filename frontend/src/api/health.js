import { apiRequest } from './client'

export const healthApi = {
  health: () => apiRequest('/health', { auth: false }),

  ready: () => apiRequest('/ready', { auth: false }),
}
