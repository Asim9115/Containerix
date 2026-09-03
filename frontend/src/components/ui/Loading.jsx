export function LoadingSpinner({ size = 'md', className = '' }) {
  const sizes = { sm: 'h-4 w-4', md: 'h-6 w-6', lg: 'h-8 w-8' }
  return (
    <div className={`flex items-center justify-center ${className}`}>
      <svg
        className={`animate-spin text-muted ${sizes[size]}`}
        viewBox="0 0 24 24"
        fill="none"
      >
        <circle
          className="opacity-25"
          cx="12"
          cy="12"
          r="10"
          stroke="currentColor"
          strokeWidth="4"
        />
        <path
          className="opacity-75"
          fill="currentColor"
          d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z"
        />
      </svg>
    </div>
  )
}

export function PageLoader() {
  return (
    <div className="flex items-center justify-center min-h-[320px]">
      <LoadingSpinner size="lg" />
    </div>
  )
}

export function EmptyState({ title, description, action }) {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center border border-dashed border-border">
      <h3 className="text-sm font-medium text-fg mb-1">{title}</h3>
      {description && (
        <p className="text-xs text-muted max-w-xs mb-4 leading-relaxed">{description}</p>
      )}
      {action}
    </div>
  )
}
