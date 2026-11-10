const API_BASE_URL = import.meta.env.VITE_API_URL ?? ''

export async function apiGet<T>(path: string): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`)

  const result = await response.json().catch(() => null)

  if (!response.ok) {
    const message =
      result?.error?.message ?? `Request failed (${response.status})`

    throw new Error(message)
  }

  return result as T
}

export async function apiPost<T>(
  path: string,
  body: unknown,
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })

  const result = await response.json().catch(() => null)

  if (!response.ok) {
    const message =
      result?.error?.message ?? `Request failed (${response.status})`

    throw new Error(message)
  }

  return result as T
}

export async function apiPut<T>(
  path: string,
  body: unknown,
): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  const result = await response.json().catch(() => null)

  if (!response.ok) {
    const message =
      result?.error?.message ?? `Request failed (${response.status})`
    throw new Error(message)
  }

  return result as T
}