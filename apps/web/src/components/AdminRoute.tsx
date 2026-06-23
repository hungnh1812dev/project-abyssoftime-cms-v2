import { type ReactNode } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '@/context/AuthContext';

export function AdminRoute({ children }: { children: ReactNode }) {
  const { token, role, loading } = useAuth();

  if (loading) return null;
  if (!token) return <Navigate to="/login" replace />;
  if (role !== 'admin') return <Navigate to="/403" replace />;
  return <>{children}</>;
}
