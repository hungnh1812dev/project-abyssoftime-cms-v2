import { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { LoginPage } from '@/pages/auth/LoginPage';
import { RegisterPage } from '@/pages/auth/RegisterPage';
import { AdminLayout } from '@/pages/admin/layout/AdminLayout';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { AdminPage } from './pages/admin/AdminPage';

const ContentTypePage = lazy(() =>
  import('@/pages/admin/panels/ContentTypePage').then((module) => ({
    default: module.ContentTypePage,
  })),
);

const CollectionDetailPage = lazy(() =>
  import('@/pages/admin/panels/collection-type/CollectionDetailPage').then((module) => ({
    default: module.CollectionDetailPage,
  })),
);

const MediaLibraryPage = lazy(() =>
  import('@/pages/admin/settings/MediaLibraryPage').then((module) => ({
    default: module.MediaLibraryPage,
  })),
);

const UsersPage = lazy(() =>
  import('@/pages/admin/settings/UsersPage').then((module) => ({
    default: module.UsersPage,
  })),
);

const AccessTokensPage = lazy(() =>
  import('@/pages/admin/settings/AccessTokensPage').then((module) => ({
    default: module.AccessTokensPage,
  })),
);

const RolesPage = lazy(() =>
  import('@/pages/admin/settings/RolesPage').then((module) => ({
    default: module.RolesPage,
  })),
);

const InternationalizePage = lazy(() =>
  import('@/pages/admin/settings/InternationalizePage').then((module) => ({
    default: module.InternationalizePage,
  })),
);

const InviteAcceptPage = lazy(() =>
  import('@/pages/auth/InviteAcceptPage').then((module) => ({
    default: module.InviteAcceptPage,
  })),
);

function PanelFallback() {
  return <div className="text-muted-foreground p-4">Loading…</div>;
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        path="/invite/:token"
        element={
          <Suspense fallback={<PanelFallback />}>
            <InviteAcceptPage />
          </Suspense>
        }
      />
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
        }>
        <Route index element={<AdminPage />} />

        <Route
          path="content-type/single-type/:slug"
          element={
            <Suspense fallback={<PanelFallback />}>
              <ContentTypePage />
            </Suspense>
          }
        />
        <Route
          path="content-type/collection-type/:slug"
          element={
            <Suspense fallback={<PanelFallback />}>
              <ContentTypePage />
            </Suspense>
          }
        />
        <Route
          path="content-type/collection-type/:slug/new"
          element={
            <Suspense fallback={<PanelFallback />}>
              <CollectionDetailPage />
            </Suspense>
          }
        />
        <Route
          path="content-type/collection-type/:slug/:id"
          element={
            <Suspense fallback={<PanelFallback />}>
              <CollectionDetailPage />
            </Suspense>
          }
        />
        <Route
          path="settings/media"
          element={
            <Suspense fallback={<PanelFallback />}>
              <MediaLibraryPage />
            </Suspense>
          }
        />
        <Route
          path="settings/users"
          element={
            <ProtectedRoute minRole="admin">
              <Suspense fallback={<PanelFallback />}>
                <UsersPage />
              </Suspense>
            </ProtectedRoute>
          }
        />
        <Route
          path="settings/access-tokens"
          element={
            <ProtectedRoute minRole="super_admin">
              <Suspense fallback={<PanelFallback />}>
                <AccessTokensPage />
              </Suspense>
            </ProtectedRoute>
          }
        />
        <Route
          path="settings/roles"
          element={
            <ProtectedRoute minRole="super_admin">
              <Suspense fallback={<PanelFallback />}>
                <RolesPage />
              </Suspense>
            </ProtectedRoute>
          }
        />
        <Route
          path="settings/internationalize"
          element={
            <ProtectedRoute minRole="super_admin">
              <Suspense fallback={<PanelFallback />}>
                <InternationalizePage />
              </Suspense>
            </ProtectedRoute>
          }
        />
      </Route>
      <Route path="*" element={<Navigate to="/admin" replace />} />
    </Routes>
  );
}
