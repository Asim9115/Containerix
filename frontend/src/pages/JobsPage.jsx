import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { fetchJobs } from '../store/slices/jobsSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { StatusBadge } from '../components/ui/Badge'
import { Button } from '../components/ui/Button'
import { PageLoader, EmptyState } from '../components/ui/Loading'
import { formatDate } from '../utils/format'

export function JobsPage() {
  const dispatch = useAppDispatch()
  const items = useAppSelector((s) => s.jobs.items) ?? []
  const loading = useAppSelector((s) => s.jobs.loading)

  useEffect(() => {
    dispatch(fetchJobs())
    const interval = setInterval(() => dispatch(fetchJobs()), 5000)
    return () => clearInterval(interval)
  }, [dispatch])

  if (loading && items.length === 0) return <PageLoader />

  const sorted = [...items].sort((a, b) => new Date(b.CreatedAt) - new Date(a.CreatedAt))

  return (
    <div>
      <PageHeader
        title="Deploys"
        description="Build and deployment history"
        actions={
          <Link to="/deploy">
            <Button size="sm">New Web Service</Button>
          </Link>
        }
      />

      <PageContent>
        {sorted.length === 0 ? (
          <EmptyState
            title="No deploys yet"
            description="Deploy a service from a Git repository to see build history here."
            action={
              <Link to="/deploy">
                <Button size="sm">New Web Service</Button>
              </Link>
            }
          />
        ) : (
          <div className="border border-border">
            <div className="hidden sm:flex items-center px-4 py-2 border-b border-border text-[11px] font-medium text-muted uppercase tracking-wider">
              <div className="flex-1">Deploy ID</div>
              <div className="w-28">Status</div>
              <div className="w-48 hidden md:block">Step</div>
              <div className="w-36 hidden lg:block">Started</div>
              <div className="w-36 hidden lg:block">Finished</div>
            </div>
            {sorted.map((job) => (
              <Link
                key={job.ID}
                to={`/jobs/${job.ID}`}
                className="flex flex-col sm:flex-row sm:items-center gap-2 sm:gap-0 px-4 py-3 border-b border-border last:border-b-0 hover:bg-surface-hover transition-colors"
              >
                <div className="flex-1 font-mono text-sm text-fg">{job.ID}</div>
                <div className="w-28">
                  <StatusBadge status={job.Status} />
                </div>
                <div className="w-48 hidden md:block text-xs text-muted truncate">
                  {job.Step || '—'}
                </div>
                <div className="w-36 hidden lg:block text-xs text-muted">
                  {formatDate(job.CreatedAt)}
                </div>
                <div className="w-36 hidden lg:block text-xs text-muted">
                  {job.CompletedAt ? formatDate(job.CompletedAt) : '—'}
                </div>
              </Link>
            ))}
          </div>
        )}
      </PageContent>
    </div>
  )
}
