import { apiRequest } from './client'

export const containersApi = {
  list: () => apiRequest('/containers'),

  get: (id) => apiRequest(`/containers/${id}`),

  delete: (id) => apiRequest(`/containers/${id}`, { method: 'DELETE' }),

  stop: (id) => apiRequest(`/containers/${id}/stop`, { method: 'POST' }),

  stopAll: () => apiRequest('/containers/stop-all', { method: 'POST' }),
}
