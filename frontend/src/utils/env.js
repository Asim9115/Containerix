/**
 * Convert env rows to the map shape expected by POST /build: { env: { KEY: "value" } }
 */
export function envRowsToMap(rows) {
  const env = {}
  for (const { key, value } of rows) {
    const trimmedKey = key.trim()
    if (!trimmedKey) continue
    env[trimmedKey] = value
  }
  return env
}

export function findDuplicateEnvKeys(rows) {
  const seen = new Set()
  const duplicates = new Set()

  for (const { key } of rows) {
    const trimmed = key.trim()
    if (!trimmed) continue
    if (seen.has(trimmed)) duplicates.add(trimmed)
    seen.add(trimmed)
  }

  return [...duplicates]
}

export function createEmptyEnvRow() {
  return { id: crypto.randomUUID(), key: '', value: '' }
}
