import { useEffect, useState } from 'react'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import {
  fetchCgroup,
  destroyCgroup,
  fetchPorts,
  fetchHealth,
} from '../store/slices/adminSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { Card } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { ConfirmModal } from '../components/ui/Modal'
import { PageLoader } from '../components/ui/Loading'
import { formatDate, formatBytes } from '../utils/format'

export function AdminPage() {
  const dispatch = useAppDispatch()
  const {
    cgroup,
    ports,
    health,
    ready,
    loading,
    cgroupActionLoading,
    error,
  } = useAppSelector((s) => s.admin)
  const [showDestroy, setShowDestroy] = useState(false)

  useEffect(() => {
    dispatch(fetchHealth())
    dispatch(fetchCgroup())
    dispatch(fetchPorts())
    const interval = setInterval(() => {
      dispatch(fetchHealth())
      dispatch(fetchCgroup())
      dispatch(fetchPorts())
    }, 10000)
    return () => clearInterval(interval)
  }, [dispatch])

  const handleDestroyCgroup = async () => {
    await dispatch(destroyCgroup())
    setShowDestroy(false)
    dispatch(fetchCgroup())
  }

  const cgroupData = cgroup?.data || cgroup
  const portList = ports ?? []

  return (
    <div>
      <PageHeader
        title="System"
        description="Platform health and resource monitoring"
      />

      <PageContent className="space-y-6">
        {error && (
          <p className="text-sm text-red-400 border border-red-900/50 bg-red-950/20 px-3 py-2">
            {error}
          </p>
        )}

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <StatusCard
            label="Health"
            ok={health?.status === 'ok'}
            detail={health?.status === 'ok' ? 'Operational' : 'Degraded'}
          />
          <StatusCard
            label="Readiness"
            ok={ready?.status === 'ready'}
            detail={
              ready?.status === 'ready'
                ? 'Ready'
                : `Database: ${ready?.checks?.database || 'unknown'}`
            }
          />
        </div>

        <Card
          title="Sandbox"
          description="Cgroup resource allocation"
          action={
            <Button
              variant="danger"
              size="sm"
              onClick={() => setShowDestroy(true)}
              disabled={!cgroupData}
            >
              Reset sandbox
            </Button>
          }
        >
          {loading && !cgroupData ? (
            <PageLoader />
          ) : cgroupData ? (
            <div className="space-y-4">
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <Metric label="Name" value={cgroupData.name} />
                <Metric label="CPU limit" value={cgroupData.cpu} />
                <Metric label="Memory limit" value={formatBytes(cgroupData.memory)} />
                <Metric label="Used CPU" value={cgroupData.usedcpu?.toFixed(2)} />
                <Metric label="Used memory" value={formatBytes(cgroupData.usedmemory)} />
              </div>

              {cgroupData.containers && Object.keys(cgroupData.containers).length > 0 && (
                <div className="border border-border">
                  <div className="px-3 py-2 border-b border-border text-[11px] text-muted uppercase tracking-wider">
                    Containers
                  </div>
                  {Object.entries(cgroupData.containers).map(([id, c]) => (
                    <div
                      key={id}
                      className="flex items-center justify-between px-3 py-2 border-b border-border last:border-b-0 text-xs"
                    >
                      <span className="font-mono text-fg-secondary">{c.ID || id}</span>
                      <span className="text-muted">{c.CPU} CPU · {c.Memory}</span>
                      <span className="text-muted">{c.Status}</span>
                    </div>
                  ))}
                </div>
              )}
            </div>
          ) : (
            <p className="text-xs text-muted">No sandbox data</p>
          )}
        </Card>

        <Card title="Port allocations" noPadding>
          {portList.length === 0 ? (
            <p className="px-4 py-6 text-xs text-muted">No ports allocated</p>
          ) : (
            <div>
              <div className="hidden sm:flex px-4 py-2 border-b border-border text-[11px] text-muted uppercase tracking-wider">
                <div className="w-24">Host</div>
                <div className="flex-1">Container</div>
                <div className="w-24">Port</div>
                <div className="w-40">Allocated</div>
              </div>
              {portList.map((p, i) => (
                <div
                  key={i}
                  className="flex flex-col sm:flex-row sm:items-center px-4 py-2.5 border-b border-border last:border-b-0 text-xs gap-1 sm:gap-0"
                >
                  <div className="w-24 text-fg">{p.HostPort}</div>
                  <div className="flex-1 font-mono text-fg-secondary truncate">{p.ContainerID}</div>
                  <div className="w-24 text-muted">{p.ContainerPort}</div>
                  <div className="w-40 text-muted">{formatDate(p.AllocatedAt)}</div>
                </div>
              ))}
            </div>
          )}
        </Card>
      </PageContent>

      <ConfirmModal
        open={showDestroy}
        onClose={() => setShowDestroy(false)}
        onConfirm={handleDestroyCgroup}
        title="Reset sandbox"
        message="Destroys the cgroup and releases sandbox resources. Running containers may be affected."
        confirmLabel="Reset"
        danger
        loading={cgroupActionLoading}
      />
    </div>
  )
}

function StatusCard({ label, ok, detail }) {
  return (
    <div className="border border-border px-4 py-3">
      <p className="text-[11px] text-muted uppercase tracking-wide mb-2">{label}</p>
      <div className="flex items-center gap-2">
        <span className={`w-1.5 h-1.5 rounded-full ${ok ? 'bg-green-500' : 'bg-red-500'}`} />
        <span className="text-sm text-fg">{detail}</span>
      </div>
    </div>
  )
}

function Metric({ label, value }) {
  return (
    <div>
      <p className="text-[11px] text-muted mb-0.5">{label}</p>
      <p className="text-sm text-fg">{value ?? '—'}</p>
    </div>
  )
}
