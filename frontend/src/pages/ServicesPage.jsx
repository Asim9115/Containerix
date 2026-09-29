import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import {
  fetchDeployments,
  deleteDeployment,
  clearDeploymentError,
} from '../store/slices/deploymentsSlice'
import { startContainer, stopContainer, stopAllContainers } from '../store/slices/containersSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { Button } from '../components/ui/Button'
import { ConfirmModal } from '../components/ui/Modal'
import { PageLoader, EmptyState } from '../components/ui/Loading'
import { ServiceListHeader, ServiceRow } from '../components/services/ServiceRow'
import { getRepoName } from '../utils/format'

export function ServicesPage() {
  const dispatch = useAppDispatch()
  const items = useAppSelector((s) => s.deployments.items) ?? []
  const loading = useAppSelector((s) => s.deployments.loading)
  const error = useAppSelector((s) => s.deployments.error)
  const [deleteTarget, setDeleteTarget] = useState(null)
  const [actionLoading, setActionLoading] = useState(false)
  const [actionError, setActionError] = useState(null)

  useEffect(() => {
    dispatch(fetchDeployments())
    const interval = setInterval(() => dispatch(fetchDeployments()), 10000)
    return () => clearInterval(interval)
  }, [dispatch])

  const handleDelete = async () => {
    if (!deleteTarget) return
    setActionLoading(true)
    dispatch(clearDeploymentError())
    const result = await dispatch(deleteDeployment(deleteTarget.ID))
    setActionLoading(false)
    if (deleteDeployment.fulfilled.match(result)) {
      setDeleteTarget(null)
    }
  }

  const handleStop = async (containerId) => {
    setActionLoading(true)
    setActionError(null)
    const result = await dispatch(stopContainer(containerId))
    if (stopContainer.fulfilled.match(result)) {
      dispatch(fetchDeployments())
    } else {
      setActionError(result.payload || 'Failed to stop service')
    }
    setActionLoading(false)
  }

  const handleStart = async (deploymentId) => {
    setActionLoading(true)
    setActionError(null)
    const result = await dispatch(startContainer(deploymentId))
    if (startContainer.fulfilled.match(result)) {
      dispatch(fetchDeployments())
    } else {
      setActionError(result.payload || 'Failed to start service')
    }
    setActionLoading(false)
  }

  const handleStopAll = async () => {
    setActionLoading(true)
    setActionError(null)
    const result = await dispatch(stopAllContainers())
    if (stopAllContainers.fulfilled.match(result)) {
      dispatch(fetchDeployments())
    } else {
      setActionError(result.payload || 'Failed to suspend services')
    }
    setActionLoading(false)
  }

  if (loading && items.length === 0) return <PageLoader />

  return (
    <div>
      <PageHeader
        title="Services"
        description="Web services deployed from your Git repositories"
        actions={
          <div className="flex items-center gap-2">
            {items.some((s) => s.Status === 'running') && (
              <Button variant="secondary" size="sm" onClick={handleStopAll} loading={actionLoading}>
                Suspend all
              </Button>
            )}
            <Link to="/deploy">
              <Button size="sm">New Web Service</Button>
            </Link>
          </div>
        }
      />

      <PageContent>
        {(actionError || error) && (
          <p className="text-sm text-red-400 border border-red-900/50 bg-red-950/20 px-3 py-2 mb-4">
            {actionError || error}
          </p>
        )}

        {items.length === 0 ? (
          <EmptyState
            title="No services"
            description="Connect a Git repository to deploy your first web service."
            action={
              <Link to="/deploy">
                <Button size="sm">New Web Service</Button>
              </Link>
            }
          />
        ) : (
          <div className="border border-border">
            <ServiceListHeader />
            {items.map((service) => (
              <ServiceRow
                key={service.ID}
                service={service}
                showActions
                actionLoading={actionLoading}
                onStart={handleStart}
                onStop={handleStop}
                onDelete={setDeleteTarget}
              />
            ))}
          </div>
        )}
      </PageContent>

      <ConfirmModal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
        title="Delete service"
        message={`Delete "${getRepoName(deleteTarget?.RepoURL)}"? The container will be stopped and removed permanently.`}
        confirmLabel="Delete service"
        danger
        loading={actionLoading}
      />
    </div>
  )
}
