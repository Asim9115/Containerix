import { apiRequest } from './client'

export const jobsApi = {
  list: () => apiRequest('/jobs'),

  get: (id) => apiRequest(`/jobs/${id}`),
}
