import { apiRequest } from './client'

export const authApi = {
  register: (data) => apiRequest('/users', { method: 'POST', auth: false, body: data }),

  getProfile: () => apiRequest('/users/me'),

  rotateApiKey: () => apiRequest('/users/api-key', { method: 'POST' }),
}
