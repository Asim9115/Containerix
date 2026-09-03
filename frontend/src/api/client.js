import { config } from '../config/env'

export class ApiError extends Error {
  constructor(message, status, data = null) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.data = data
  }
}

export function getStoredApiKey() {
  return localStorage.getItem(config.apiKeyStorageKey)
}

export function setStoredApiKey(key) {
  if (key) {
    localStorage.setItem(config.apiKeyStorageKey, key)
  } else {
    localStorage.removeItem(config.apiKeyStorageKey)
  }
}

export async function apiRequest(path, options = {}) {
  const { auth = true, body, headers = {}, ...rest } = options

  const reqHeaders = { ...headers }

  if (body !== undefined && body !== null) {
    reqHeaders['Content-Type'] = 'application/json'
  }

  if (auth) {
    const apiKey = getStoredApiKey()
    if (!apiKey) {
      throw new ApiError('Not authenticated', 401)
    }
    reqHeaders['X-API-Key'] = apiKey
  }

  const url = `${config.apiBaseUrl}${path}`

  const response = await fetch(url, {
    ...rest,
    headers: reqHeaders,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  if (!response.ok) {
    let data = null
    try {
      data = await response.json()
    } catch {
      // response may not be JSON
    }
    throw new ApiError(data?.error || response.statusText, response.status, data)
  }

  if (response.status === 204) {
    return null
  }

  const contentType = response.headers.get('content-type')
  if (contentType?.includes('application/json')) {
    return response.json()
  }

  return response.text()
}

/**
 * Fetch-based SSE client — EventSource cannot send custom auth headers.
 */
export async function streamSSE(path, { onEvent, onError, signal }) {
  const apiKey = getStoredApiKey()
  if (!apiKey) {
    throw new ApiError('Not authenticated', 401)
  }

  const response = await fetch(`${config.apiBaseUrl}${path}`, {
    headers: {
      'X-API-Key': apiKey,
      Accept: 'text/event-stream',
    },
    signal,
  })

  if (!response.ok) {
    let data = null
    try {
      data = await response.json()
    } catch {
      // ignore
    }
    throw new ApiError(data?.error || 'Failed to connect to log stream', response.status, data)
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  try {
    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const parts = buffer.split('\n\n')
      buffer = parts.pop() || ''

      for (const part of parts) {
        if (!part.trim()) continue

        let eventType = 'message'
        let data = ''

        for (const line of part.split('\n')) {
          if (line.startsWith('event:')) {
            eventType = line.slice(6).trim()
          } else if (line.startsWith('data:')) {
            data += line.slice(5).trim()
          }
        }

        onEvent?.({ type: eventType, data })
      }
    }
  } catch (err) {
    if (err.name !== 'AbortError') {
      onError?.(err)
      throw err
    }
  }
}
