import { Outlet, useNavigate } from 'react-router-dom'
import { config } from '../../config/env'
import { useAppDispatch } from '../../store/hooks'
import { logoutAdmin } from '../../store/slices/adminSlice'

export function AdminLayout() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()

  const handleLogout = () => {
    dispatch(logoutAdmin())
    navigate('/admin/login')
  }

  return (
    <div className="min-h-screen bg-surface">
      <header className="border-b border-border bg-sidebar">
        <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
          <div>
            <p className="text-sm font-semibold text-fg tracking-tight">
              {config.appName}
            </p>
            <p className="text-[11px] text-muted uppercase tracking-wider">
              Admin console
            </p>
          </div>
          <button
            type="button"
            onClick={handleLogout}
            className="text-xs text-muted hover:text-fg-secondary transition-colors"
          >
            Sign out
          </button>
        </div>
      </header>
      <main className="max-w-5xl mx-auto">
        <Outlet />
      </main>
    </div>
  )
}
