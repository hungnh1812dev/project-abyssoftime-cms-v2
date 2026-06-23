import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { screen, waitFor, act } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';
import { api, getAccessToken, setAccessToken } from '@/lib/api';
import { renderWithProviders } from '@/test-utils';
import { AuthProvider, useAuth } from '@/context/AuthContext';

/** Builds a structurally valid JWT whose payload can be decoded client-side. */
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
  vi.clearAllMocks();
});

// Helper component to inspect auth context values
function AuthDisplay() {
  const { token, role, userId, loading } = useAuth();
  if (loading) return <span data-testid="loading">loading</span>;
  return (
    <div>
      <span data-testid="token">{token ?? 'none'}</span>
      <span data-testid="role">{role ?? 'none'}</span>
      <span data-testid="userId">{userId ?? 'none'}</span>
    </div>
  );
}

describe('AuthProvider', () => {
  it('starts in loading state then resolves when refresh fails (unauthenticated)', async () => {
    mock.onPost('/auth/refresh').reply(401);

    renderWithProviders(
      <AuthProvider>
        <AuthDisplay />
      </AuthProvider>,
    );

    expect(screen.getByTestId('loading')).toBeInTheDocument();
    await waitFor(() => expect(screen.getByTestId('token')).toHaveTextContent('none'));
    expect(screen.getByTestId('role')).toHaveTextContent('none');
  });

  it('hydrates token + role from refresh cookie on mount', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN });

    renderWithProviders(
      <AuthProvider>
        <AuthDisplay />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByTestId('role')).toHaveTextContent('admin'));
    expect(screen.getByTestId('userId')).toHaveTextContent('u1');
    expect(getAccessToken()).toBe(ADMIN_TOKEN);
  });

  it('login() updates context and sets api token', async () => {
    mock.onPost('/auth/refresh').reply(401);

    function LoginTrigger() {
      const { login, token } = useAuth();
      return (
        <div>
          <span data-testid="token">{token ?? 'none'}</span>
          <button onClick={() => login(ADMIN_TOKEN)}>login</button>
        </div>
      );
    }

    const { getByRole } = renderWithProviders(
      <AuthProvider>
        <LoginTrigger />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.queryByTestId('loading')).not.toBeInTheDocument());
    await act(async () => getByRole('button', { name: 'login' }).click());

    expect(screen.getByTestId('token')).toHaveTextContent(ADMIN_TOKEN);
    expect(getAccessToken()).toBe(ADMIN_TOKEN);
  });

  it('logout() clears context and calls POST /auth/logout', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN });
    let logoutCalled = false;
    mock.onPost('/auth/logout').reply(() => {
      logoutCalled = true;
      return [200, {}];
    });

    function LogoutTrigger() {
      const { logout, token } = useAuth();
      return (
        <div>
          <span data-testid="token">{token ?? 'none'}</span>
          <button onClick={() => logout()}>logout</button>
        </div>
      );
    }

    const { getByRole } = renderWithProviders(
      <AuthProvider>
        <LogoutTrigger />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByTestId('token')).not.toHaveTextContent('none'));
    await act(async () => getByRole('button', { name: 'logout' }).click());

    await waitFor(() => expect(screen.getByTestId('token')).toHaveTextContent('none'));
    expect(logoutCalled).toBe(true);
    expect(getAccessToken()).toBeNull();
  });
});

describe('useAuth (context access)', () => {
  it('provides correct role from token', async () => {
    mock.onPost('/auth/refresh').reply(200, { accessToken: GUEST_TOKEN });

    renderWithProviders(
      <AuthProvider>
        <AuthDisplay />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByTestId('role')).toHaveTextContent('guest'));
  });
});
