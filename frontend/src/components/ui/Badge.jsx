const STATUS = {
  running: { dot: 'bg-green-500', label: 'Live' },
  building: { dot: 'bg-yellow-500', label: 'Building' },
  queued: { dot: 'bg-blue-400', label: 'Queued' },
  completed: { dot: 'bg-green-500', label: 'Completed' },
  stopped: { dot: 'bg-neutral-500', label: 'Stopped' },
  failed: { dot: 'bg-red-500', label: 'Failed' },
}

export function StatusBadge({ status }) {
  const key = status?.toLowerCase()
  const config = STATUS[key] || { dot: 'bg-neutral-500', label: status || 'Unknown' }

  return (
    <span className="inline-flex items-center gap-1.5 text-xs text-fg-secondary">
      <span className={`w-1.5 h-1.5 rounded-full ${config.dot}`} />
      <span className="capitalize">{config.label}</span>
    </span>
  )
}

/** @deprecated use StatusBadge */
export function Badge({ status, children }) {
  return <StatusBadge status={children || status} />
}
