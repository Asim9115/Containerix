import { apiRequest } from './client'

export const buildApi = {
  trigger: (data) => apiRequest('/build', { method: 'POST', body: data }),
}
