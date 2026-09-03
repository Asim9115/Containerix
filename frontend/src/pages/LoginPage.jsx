import { useState } from 'react'
import { config } from '../config/env'
import { LoginForm } from '../components/auth/LoginForm'
import { RegisterForm } from '../components/auth/RegisterForm'

export function LoginPage() {
  const [tab, setTab] = useState('login')

  return (
    <div className="min-h-screen bg-sidebar flex">
      <div className="hidden lg:flex flex-1 items-center justify-center border-r border-border p-12">
        <div className="max-w-sm">
          <h1 className="text-2xl font-medium text-fg tracking-tight">{config.appName}</h1>
          <p className="text-sm text-muted mt-3 leading-relaxed">
            Deploy web services from Git repositories. Build, run, and manage containers from a single dashboard.
          </p>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-6">
        <div className="w-full max-w-sm">
          <div className="lg:hidden mb-8">
            <h1 className="text-lg font-medium text-fg">{config.appName}</h1>
          </div>

          <div className="border border-border bg-surface-raised p-6">
            <div className="flex border-b border-border mb-6">
              <button
                onClick={() => setTab('login')}
                className={`flex-1 pb-2.5 text-sm transition-colors border-b-2 -mb-px ${
                  tab === 'login'
                    ? 'text-fg border-fg'
                    : 'text-muted border-transparent hover:text-fg-secondary'
                }`}
              >
                Sign in
              </button>
              <button
                onClick={() => setTab('register')}
                className={`flex-1 pb-2.5 text-sm transition-colors border-b-2 -mb-px ${
                  tab === 'register'
                    ? 'text-fg border-fg'
                    : 'text-muted border-transparent hover:text-fg-secondary'
                }`}
              >
                Create account
              </button>
            </div>

            {tab === 'login' ? (
              <>
                <LoginForm />
                <p className="text-xs text-muted mt-4">
                  Enter the API key from your account settings.
                </p>
              </>
            ) : (
              <RegisterForm />
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
