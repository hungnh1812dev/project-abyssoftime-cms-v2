import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? '/api',
  withCredentials: true,
})

// Request interceptor — attach access token (populated by auth context later)
api.interceptors.request.use((config) => {
  const token = getAccessToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Response interceptor — auto-refresh on 401 (wired up fully in T1.5)
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    return Promise.reject(error)
  }
)

// Token storage — in-memory only, replaced by auth context in T1.5
let _accessToken: string | null = null

export function setAccessToken(token: string | null) {
  _accessToken = token
}

export function getAccessToken(): string | null {
  return _accessToken
}
