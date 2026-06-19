import { type ReactNode } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuth } from '@/context/AuthContext'
import { roleLevel } from '@/lib/roles'

export function ProtectedRoute({
  children,
  minRole,
}: {
  children: ReactNode
  minRole?: string
}) {
  const { token, role, loading } = useAuth()

  if (loading) return null
  if (!token) return <Navigate to="/login" replace />
  if (minRole && roleLevel(role) < roleLevel(minRole)) {
    return <Navigate to="/403" replace />
  }
  return <>{children}</>
}
