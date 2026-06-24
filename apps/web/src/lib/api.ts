import axios from 'axios';

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '',
  withCredentials: true,
});

// Access token — in-memory only; populated by auth context
let _accessToken: string | null = null;

export function setAccessToken(token: string | null) {
  _accessToken = token;
}

export function getAccessToken(): string | null {
  return _accessToken;
}

// Refresh token — localStorage (rememberMe) or sessionStorage (session-only)
const REFRESH_KEY = 'refresh_token';

export function storeRefreshToken(token: string, remember?: boolean) {
  const useLocal = remember ?? localStorage.getItem(REFRESH_KEY) !== null;
  const [target, other] = useLocal ? [localStorage, sessionStorage] : [sessionStorage, localStorage];
  target.setItem(REFRESH_KEY, token);
  other.removeItem(REFRESH_KEY);
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_KEY) || sessionStorage.getItem(REFRESH_KEY);
}

export function clearRefreshToken() {
  localStorage.removeItem(REFRESH_KEY);
  sessionStorage.removeItem(REFRESH_KEY);
}

let _onSessionExpired: (() => void) | null = null;

export function onSessionExpired(callback: (() => void) | null) {
  _onSessionExpired = callback;
}

api.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Track in-flight refresh to avoid concurrent duplicate calls
let _refreshPromise: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  if (_refreshPromise) return _refreshPromise;

  const stored = getRefreshToken();

  _refreshPromise = api
    .post<{ accessToken: string; refreshToken: string }>(
      '/auth/refresh',
      { refreshToken: stored },
      { _retried: true } as object,
    )
    .then((res) => {
      const { accessToken, refreshToken } = res.data;
      setAccessToken(accessToken);
      if (refreshToken) storeRefreshToken(refreshToken);
      return accessToken;
    })
    .finally(() => {
      _refreshPromise = null;
    });

  return _refreshPromise;
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const original = error.config;

    if (error.response?.status === 401 && !original._retried) {
      original._retried = true;
      try {
        const newToken = await refreshAccessToken();
        original.headers.Authorization = `Bearer ${newToken}`;
        return api(original);
      } catch {
        setAccessToken(null);
        clearRefreshToken();
        _onSessionExpired?.();
        return Promise.reject(error);
      }
    }

    return Promise.reject(error);
  },
);
