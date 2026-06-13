import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api, setAccessToken } from '@/lib/api'

interface JwtPayload {
  userId: string
  role: string
  exp: number
}

function decodeToken(token: string): JwtPayload {
  const payload = token.split('.')[1]
  return JSON.parse(atob(payload)) as JwtPayload
}

interface AuthState {
  token: string | null
  role: string | null
  userId: string | null
  loading: boolean
}

interface AuthContextValue extends AuthState {
  login: (token: string) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    token: null,
    role: null,
    userId: null,
    loading: true,
  })

  useEffect(() => {
    api
      .post<{ accessToken: string }>('/auth/refresh', {}, { _retried: true } as object)
      .then((res) => {
        const token = res.data.accessToken
        const { userId, role } = decodeToken(token)
        setAccessToken(token)
        setState({ token, role, userId, loading: false })
      })
      .catch(() => {
        setState({ token: null, role: null, userId: null, loading: false })
      })
  }, [])

  function login(token: string) {
    const { userId, role } = decodeToken(token)
    setAccessToken(token)
    setState({ token, role, userId, loading: false })
  }

  function logout() {
    api.post('/auth/logout').catch(() => {})
    setAccessToken(null)
    setState({ token: null, role: null, userId: null, loading: false })
  }

  return <AuthContext.Provider value={{ ...state, login, logout }}>{children}</AuthContext.Provider>
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider')
  return ctx
}
