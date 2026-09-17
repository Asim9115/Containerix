import { apiRequest } from './client'

export const adminApi = {
  /** Verifies the admin key against a protected endpoint. */
  verify: (apiKey) =>
    apiRequest('/dbports', {
      admin: false,
      auth: false,
      headers: { 'X-API-Key': apiKey },
    }),

  getCgroup: () => apiRequest('/cgroup', { admin: true, auth: false }),

  destroyCgroup: () =>
    apiRequest('/cgroup', { method: 'DELETE', admin: true, auth: false }),

  getPorts: () => apiRequest('/dbports', { admin: true, auth: false }),
}
