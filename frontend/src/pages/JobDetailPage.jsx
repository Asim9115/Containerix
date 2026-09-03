import { useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { fetchJob, clearCurrentJob } from '../store/slices/jobsSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { Card } from '../components/ui/Card'
import { StatusBadge } from '../components/ui/Badge'
import { Button } from '../components/ui/Button'
import { PageLoader } from '../components/ui/Loading'
import { LogViewer } from '../components/logs/LogViewer'
import { formatDate } from '../utils/format'

export function JobDetailPage() {
  const { id } = useParams()
  const dispatch = useAppDispatch()
  const job = useAppSelector((s) => s.jobs.current)
  const loading = useAppSelector((s) => s.jobs.loading)

  useEffect(() => {
    dispatch(fetchJob(id))
    const interval = setInterval(() => dispatch(fetchJob(id)), 3000)
    return () => {
      clearInterval(interval)
      dispatch(clearCurrentJob())
    }
  }, [dispatch, id])

  if (loading && !job) return <PageLoader />

  const status = job?.status || job?.Status
  const step = job?.step || job?.Step
  const error = job?.error || job?.Error
  const deploymentId = job?.deployment_id || job?.DeploymentID

  return (
    <div>
      <PageHeader
        title={id}
        description={step || 'Deployment in progress'}
        breadcrumbs={[
          { label: 'Deploys', to: '/jobs' },
          { label: id },
        ]}
        actions={
          deploymentId && (
            <Link to={`/services/${deploymentId}`}>
              <Button variant="secondary" size="sm">View service</Button>
            </Link>
          )
        }
      />

      <PageContent className="space-y-6">
        <div className="grid grid-cols-2 sm:grid-cols-3 gap-px bg-border border border-border">
          <InfoCell label="Status" value={<StatusBadge status={status} />} />
          <InfoCell label="Step" value={step || '—'} />
          <InfoCell
            label="Service"
            value={
              deploymentId ? (
                <Link
                  to={`/services/${deploymentId}`}
                  className="font-mono text-fg-secondary hover:text-fg hover:underline underline-offset-2"
                >
                  {deploymentId}
                </Link>
              ) : (
                '—'
              )
            }
          />
        </div>

        <Card title="Timeline" noPadding>
          <dl className="divide-y divide-border">
            <DetailRow label="Started" value={formatDate(job?.created_at || job?.CreatedAt)} />
            <DetailRow label="Finished" value={formatDate(job?.completed_at || job?.CompletedAt)} />
            {error && <DetailRow label="Error" value={error} error />}
          </dl>
        </Card>

        <div>
          <h2 className="text-sm font-medium text-fg mb-3">Build logs</h2>
          <LogViewer jobId={id} />
        </div>
      </PageContent>
    </div>
  )
}

function InfoCell({ label, value }) {
  return (
    <div className="bg-surface-raised px-4 py-3">
      <p className="text-[11px] text-muted uppercase tracking-wide mb-1">{label}</p>
      <div className="text-sm text-fg">{value}</div>
    </div>
  )
}

function DetailRow({ label, value, error }) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-start gap-1 sm:gap-4 px-4 py-3">
      <dt className="text-xs text-muted sm:w-28 shrink-0">{label}</dt>
      <dd className={`text-sm flex-1 ${error ? 'text-red-400' : 'text-fg'}`}>{value}</dd>
    </div>
  )
}
