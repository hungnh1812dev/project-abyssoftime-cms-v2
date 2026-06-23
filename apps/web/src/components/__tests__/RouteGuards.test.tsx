import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';
import { api, setAccessToken } from '@/lib/api';
import { renderWithProviders } from '@/test-utils';
import { AuthProvider } from '@/context/AuthContext';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { AdminRoute } from '@/components/AdminRoute';
import { Routes, Route } from 'react-router-dom';

function makeToken(payload: Record<string, unknown>) {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.fakesig`;
}

const ADMIN_TOKEN = makeToken({ userId: 'u1', role: 'admin', exp: 9999999999 });
const GUEST_TOKEN = makeToken({ userId: 'u2', role: 'guest', exp: 9999999999 });

let mock: MockAdapter;

beforeEach(() => {
  mock = new MockAdapter(api);
  setAccessToken(null);
});

afterEach(() => {
  mock.restore();
});

function wrap(ui: React.ReactNode, initialEntries = ['/']) {
  return renderWithProviders(
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<p>Login page</p>} />
        <Route path="/403" element={<p>403 Forbidden</p>} />
        {ui}
      </Routes>
    </AuthProvider>,
    { initialEntries },
  );
}

describe('ProtectedRoute', () => {
  it('redirects to /login when not authenticated', async () => {
    mock.onPost('/auth/refresh').reply(401);

    wrap(
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <p>Secret content</p>
          </ProtectedRoute>
        }
      />,
    );

    await waitFor(() => expect(screen.getByText('Login page')).toBeInTheDocument());
    expect(screen.queryByText('Secret content')).not.toBeInTheDocument();
  });

  it('renders children when authenticated', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN });

    wrap(
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <p>Secret content</p>
          </ProtectedRoute>
        }
      />,
    );

    await waitFor(() => expect(screen.getByText('Secret content')).toBeInTheDocument());
  });
});

describe('AdminRoute', () => {
  it('redirects to /login when not authenticated', async () => {
    mock.onPost('/auth/refresh').reply(401);

    wrap(
      <Route
        path="/"
        element={
          <AdminRoute>
            <p>Admin content</p>
          </AdminRoute>
        }
      />,
    );

    await waitFor(() => expect(screen.getByText('Login page')).toBeInTheDocument());
  });

  it('shows 403 when role is not admin', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: GUEST_TOKEN });

    wrap(
      <Route
        path="/"
        element={
          <AdminRoute>
            <p>Admin content</p>
          </AdminRoute>
        }
      />,
    );

    await waitFor(() => expect(screen.getByText('403 Forbidden')).toBeInTheDocument());
    expect(screen.queryByText('Admin content')).not.toBeInTheDocument();
  });

  it('renders children when role is admin', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN });

    wrap(
      <Route
        path="/"
        element={
          <AdminRoute>
            <p>Admin content</p>
          </AdminRoute>
        }
      />,
    );

    await waitFor(() => expect(screen.getByText('Admin content')).toBeInTheDocument());
  });
});
