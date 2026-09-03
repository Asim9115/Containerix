import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import {
  fetchDeployment,
  deleteDeployment,
  clearCurrentDeployment,
} from '../store/slices/deploymentsSlice'
import { stopContainer } from '../store/slices/containersSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { Card } from '../components/ui/Card'
import { StatusBadge } from '../components/ui/Badge'
import { Button } from '../components/ui/Button'
import { ConfirmModal } from '../components/ui/Modal'
import { PageLoader } from '../components/ui/Loading'
import { LogViewer } from '../components/logs/LogViewer'
import {
  getRepoName,
  formatDate,
  getServiceUrl,
  parseEnvJson,
} from '../utils/format'

export function ServiceDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const dispatch = useAppDispatch()
  const service = useAppSelector((s) => s.deployments.current)
  const loading = useAppSelector((s) => s.deployments.loading)
  const [showDelete, setShowDelete] = useState(false)
  const [actionLoading, setActionLoading] = useState(false)

  useEffect(() => {
    dispatch(fetchDeployment(id))
    const interval = setInterval(() => dispatch(fetchDeployment(id)), 5000)
    return () => {
      clearInterval(interval)
      dispatch(clearCurrentDeployment())
    }
  }, [dispatch, id])

  const handleStop = async () => {
    if (!service?.ContainerID) return
    setActionLoading(true)
    await dispatch(stopContainer(service.ContainerID))
    dispatch(fetchDeployment(id))
    setActionLoading(false)
  }

  const handleDelete = async () => {
    setActionLoading(true)
    await dispatch(deleteDeployment(id))
    setActionLoading(false)
    navigate('/services')
  }

  if (loading && !service) return <PageLoader />

  if (!service) {
    return (
      <PageContent>
        <p className="text-muted text-sm">Service not found.</p>
        <Link to="/services" className="text-sm text-fg-secondary hover:text-fg mt-2 inline-block">
          Back to services
        </Link>
      </PageContent>
    )
  }

  const env = parseEnvJson(service.EnvJSON)
  const name = getRepoName(service.RepoURL)

  return (
    <div>
      <PageHeader
        title={name}
        breadcrumbs={[
          { label: 'Services', to: '/services' },
          { label: name },
        ]}
        actions={
          <div className="flex items-center gap-2">
            {service.Status === 'running' && service.ContainerID && (
              <Button variant="secondary" size="sm" onClick={handleStop} loading={actionLoading}>
                Suspend
              </Button>
            )}
            <Button variant="danger" size="sm" onClick={() => setShowDelete(true)}>
              Delete
            </Button>
          </div>
        }
      />

      <PageContent className="space-y-6">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-px bg-border border border-border">
          <InfoCell label="Status" value={<StatusBadge status={service.Status} />} />
          <InfoCell
            label="URL"
            value={
              service.HostPort ? (
                <a
                  href={getServiceUrl(service.HostPort)}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-fg-secondary hover:text-fg hover:underline underline-offset-2"
                >
                  {getServiceUrl(service.HostPort)}
                </a>
              ) : (
                '—'
              )
            }
          />
          <InfoCell label="Plan" value={`${service.TierName} · ${service.TierCPU} CPU`} />
          <InfoCell label="Port" value={service.HostPort ? `${service.HostPort} → ${service.ContainerPort}` : '—'} />
        </div>

        <Card title="Details" noPadding>
          <dl className="divide-y divide-border">
            <DetailRow label="Repository" value={service.RepoURL} />
            <DetailRow label="Service ID" value={service.ID} mono />
            <DetailRow label="Container" value={service.ContainerID || '—'} mono />
            <DetailRow label="Image" value={service.ImageTag || '—'} mono />
            <DetailRow label="Created" value={formatDate(service.CreatedAt)} />
            <DetailRow label="Updated" value={formatDate(service.UpdatedAt)} />
            {service.Error && (
              <DetailRow label="Error" value={service.Error} error />
            )}
          </dl>
        </Card>

        {Object.keys(env).length > 0 && (
          <Card title="Environment" noPadding>
            <dl className="divide-y divide-border">
              {Object.entries(env).map(([key, value]) => (
                <DetailRow key={key} label={key} value={value} mono />
              ))}
            </dl>
          </Card>
        )}

        <div>
          <h2 className="text-sm font-medium text-fg mb-3">Deploy logs</h2>
          <LogViewer jobId={service.ID} />
        </div>
      </PageContent>

      <ConfirmModal
        open={showDelete}
        onClose={() => setShowDelete(false)}
        onConfirm={handleDelete}
        title="Delete service"
        message="This permanently stops and removes the container. This cannot be undone."
        confirmLabel="Delete service"
        danger
        loading={actionLoading}
      />
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

function DetailRow({ label, value, mono, error }) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-start gap-1 sm:gap-4 px-4 py-3">
      <dt className="text-xs text-muted sm:w-36 shrink-0">{label}</dt>
      <dd
        className={`text-sm flex-1 break-all ${mono ? 'font-mono text-xs text-fg-secondary' : 'text-fg'} ${error ? 'text-red-400' : ''}`}
      >
        {value}
      </dd>
    </div>
  )
}
