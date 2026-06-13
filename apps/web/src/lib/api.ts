import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? '/api',
  withCredentials: true,
})

// Token storage — in-memory only; populated by auth context (T1.7)
let _accessToken: string | null = null

export function setAccessToken(token: string | null) {
  _accessToken = token
}

export function getAccessToken(): string | null {
  return _accessToken
}

api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Track in-flight refresh to avoid concurrent duplicate calls
let _refreshPromise: Promise<string> | null = null

async function refreshAccessToken(): Promise<string> {
  if (_refreshPromise) return _refreshPromise

  // Use api instance so interceptors apply, but mark _retried=true to
  // prevent the 401 interceptor from recursing into another refresh.
  _refreshPromise = api
    .post<{ accessToken: string }>('/auth/refresh', {}, { _retried: true } as object)
    .then((res) => {
      const newToken = res.data.accessToken
      setAccessToken(newToken)
      return newToken
    })
    .finally(() => {
      _refreshPromise = null
    })

  return _refreshPromise
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config

    if (error.response?.status === 401 && !original._retried) {
      original._retried = true
      try {
        const newToken = await refreshAccessToken()
        original.headers.Authorization = `Bearer ${newToken}`
        return api(original)
      } catch {
        setAccessToken(null)
        return Promise.reject(error)
      }
    }

    return Promise.reject(error)
  },
)
