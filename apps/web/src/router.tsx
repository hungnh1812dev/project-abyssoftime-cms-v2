import { lazy, Suspense } from "react";
import { Routes, Route, Navigate, useParams } from "react-router-dom";
import { LoginPage } from "@/pages/auth/LoginPage";
import { RegisterPage } from "@/pages/auth/RegisterPage";
import { AdminLayout } from "@/pages/admin/layout/AdminLayout";
import { ProtectedRoute } from "@/components/ProtectedRoute";
import { useContentTypes } from "@/hooks/useContentTypes";

const FormTestPanel = lazy(() =>
  import("@/pages/FormTestPanel").then((m) => ({ default: m.FormTestPanel })),
);

const SingleTypePage = lazy(() =>
  import("@/pages/admin/panels/single-type/SingleTypePage").then((m) => ({
    default: m.SingleTypePage,
  })),
);

const CollectionTypePage = lazy(() =>
  import("@/pages/admin/panels/collection-type/layout/CollectionTypePage").then((m) => ({
    default: m.CollectionTypePage,
  })),
);

const CollectionDetailPage = lazy(() =>
  import("@/pages/admin/panels/collection-type/CollectionDetailPage").then((m) => ({
    default: m.CollectionDetailPage,
  })),
);

const AboutPagePanel = lazy(() =>
  import("@/pages/admin/panels/AboutPagePanel").then((m) => ({
    default: m.AboutPagePanel,
  })),
);

const BlogPostDetailPanel = lazy(() =>
  import("@/pages/admin/panels/BlogPostDetailPanel").then((m) => ({
    default: m.BlogPostDetailPanel,
  })),
);

function PanelFallback() {
  return <div className="text-muted-foreground p-4">Loading…</div>;
}

function AboutPageWrapper() {
  const { data: contentTypes = [], isLoading } = useContentTypes();
  const ct = contentTypes.find((c) => c.Slug === "about-page");
  if (isLoading) return <PanelFallback />;
  if (!ct)
    return <p className="text-muted-foreground">Content type not found.</p>;
  return <AboutPagePanel contentType={ct} />;
}

function BlogPostDetailWrapper() {
  const { id } = useParams<{ id: string }>();
  const { data: contentTypes = [], isLoading } = useContentTypes();
  const ct = contentTypes.find((c) => c.Slug === "blog-posts");
  if (isLoading) return <PanelFallback />;
  if (!ct || !id)
    return <p className="text-muted-foreground">Not found.</p>;
  return <BlogPostDetailPanel contentType={ct} documentId={id} />;
}

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
        <Route
          index
          element={
            <p className="text-muted-foreground">
              Select a panel from the sidebar.
            </p>
          }
        />
        {/* Custom panels — static segments match before the generic :slug routes */}
        <Route
          path="content-type/single-type/about-page"
          element={
            <Suspense fallback={<PanelFallback />}>
              <AboutPageWrapper />
            </Suspense>
          }
        />
        <Route
          path="content-type/collection-type/blog-posts/:id"
          element={
            <Suspense fallback={<PanelFallback />}>
              <BlogPostDetailWrapper />
            </Suspense>
          }
        />

        {/* Generic panels — catch all remaining slugs */}
        <Route
          path="content-type/single-type/:slug"
          element={
            <Suspense fallback={<PanelFallback />}>
              <SingleTypePage />
            </Suspense>
          }
        />
        <Route
          path="content-type/collection-type/:slug"
          element={
            <Suspense fallback={<PanelFallback />}>
              <CollectionTypePage />
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
      </Route>
      <Route
        path="/form-test"
        element={
          <Suspense fallback={<PanelFallback />}>
            <FormTestPanel />
          </Suspense>
        }
      />
      <Route path="*" element={<Navigate to="/admin" replace />} />
    </Routes>
  );
}
