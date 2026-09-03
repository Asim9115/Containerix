export function Card({ children, className = '', title, description, action, noPadding = false }) {
  return (
    <div className={`border border-border bg-surface-raised ${className}`}>
      {(title || action) && (
        <div className="flex items-start justify-between gap-4 px-4 py-3 border-b border-border">
          <div>
            {title && <h3 className="text-sm font-medium text-fg">{title}</h3>}
            {description && <p className="text-xs text-muted mt-0.5">{description}</p>}
          </div>
          {action}
        </div>
      )}
      <div className={noPadding ? '' : 'p-4'}>{children}</div>
    </div>
  )
}

export function Section({ title, description, children, className = '' }) {
  return (
    <section className={className}>
      {(title || description) && (
        <div className="mb-4">
          {title && <h2 className="text-sm font-medium text-fg">{title}</h2>}
          {description && <p className="text-xs text-muted mt-1">{description}</p>}
        </div>
      )}
      {children}
    </section>
  )
}
