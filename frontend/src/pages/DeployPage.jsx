import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '../store/hooks'
import { triggerBuild, clearBuildError } from '../store/slices/buildSlice'
import { PageHeader, PageContent } from '../components/layout/PageHeader'
import { Section } from '../components/ui/Card'
import { Button } from '../components/ui/Button'
import { Input, Select } from '../components/ui/Input'
import { EnvVarEditor } from '../components/deploy/EnvVarEditor'
import {
  createEmptyEnvRow,
  envRowsToMap,
  findDuplicateEnvKeys,
} from '../utils/env'

/** Runtime presets — match server templates / Render-style defaults. */
const RUNTIME_PRESETS = {
  docker: {
    label: 'Docker',
    buildCommand: '',
    startCommand: '',
  },
  node: {
    label: 'Node',
    buildCommand: 'npm ci',
    startCommand: 'npm start',
  },
  python: {
    label: 'Python',
    buildCommand: 'pip install -r requirements.txt',
    startCommand: 'gunicorn app:app --bind 0.0.0.0:$PORT',
  },
  go: {
    label: 'Go',
    buildCommand: 'go build -o server .',
    startCommand: './server',
  },
  static: {
    label: 'Static site',
    buildCommand: 'npm ci && npm run build',
    startCommand: 'npx --yes serve -s dist -l $PORT',
  },
}

