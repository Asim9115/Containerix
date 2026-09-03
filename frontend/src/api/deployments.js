import { apiRequest } from './client'

export const deploymentsApi = {
  list: () => apiRequest('/deployments'),

  get: (id) => apiRequest(`/deployments/${id}`),

  delete: (id) => apiRequest(`/deployments/${id}`, { method: 'DELETE' }),
}
