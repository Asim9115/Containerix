export function formatDate(dateStr) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString()
}

export function formatRelativeTime(dateStr) {
  if (!dateStr) return '—'
  const diff = Date.now() - new Date(dateStr).getTime()
  const minutes = Math.floor(diff / 60000)
  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  const hours = Math.floor(minutes / 60)
  if (hours < 24) return `${hours}h ago`
  const days = Math.floor(hours / 24)
  return `${days}d ago`
}

export function getRepoName(url) {
  if (!url) return 'Unknown'
  try {
    const parts = new URL(url).pathname.split('/').filter(Boolean)
    return parts.slice(-2).join('/') || url
  } catch {
    return url
  }
}

export function getServiceUrl(hostPort) {
  if (!hostPort) return null
  return `http://localhost:${hostPort}`
}

export function parseEnvJson(envJson) {
  if (!envJson) return {}
  try {
    return typeof envJson === 'string' ? JSON.parse(envJson) : envJson
  } catch {
    return {}
  }
}

export function formatBytes(bytes) {
  if (!bytes) return '0 B'
  const num = typeof bytes === 'string' ? parseInt(bytes, 10) : bytes
  if (isNaN(num)) return bytes
  const units = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let val = num
  while (val >= 1024 && i < units.length - 1) {
    val /= 1024
    i++
  }
  return `${val.toFixed(1)} ${units[i]}`
}

export const STATUS_COLORS = {
  running: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  building: 'bg-amber-500/15 text-amber-400 border-amber-500/30',
  queued: 'bg-blue-500/15 text-blue-400 border-blue-500/30',
  completed: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30',
  stopped: 'bg-zinc-500/15 text-zinc-400 border-zinc-500/30',
  failed: 'bg-red-500/15 text-red-400 border-red-500/30',
}

export function getStatusColor(status) {
  return STATUS_COLORS[status?.toLowerCase()] || STATUS_COLORS.stopped
}