export function DeployPage() {
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const { loading, error } = useAppSelector((s) => s.build)

  const [url, setUrl] = useState('')
  const [rootDirectory, setRootDirectory] = useState('')
  const [runtime, setRuntime] = useState('docker')
  const [dockerfilePath, setDockerfilePath] = useState('Dockerfile')
  const [buildCommand, setBuildCommand] = useState('')
  const [startCommand, setStartCommand] = useState('')
  const [port, setPort] = useState('10000')
  const [healthCheckPath, setHealthCheckPath] = useState('')
  const [tier, setTier] = useState('tier1')
  const [envRows, setEnvRows] = useState([createEmptyEnvRow()])
  const [envError, setEnvError] = useState('')
  const [formError, setFormError] = useState('')

  const isDocker = runtime === 'docker'

  const runtimeHint = useMemo(() => {
    if (isDocker) {
      return 'Uses an existing Dockerfile in the repo (or the path you set).'
    }
    return 'No Dockerfile needed — we generate one from language + build + start.'
  }, [isDocker])

  const handleRuntimeChange = (next) => {
    setRuntime(next)
    setFormError('')
    const preset = RUNTIME_PRESETS[next]
    if (!preset) return
    if (next === 'docker') {
      setBuildCommand('')
      setStartCommand('')
      if (!dockerfilePath.trim()) setDockerfilePath('Dockerfile')
      return
    }
    setBuildCommand(preset.buildCommand)
    setStartCommand(preset.startCommand)
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    dispatch(clearBuildError())
    setEnvError('')
    setFormError('')

    const duplicates = findDuplicateEnvKeys(envRows)
    if (duplicates.length > 0) {
      setEnvError(`Duplicate keys: ${duplicates.join(', ')}`)
      return
    }

    const portNum = Number.parseInt(port, 10)
    if (!Number.isFinite(portNum) || portNum < 1 || portNum > 65535) {
      setFormError('Port must be a number between 1 and 65535')
      return
    }

    if (!isDocker) {
      if (!buildCommand.trim() || !startCommand.trim()) {
        setFormError('Build command and start command are required for native runtimes')
        return
      }
    }

    const payload = {
      url: url.trim(),
      tier,
      port: portNum,
    }

    if (rootDirectory.trim()) {
      payload.root_directory = rootDirectory.trim()
    }

    if (healthCheckPath.trim()) {
      payload.health_check_path = healthCheckPath.trim()
    }

    if (isDocker) {
      if (dockerfilePath.trim()) {
        payload.dockerfile_path = dockerfilePath.trim()
      }
    } else {
      payload.language = runtime
      payload.build_command = buildCommand.trim()
      payload.start_command = startCommand.trim()
    }

    const env = envRowsToMap(envRows)
    if (Object.keys(env).length > 0) {
      payload.env = env
    }

    const result = await dispatch(triggerBuild(payload))
    if (triggerBuild.fulfilled.match(result)) {
      navigate(`/jobs/${result.payload.job_id}`)
    }
  }

  return (
    <div>
      <PageHeader
        title="New Web Service"
        description="Deploy from a public GitHub repository"
        breadcrumbs={[
          { label: 'Services', to: '/services' },
          { label: 'New' },
        ]}
      />

      <PageContent className="max-w-2xl">
        <form onSubmit={handleSubmit} className="space-y-8">
          <Section
            title="Source"
            description="Containerix clones the repository and builds a Docker image."
          >
            <div className="space-y-4">
              <Input
                label="Repository URL"
                placeholder="https://github.com/owner/repo"
                value={url}
                onChange={(e) => setUrl(e.target.value)}
                hint="Public GitHub repositories only"
                required
              />
              <Input
                label="Root directory"
                placeholder="(optional) e.g. apps/api"
                value={rootDirectory}
                onChange={(e) => setRootDirectory(e.target.value)}
                hint="Subdirectory used as the Docker build context"
              />
            </div>
          </Section>

          <Section
            title="Build & start"
            description="Same model as Render: Dockerfile, or language + build + start commands."
          >
            <div className="space-y-4">
              <Select
                label="Runtime"
                value={runtime}
                onChange={(e) => handleRuntimeChange(e.target.value)}
                hint={runtimeHint}
              >
                {Object.entries(RUNTIME_PRESETS).map(([value, preset]) => (
                  <option key={value} value={value}>
                    {preset.label}
                  </option>
                ))}
              </Select>

              {isDocker ? (
                <Input
                  label="Dockerfile path"
                  placeholder="Dockerfile"
                  value={dockerfilePath}
                  onChange={(e) => setDockerfilePath(e.target.value)}
                  hint="Relative to the repo (or root directory). Leave as Dockerfile for the default."
                />
              ) : (
                <>
                  <Input
                    label="Build command"
                    placeholder="npm ci"
                    value={buildCommand}
                    onChange={(e) => setBuildCommand(e.target.value)}
                    hint="Runs during image build (RUN)"
                    required
                  />
                  <Input
                    label="Start command"
                    placeholder="npm start"
                    value={startCommand}
                    onChange={(e) => setStartCommand(e.target.value)}
                    hint="Process entrypoint — must listen on 0.0.0.0 and $PORT"
                    required
                  />
                </>
              )}

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <Input
                  label="Port"
                  type="number"
                  min={1}
                  max={65535}
                  value={port}
                  onChange={(e) => setPort(e.target.value)}
                  hint="Injected as PORT (default 10000)"
                  required
                />
                <Input
                  label="Health check path"
                  placeholder="(optional) e.g. /"
                  value={healthCheckPath}
                  onChange={(e) => setHealthCheckPath(e.target.value)}
                  hint="HTTP GET after TCP is open; 2xx/3xx = ready"
                />
              </div>

              <p className="text-xs text-muted border border-border rounded px-3 py-2 bg-surface">
                Your process must bind <code className="text-fg">0.0.0.0</code> and listen on{' '}
                <code className="text-fg">$PORT</code> (default 10000).
              </p>
            </div>
          </Section>

          <Section
            title="Instance type"
            description="Resource limits applied to the running container."
          >
            <Select
              label="Plan"
              value={tier}
              onChange={(e) => setTier(e.target.value)}
            >
              <option value="tier1">Starter — 0.2 CPU, 500 MB RAM</option>
              <option value="tier2">Standard — 0.5 CPU, 750 MB RAM</option>
            </Select>
          </Section>

          <Section
            title="Environment variables"
            description="Optional key-value pairs passed to the container at runtime."
          >
            <EnvVarEditor
              rows={envRows}
              onChange={setEnvRows}
              error={envError}
            />
          </Section>

          {(formError || error) && (
            <p className="text-sm text-red-400 border border-red-900/50 bg-red-950/20 px-3 py-2">
              {formError || error}
            </p>
          )}

          <div className="flex items-center gap-3 pt-2 border-t border-border">
            <Button type="submit" loading={loading}>
              Create Web Service
            </Button>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => navigate('/services')}
            >
              Cancel
            </Button>
          </div>
        </form>
      </PageContent>
    </div>
  )
}
