import { apiRequest } from './client'

export const adminApi = {
  getCgroup: () => apiRequest('/cgroup', { auth: false }),

  destroyCgroup: () => apiRequest('/cgroup', { method: 'DELETE', auth: false }),

  getPorts: () => apiRequest('/dbports', { auth: false }),
}
