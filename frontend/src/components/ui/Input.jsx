const fieldClass =
  'w-full px-3 py-2 bg-surface border border-border rounded text-fg text-sm placeholder:text-neutral-600 focus:outline-none focus:border-neutral-500 transition-colors'

export function Input({ label, hint, error, className = '', ...props }) {
  return (
    <div className="space-y-1.5">
      {label && (
        <label className="block text-xs font-medium text-fg-secondary uppercase tracking-wide">
          {label}
        </label>
      )}
      <input
        className={`${fieldClass} ${error ? 'border-red-800' : ''} ${className}`}
        {...props}
      />
      {hint && !error && <p className="text-xs text-muted">{hint}</p>}
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
}

export function Textarea({ label, hint, error, className = '', ...props }) {
  return (
    <div className="space-y-1.5">
      {label && (
        <label className="block text-xs font-medium text-fg-secondary uppercase tracking-wide">
          {label}
        </label>
      )}
      <textarea
        className={`${fieldClass} resize-y min-h-[96px] font-mono text-xs ${error ? 'border-red-800' : ''} ${className}`}
        {...props}
      />
      {hint && !error && <p className="text-xs text-muted">{hint}</p>}
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
}

export function Select({ label, hint, error, children, className = '', ...props }) {
  return (
    <div className="space-y-1.5">
      {label && (
        <label className="block text-xs font-medium text-fg-secondary uppercase tracking-wide">
          {label}
        </label>
      )}
      <select
        className={`${fieldClass} ${error ? 'border-red-800' : ''} ${className}`}
        {...props}
      >
        {children}
      </select>
      {hint && !error && <p className="text-xs text-muted">{hint}</p>}
      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
}
