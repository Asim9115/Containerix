import { useEffect, useRef } from 'react'
import { streamSSE } from '../../api'

export function LogViewer({ jobId, className = '' }) {
  const containerRef = useRef(null)
  const logsRef = useRef([])

  useEffect(() => {
    if (!jobId) return

    const controller = new AbortController()
    logsRef.current = []

    const appendLog = (text) => {
      logsRef.current.push(text)
      if (containerRef.current) {
        containerRef.current.textContent = logsRef.current.join('\n')
        containerRef.current.scrollTop = containerRef.current.scrollHeight
      }
    }

    appendLog(`Connecting to log stream...`)

    streamSSE(`/containers/${jobId}/logs`, {
      signal: controller.signal,
      onEvent: ({ type, data }) => {
        if (type === 'log') appendLog(data)
        else if (type === 'deployed') appendLog(`\nDeployed: ${data}`)
        else if (type === 'error') appendLog(`\nError: ${data}`)
        else if (type === 'done') appendLog(`\n— ${data}`)
      },
      onError: (err) => appendLog(`\nConnection error: ${err.message}`),
    }).catch(() => {})

    return () => controller.abort()
  }, [jobId])

  return (
    <div className={`border border-border bg-black ${className}`}>
      <div className="flex items-center justify-between px-3 py-2 border-b border-border bg-surface-raised">
        <span className="text-xs text-muted">Logs</span>
        <span className="text-[11px] text-muted">Live</span>
      </div>
      <pre
        ref={containerRef}
        className="p-3 text-xs font-mono text-neutral-400 h-[420px] overflow-y-auto whitespace-pre-wrap break-words leading-relaxed"
      />
    </div>
  )
}
