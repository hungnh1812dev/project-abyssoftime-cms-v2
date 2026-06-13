import { Routes, Route, Navigate } from 'react-router-dom'
import { LoginPage } from '@/pages/auth/LoginPage'
import { RegisterPage } from '@/pages/auth/RegisterPage'
import { AdminLayout } from '@/pages/admin/layout/AdminLayout'
import { ProtectedRoute } from '@/components/ProtectedRoute'
import { FormTestPanel } from '@/pages/FormTestPanel'

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        path="/403"
        element={
          <div className="flex min-h-screen items-center justify-center">
            <p className="text-muted-foreground">403 — Forbidden</p>
          </div>
        }
      />
      <Route
        path="/admin"
        element={
          <ProtectedRoute>
            <AdminLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<p className="text-muted-foreground">Select a panel from the sidebar.</p>} />
      </Route>
      <Route path="/form-test" element={<FormTestPanel />} />
      <Route path="*" element={<Navigate to="/admin" replace />} />
    </Routes>
  )
}
