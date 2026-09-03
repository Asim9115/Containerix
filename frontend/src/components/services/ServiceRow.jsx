import { Link } from 'react-router-dom'
import { StatusBadge } from '../ui/Badge'
import { getRepoName, formatRelativeTime, getServiceUrl } from '../../utils/format'

export function ServiceRow({ service, showActions, onStop, onDelete }) {
  return (
    <div className="flex items-center gap-4 px-4 py-3 border-b border-border last:border-b-0 hover:bg-surface-hover transition-colors group">
      <Link to={`/services/${service.ID}`} className="flex items-center gap-3 min-w-0 flex-1">
        <div className="w-8 h-8 border border-border bg-surface flex items-center justify-center shrink-0">
          <svg className="w-4 h-4 text-muted" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2" />
          </svg>
        </div>
        <div className="min-w-0">
          <p className="text-sm text-fg truncate group-hover:underline underline-offset-2">
            {getRepoName(service.RepoURL)}
          </p>
          <p className="text-xs text-muted font-mono truncate">{service.ID}</p>
        </div>
      </Link>

      <div className="hidden sm:block w-24 shrink-0">
        <StatusBadge status={service.Status} />
      </div>

      <div className="hidden md:block w-32 shrink-0 text-xs text-muted">
        {service.HostPort ? (
          <a
            href={getServiceUrl(service.HostPort)}
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-fg transition-colors"
            onClick={(e) => e.stopPropagation()}
          >
            :{service.HostPort}
          </a>
        ) : (
          '—'
        )}
      </div>

      <div className="hidden lg:block w-28 shrink-0 text-xs text-muted">
        {formatRelativeTime(service.CreatedAt)}
      </div>

      {showActions && (
        <div className="flex items-center gap-3 shrink-0">
          {service.Status === 'running' && service.ContainerID && (
            <button
              onClick={() => onStop?.(service.ContainerID)}
              className="text-xs text-fg-secondary hover:text-fg transition-colors"
            >
              Suspend
            </button>
          )}
          <button
            onClick={() => onDelete?.(service)}
            className="text-xs text-red-400/80 hover:text-red-400 transition-colors"
          >
            Delete
          </button>
        </div>
      )}
    </div>
  )
}

export function ServiceListHeader() {
  return (
    <div className="hidden sm:flex items-center gap-4 px-4 py-2 border-b border-border bg-surface text-[11px] font-medium text-muted uppercase tracking-wider">
      <div className="flex-1">Service</div>
      <div className="w-24 shrink-0">Status</div>
      <div className="hidden md:block w-32 shrink-0">URL</div>
      <div className="hidden lg:block w-28 shrink-0">Created</div>
      <div className="w-24 shrink-0" />
    </div>
  )
}
