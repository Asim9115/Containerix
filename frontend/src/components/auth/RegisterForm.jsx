import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../../store/hooks'
import { registerUser } from '../../store/slices/authSlice'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Modal } from '../ui/Modal'

export function RegisterForm() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const { loading, error, newApiKey } = useAppSelector((s) => s.auth)
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [showKeyModal, setShowKeyModal] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleSubmit = async (e) => {
    e.preventDefault()
    const result = await dispatch(registerUser({ name, email }))
    if (registerUser.fulfilled.match(result)) {
      setShowKeyModal(true)
    }
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(newApiKey)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleContinue = () => {
    setShowKeyModal(false)
    navigate('/')
  }

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Name"
          placeholder="Jane Doe"
          value={name}
          onChange={(e) => setName(e.target.value)}
          required
        />
        <Input
          label="Email"
          type="email"
          placeholder="you@example.com"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          required
        />
        {error && (
          <p className="text-xs text-red-400 border border-red-900/50 bg-red-950/20 px-3 py-2">
            {error}
          </p>
        )}
        <Button type="submit" className="w-full" loading={loading}>
          Create account
        </Button>
      </form>

      <Modal
        open={showKeyModal}
        onClose={() => {}}
        title="Your API key"
        footer={<Button size="sm" onClick={handleContinue}>Continue</Button>}
      >
        <p className="text-xs text-muted mb-3">
          Copy this key now. It will not be shown again.
        </p>
        <div className="flex gap-2">
          <code className="flex-1 px-3 py-2 border border-border bg-surface text-xs text-fg-secondary break-all font-mono">
            {newApiKey}
          </code>
          <Button variant="secondary" size="sm" onClick={handleCopy}>
            {copied ? 'Copied' : 'Copy'}
          </Button>
        </div>
      </Modal>
    </>
  )
}
