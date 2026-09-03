import { NavLink, useNavigate } from 'react-router-dom'
import { config } from '../../config/env'
import { navItems } from '../../config/routes'
import { NavIcon } from './NavIcon'
import { useAppDispatch, useAppSelector } from '../../store/hooks'
import { logout } from '../../store/slices/authSlice'

export function Sidebar() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const user = useAppSelector((s) => s.auth.user)

  const handleLogout = () => {
    dispatch(logout())
    navigate('/login')
  }

  return (
    <aside className="w-[220px] shrink-0 bg-sidebar border-r border-border flex flex-col h-screen sticky top-0">
      <div className="px-4 py-4 border-b border-border">
        <span className="text-sm font-semibold text-fg tracking-tight">
          {config.appName}
        </span>
      </div>

      <nav className="flex-1 px-2 py-3 space-y-0.5 overflow-y-auto">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            end={item.path === '/'}
            className={({ isActive }) =>
              `relative flex items-center gap-2.5 px-3 py-2 text-[13px] transition-colors ${
                isActive
                  ? 'text-fg bg-surface-hover before:absolute before:left-0 before:top-1 before:bottom-1 before:w-0.5 before:bg-fg before:rounded-r'
                  : 'text-muted hover:text-fg-secondary hover:bg-surface-hover/60'
              }`
            }
          >
            <NavIcon name={item.icon} className="w-4 h-4 shrink-0 opacity-70" />
            {item.label}
          </NavLink>
        ))}
      </nav>

      {user && (
        <div className="px-3 py-3 border-t border-border">
          <div className="px-2 py-2 mb-1">
            <p className="text-[13px] text-fg truncate">{user.name || 'User'}</p>
            <p className="text-xs text-muted truncate">{user.email}</p>
          </div>
          <button
            onClick={handleLogout}
            className="w-full text-left px-2 py-1.5 text-xs text-muted hover:text-fg-secondary transition-colors"
          >
            Sign out
          </button>
        </div>
      )}
    </aside>
  )
}
