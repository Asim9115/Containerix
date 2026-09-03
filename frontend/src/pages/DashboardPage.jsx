import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { fetchDeployments } from '../store/slices/deploymentsSlice'
import { fetchJobs } from '../store/slices/jobsSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { Button } from '../components/ui/Button'
import { PageLoader } from '../components/ui/Loading'
import { ServiceListHeader, ServiceRow } from '../components/services/ServiceRow'
import { StatusBadge } from '../components/ui/Badge'
import { formatRelativeTime } from '../utils/format'

export function DashboardPage() {
  const dispatch = useAppDispatch()
  const deployments = useAppSelector((s) => s.deployments.items) ?? []
  const jobs = useAppSelector((s) => s.jobs.items) ?? []
  const loading = useAppSelector((s) => s.deployments.loading)

  useEffect(() => {
    dispatch(fetchDeployments())
    dispatch(fetchJobs())
  }, [dispatch])

  const running = deployments.filter((d) => d.Status === 'running').length
  const recentDeployments = [...deployments]
    .sort((a, b) => new Date(b.CreatedAt) - new Date(a.CreatedAt))
    .slice(0, 8)

  const recentJobs = [...jobs]
    .sort((a, b) => new Date(b.CreatedAt) - new Date(a.CreatedAt))
    .slice(0, 5)

  if (loading && deployments.length === 0) {
    return <PageLoader />
  }

  return (
    <div>
      <PageHeader
        title="Overview"
        description={`${deployments.length} service${deployments.length !== 1 ? 's' : ''} · ${running} live`}
        actions={
          <Link to="/deploy">
            <Button size="sm">New Web Service</Button>
          </Link>
        }
      />

      <PageContent className="space-y-6">
        <div>
          <div className="flex items-center justify-between mb-3">
            <h2 className="text-sm font-medium text-fg">Services</h2>
            <Link to="/services" className="text-xs text-muted hover:text-fg-secondary transition-colors">
              View all
            </Link>
          </div>

          {recentDeployments.length === 0 ? (
            <div className="border border-dashed border-border py-12 text-center">
              <p className="text-sm text-muted mb-3">No services deployed yet</p>
              <Link to="/deploy">
                <Button size="sm">Create Web Service</Button>
              </Link>
            </div>
          ) : (
            <div className="border border-border">
              <ServiceListHeader />
              {recentDeployments.map((service) => (
                <ServiceRow key={service.ID} service={service} />
              ))}
            </div>
          )}
        </div>

        {recentJobs.length > 0 && (
          <div>
            <div className="flex items-center justify-between mb-3">
              <h2 className="text-sm font-medium text-fg">Recent deploys</h2>
              <Link to="/jobs" className="text-xs text-muted hover:text-fg-secondary transition-colors">
                View all
              </Link>
            </div>
            <div className="border border-border divide-y divide-border">
              {recentJobs.map((job) => (
                <Link
                  key={job.ID}
                  to={`/jobs/${job.ID}`}
                  className="flex items-center justify-between px-4 py-3 hover:bg-surface-hover transition-colors"
                >
                  <div>
                    <p className="text-sm font-mono text-fg">{job.ID}</p>
                    <p className="text-xs text-muted mt-0.5">{job.Step || '—'}</p>
                  </div>
                  <div className="flex items-center gap-4">
                    <span className="text-xs text-muted hidden sm:block">
                      {formatRelativeTime(job.CreatedAt)}
                    </span>
                    <StatusBadge status={job.Status} />
                  </div>
                </Link>
              ))}
            </div>
          </div>
        )}
      </PageContent>
    </div>
  )
}
