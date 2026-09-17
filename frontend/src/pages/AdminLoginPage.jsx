import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { config } from '../config/env'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { loginAdmin, clearAdminError } from '../store/slices/adminSlice'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'

export function AdminLoginPage() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const { authLoading, error } = useAppSelector((s) => s.admin)
  const [apiKey, setApiKey] = useState('')

  const handleSubmit = async (e) => {
    e.preventDefault()
    dispatch(clearAdminError())
    const result = await dispatch(loginAdmin(apiKey))
    if (loginAdmin.fulfilled.match(result)) {
      navigate('/admin')
    }
  }

  return (
    <div className="min-h-screen bg-sidebar flex items-center justify-center p-6">
      <div className="w-full max-w-sm">
        <div className="mb-8">
          <h1 className="text-lg font-medium text-fg">{config.appName}</h1>
          <p className="text-xs text-muted mt-1">Admin console — not for end users</p>
        </div>

        <div className="border border-border bg-surface-raised p-6">
          <h2 className="text-sm text-fg mb-4">Sign in with admin API key</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Admin API key"
              type="password"
              placeholder="CONTAINERIX_ADMIN_API_KEY"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              required
            />
            {error && (
              <p className="text-xs text-red-400 border border-red-900/50 bg-red-950/20 px-3 py-2">
                {error}
              </p>
            )}
            <Button type="submit" className="w-full" loading={authLoading}>
              Sign in
            </Button>
          </form>
          <p className="text-xs text-muted mt-4">
            Use the server env key. User API keys cannot access this console.
          </p>
        </div>
      </div>
    </div>
  )
}
