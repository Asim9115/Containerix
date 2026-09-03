import { useState } from 'react'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { rotateApiKey, clearNewApiKey } from '../store/slices/authSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { Card } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { ConfirmModal, Modal } from '../components/ui/Modal'
import { formatDate } from '../utils/format'

export function SettingsPage() {
  const dispatch = useAppDispatch()
  const { user, loading, newApiKey } = useAppSelector((s) => s.auth)
  const [showRotate, setShowRotate] = useState(false)
  const [showKeyModal, setShowKeyModal] = useState(false)
  const [copied, setCopied] = useState(false)

  const handleRotate = async () => {
    const result = await dispatch(rotateApiKey())
    if (rotateApiKey.fulfilled.match(result)) {
      setShowRotate(false)
      setShowKeyModal(true)
    }
  }

  const handleCopy = () => {
    navigator.clipboard.writeText(newApiKey)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const handleCloseKeyModal = () => {
    setShowKeyModal(false)
    dispatch(clearNewApiKey())
  }

  return (
    <div>
      <PageHeader title="Account settings" description="Manage your profile and API credentials" />

      <PageContent className="max-w-lg space-y-6">
        <Card title="Profile" noPadding>
          <dl className="divide-y divide-border">
            <SettingRow label="Name" value={user?.name || '—'} />
            <SettingRow label="Email" value={user?.email || '—'} />
            <SettingRow label="User ID" value={user?.id || '—'} mono />
            <SettingRow label="Member since" value={formatDate(user?.created_at)} />
          </dl>
        </Card>

        <Card title="API key">
          <p className="text-xs text-muted leading-relaxed mb-4">
            Used to authenticate API requests. Rotating invalidates the current key immediately.
          </p>
          <Button variant="secondary" size="sm" onClick={() => setShowRotate(true)}>
            Rotate API key
          </Button>
        </Card>
      </PageContent>

      <ConfirmModal
        open={showRotate}
        onClose={() => setShowRotate(false)}
        onConfirm={handleRotate}
        title="Rotate API key"
        message="Your current key will stop working immediately. Copy the new key before closing the dialog."
        confirmLabel="Rotate key"
        danger
        loading={loading}
      />

      <Modal
        open={showKeyModal}
        onClose={handleCloseKeyModal}
        title="New API key"
        footer={<Button size="sm" onClick={handleCloseKeyModal}>Done</Button>}
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
    </div>
  )
}

function SettingRow({ label, value, mono }) {
  return (
    <div className="flex flex-col sm:flex-row sm:items-center gap-1 sm:gap-4 px-4 py-3">
      <dt className="text-xs text-muted sm:w-32 shrink-0">{label}</dt>
      <dd className={`text-sm text-fg ${mono ? 'font-mono text-xs text-fg-secondary' : ''}`}>
        {value}
      </dd>
    </div>
  )
}
