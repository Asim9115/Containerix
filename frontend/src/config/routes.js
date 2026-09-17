export const routes = {
  login: '/login',
  dashboard: '/',
  services: '/services',
  serviceDetail: '/services/:id',
  deploy: '/deploy',
  jobs: '/jobs',
  jobDetail: '/jobs/:id',
  settings: '/settings',
  admin: '/admin',
  adminLogin: '/admin/login',
}

export const navItems = [
  { label: 'Overview', path: routes.dashboard, icon: 'dashboard' },
  { label: 'Services', path: routes.services, icon: 'services' },
  { label: 'Deploy', path: routes.deploy, icon: 'deploy' },
  { label: 'Deploys', path: routes.jobs, icon: 'jobs' },
  { label: 'Settings', path: routes.settings, icon: 'settings' },
]
