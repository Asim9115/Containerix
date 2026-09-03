import { Link } from 'react-router-dom'

export function Breadcrumbs({ items }) {
  if (!items?.length) return null

  return (
    <nav className="flex items-center gap-2 text-xs text-muted mb-2">
      {items.map((item, i) => (
        <span key={item.label} className="flex items-center gap-2">
          {i > 0 && <span className="text-neutral-700">/</span>}
          {item.to ? (
            <Link to={item.to} className="hover:text-fg-secondary transition-colors">
              {item.label}
            </Link>
          ) : (
            <span className="text-fg-secondary">{item.label}</span>
          )}
        </span>
      ))}
    </nav>
  )
}

export function PageHeader({ title, description, breadcrumbs, actions }) {
  return (
    <header className="border-b border-border bg-surface px-6 py-5">
      <Breadcrumbs items={breadcrumbs} />
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-lg font-medium text-fg tracking-tight">{title}</h1>
          {description && (
            <p className="text-sm text-muted mt-1">{description}</p>
          )}
        </div>
        {actions && <div className="flex items-center gap-2 shrink-0">{actions}</div>}
      </div>
    </header>
  )
}

export function PageContent({ children, className = '' }) {
  return <div className={`px-6 py-6 ${className}`}>{children}</div>
}
