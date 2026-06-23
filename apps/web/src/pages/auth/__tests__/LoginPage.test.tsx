import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import MockAdapter from 'axios-mock-adapter';
import { api, getAccessToken, setAccessToken, clearRefreshToken } from '@/lib/api';
import { renderWithProviders } from '@/test-utils';
import { AuthProvider } from '@/context/AuthContext';
import { LoginPage } from '@/pages/auth/LoginPage';

function makeToken(payload: Record<string, unknown>) {
  const header = btoa(JSON.stringify({ alg: 'HS256', typ: 'JWT' }));
  const body = btoa(JSON.stringify(payload));
  return `${header}.${body}.fakesig`;
}

const ADMIN_TOKEN = makeToken({ userId: 'u1', role: 'admin', exp: 9999999999 });

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

function renderLogin() {
  return renderWithProviders(
    <AuthProvider>
      <LoginPage />
    </AuthProvider>,
  );
}

describe('LoginPage', () => {
  it('renders email and password fields with a submit button', async () => {
    renderLogin();
    await waitFor(() => expect(screen.getByLabelText(/email/i)).toBeInTheDocument());
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });

  it('shows validation error for invalid email', async () => {
    const user = userEvent.setup();
    renderLogin();

    await waitFor(() => expect(screen.getByLabelText(/email/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/email/i), 'not-an-email');
    await user.type(screen.getByLabelText(/password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText(/valid email/i)).toBeInTheDocument();
    });
  });

  it('shows validation error for password shorter than 8 characters', async () => {
    const user = userEvent.setup();
    renderLogin();

    await waitFor(() => expect(screen.getByLabelText(/email/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/email/i), 'user@example.com');
    await user.type(screen.getByLabelText(/password/i), 'short');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByText(/at least 8/i)).toBeInTheDocument();
    });
  });

  it('calls POST /auth/login and stores access + refresh tokens on success', async () => {
    const user = userEvent.setup();
    mock.onPost('/auth/login').reply(200, { accessToken: ADMIN_TOKEN, refreshToken: 'login-refresh-tok' });
    renderLogin();

    await waitFor(() => expect(screen.getByLabelText(/email/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/email/i), 'user@example.com');
    await user.type(screen.getByLabelText(/password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(getAccessToken()).toBe(ADMIN_TOKEN);
    });
    expect(sessionStorage.getItem('refresh_token')).toBe('login-refresh-tok');
  });

  it('shows error message when login fails', async () => {
    const user = userEvent.setup();
    mock.onPost('/auth/login').reply(401, { message: 'Invalid credentials' });
    renderLogin();

    await waitFor(() => expect(screen.getByLabelText(/email/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/email/i), 'user@example.com');
    await user.type(screen.getByLabelText(/password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /sign in/i }));

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeInTheDocument();
    });
  });
});
