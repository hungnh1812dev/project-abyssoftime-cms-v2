import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { screen, waitFor, act } from '@testing-library/react';
import MockAdapter from 'axios-mock-adapter';
import { api, getAccessToken, setAccessToken, clearRefreshToken } from '@/lib/api';
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
const FAKE_REFRESH = 'fake-refresh-token';

let mock: MockAdapter;

beforeEach(() => {
  mock = new MockAdapter(api);
  setAccessToken(null);
  clearRefreshToken();
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
  it('immediately resolves as unauthenticated when no stored refresh token', async () => {
    renderWithProviders(
      <AuthProvider>
        <AuthDisplay />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByTestId('token')).toHaveTextContent('none'));
    expect(screen.getByTestId('role')).toHaveTextContent('none');
  });

  it('starts in loading state then resolves when refresh fails', async () => {
    localStorage.setItem('refresh_token', FAKE_REFRESH);
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

  it('hydrates token + role from stored refresh token on mount', async () => {
    localStorage.setItem('refresh_token', FAKE_REFRESH);
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN, refreshToken: 'new-refresh' });

    renderWithProviders(
      <AuthProvider>
        <AuthDisplay />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByTestId('role')).toHaveTextContent('admin'));
    expect(screen.getByTestId('userId')).toHaveTextContent('u1');
    expect(getAccessToken()).toBe(ADMIN_TOKEN);
  });

  it('login() updates context, sets api token, and stores refresh token', async () => {
    function LoginTrigger() {
      const { login, token } = useAuth();
      return (
        <div>
          <span data-testid="token">{token ?? 'none'}</span>
          <button onClick={() => login(ADMIN_TOKEN, FAKE_REFRESH, true)}>login</button>
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
    expect(localStorage.getItem('refresh_token')).toBe(FAKE_REFRESH);
  });

  it('logout() clears context, stored refresh token, and calls POST /auth/logout', async () => {
    localStorage.setItem('refresh_token', FAKE_REFRESH);
    mock.onPost('/auth/refresh').reply(200, { accessToken: ADMIN_TOKEN, refreshToken: 'new-refresh' });
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
    expect(localStorage.getItem('refresh_token')).toBeNull();
  });
});

describe('useAuth (context access)', () => {
  it('provides correct role from token', async () => {
    localStorage.setItem('refresh_token', FAKE_REFRESH);
    mock.onPost('/auth/refresh').reply(200, { accessToken: GUEST_TOKEN, refreshToken: 'new-refresh' });

    renderWithProviders(
      <AuthProvider>
        <AuthDisplay />
      </AuthProvider>,
    );

    await waitFor(() => expect(screen.getByTestId('role')).toHaveTextContent('guest'));
  });
});
