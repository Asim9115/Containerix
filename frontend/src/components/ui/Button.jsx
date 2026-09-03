const variants = {
  primary:
    'bg-fg text-surface hover:bg-neutral-200 disabled:opacity-50 disabled:cursor-not-allowed',
  secondary:
    'bg-transparent text-fg-secondary border border-border hover:border-neutral-600 hover:text-fg disabled:opacity-50',
  danger:
    'bg-transparent text-red-400 border border-red-900/60 hover:bg-red-950/40 hover:border-red-800 disabled:opacity-50',
  ghost:
    'bg-transparent text-muted hover:text-fg hover:bg-surface-hover disabled:opacity-50',
  link:
    'bg-transparent text-fg-secondary hover:text-fg underline-offset-2 hover:underline p-0',
}

const sizes = {
  sm: 'px-2.5 py-1 text-xs',
  md: 'px-3 py-1.5 text-[13px]',
  lg: 'px-4 py-2 text-sm',
}

export function Button({
  children,
  variant = 'primary',
  size = 'md',
  className = '',
  loading = false,
  ...props
}) {
  const isLink = variant === 'link'

  return (
    <button
      className={`inline-flex items-center justify-center gap-1.5 rounded font-medium transition-colors ${
        isLink ? variants.link : `rounded ${variants[variant]} ${sizes[size]}`
      } ${className}`}
      disabled={loading || props.disabled}
      {...props}
    >
      {loading && (
        <svg className="animate-spin h-3.5 w-3.5" viewBox="0 0 24 24" fill="none">
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      )}
      {children}
    </button>
  )
}
