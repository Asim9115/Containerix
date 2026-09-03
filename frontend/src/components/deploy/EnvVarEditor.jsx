import { Button } from '../ui/Button'
import { createEmptyEnvRow } from '../../utils/env'

export function EnvVarEditor({ rows, onChange, error }) {
  const updateRow = (id, field, value) => {
    onChange(rows.map((row) => (row.id === id ? { ...row, [field]: value } : row)))
  }

  const removeRow = (id) => {
    if (rows.length === 1) {
      onChange([createEmptyEnvRow()])
      return
    }
    onChange(rows.filter((row) => row.id !== id))
  }

  const addRow = () => {
    onChange([...rows, createEmptyEnvRow()])
  }

  return (
    <div className="space-y-2">
      <div className="grid grid-cols-[1fr_1fr_auto] gap-2 items-center">
        <span className="text-[11px] font-medium text-muted uppercase tracking-wide">
          Key
        </span>
        <span className="text-[11px] font-medium text-muted uppercase tracking-wide">
          Value
        </span>
        <span className="w-8" />
      </div>

      {rows.map((row) => (
        <div key={row.id} className="grid grid-cols-[1fr_1fr_auto] gap-2 items-start">
          <input
            type="text"
            placeholder="NODE_ENV"
            value={row.key}
            onChange={(e) => updateRow(row.id, 'key', e.target.value)}
            className="w-full px-3 py-2 bg-surface border border-border rounded text-fg text-sm placeholder:text-neutral-600 focus:outline-none focus:border-neutral-500"
          />
          <input
            type="text"
            placeholder="production"
            value={row.value}
            onChange={(e) => updateRow(row.id, 'value', e.target.value)}
            className="w-full px-3 py-2 bg-surface border border-border rounded text-fg text-sm placeholder:text-neutral-600 focus:outline-none focus:border-neutral-500"
          />
          <button
            type="button"
            onClick={() => removeRow(row.id)}
            className="w-8 h-[38px] flex items-center justify-center text-muted hover:text-red-400 transition-colors"
            aria-label="Remove variable"
          >
            ×
          </button>
        </div>
      ))}

      <Button type="button" variant="secondary" size="sm" onClick={addRow}>
        Add variable
      </Button>

      {error && <p className="text-xs text-red-400">{error}</p>}
    </div>
  )
}
