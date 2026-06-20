import { type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useAuth } from '@/context/AuthContext'
import { roleLevel } from '@/lib/roles'
import { api } from '@/lib/api'

export function ProtectedRoute({
  children,
  minRole,
}: {
  children: ReactNode
  minRole?: string
}) {
  const { token, role, loading } = useAuth()

  const { data: setupData, isLoading: setupLoading } = useQuery({
    queryKey: ['auth-setup'],
    queryFn: () =>
      api.get<{ adminExists: boolean }>('/auth/setup').then((r) => r.data),
    enabled: !loading && !token,
    staleTime: 30_000,
  })

  if (loading || setupLoading) return null
  if (!token) {
    const hasAdmin = setupData?.adminExists ?? true
    return <Navigate to={hasAdmin ? '/login' : '/register'} replace />
  }
  if (minRole && roleLevel(role) < roleLevel(minRole)) {
    return <Navigate to="/403" replace />
  }
  return <>{children}</>
}
