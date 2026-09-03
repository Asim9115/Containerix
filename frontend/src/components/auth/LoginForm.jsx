import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../../store/hooks'
import { loginWithApiKey } from '../../store/slices/authSlice'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'

export function LoginForm() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const { loading, error } = useAppSelector((s) => s.auth)
  const [apiKey, setApiKey] = useState('')

  const handleSubmit = async (e) => {
    e.preventDefault()
    const result = await dispatch(loginWithApiKey(apiKey.trim()))
    if (loginWithApiKey.fulfilled.match(result)) {
      navigate('/')
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <Input
        label="API key"
        type="password"
        placeholder="ctx-..."
        value={apiKey}
        onChange={(e) => setApiKey(e.target.value)}
        required
      />
      {error && (
        <p className="text-xs text-red-400 border border-red-900/50 bg-red-950/20 px-3 py-2">
          {error}
        </p>
      )}
      <Button type="submit" className="w-full" loading={loading}>
        Sign in
      </Button>
    </form>
  )
}
