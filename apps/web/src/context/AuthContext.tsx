import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react';
import { api, setAccessToken, onSessionExpired, storeRefreshToken, getRefreshToken, clearRefreshToken } from '@/lib/api';

interface JwtPayload {
  userId: string;
  role: string;
  exp: number;
}

function decodeToken(token: string): JwtPayload {
  const payload = token.split('.')[1];
  return JSON.parse(atob(payload)) as JwtPayload;
}

interface AuthState {
  token: string | null;
  role: string | null;
  userId: string | null;
  loading: boolean;
}

interface AuthContextValue extends AuthState {
  login: (accessToken: string, refreshToken: string, rememberMe: boolean) => void;
  logout: () => void;
}

const LOGGED_OUT_STATE: AuthState = { token: null, role: null, userId: null, loading: false };

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>(() => ({
    token: null,
    role: null,
    userId: null,
    loading: getRefreshToken() !== null,
  }));

  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;

    onSessionExpired(() => {
      if (mountedRef.current) {
        setAccessToken(null);
        clearRefreshToken();
        setState(LOGGED_OUT_STATE);
      }
    });

    const storedRefresh = getRefreshToken();
    if (!storedRefresh) {
      return () => {
        mountedRef.current = false;
        onSessionExpired(null);
      };
    }

    api
      .post<{ accessToken: string; refreshToken: string }>(
        '/auth/refresh',
        { refreshToken: storedRefresh },
        {
          _retried: true,
          headers: { 'Cache-Control': 'no-cache, no-store' },
        } as object,
      )
      .then((response) => {
        if (!mountedRef.current) return;
        const { accessToken, refreshToken } = response.data;
        setAccessToken(accessToken);
        if (refreshToken) storeRefreshToken(refreshToken);
        const { userId, role } = decodeToken(accessToken);
        setState({ token: accessToken, role, userId, loading: false });
      })
      .catch(() => {
        if (mountedRef.current) {
          clearRefreshToken();
          setState(LOGGED_OUT_STATE);
        }
      });

    return () => {
      mountedRef.current = false;
      onSessionExpired(null);
    };
  }, []);

  function login(accessToken: string, refreshToken: string, rememberMe: boolean) {
    const { userId, role } = decodeToken(accessToken);
    setAccessToken(accessToken);
    storeRefreshToken(refreshToken, rememberMe);
    setState({ token: accessToken, role, userId, loading: false });
  }

  const logout = useCallback(async () => {
    setAccessToken(null);
    clearRefreshToken();
    setState(LOGGED_OUT_STATE);
    try {
      await api.post('/auth/logout');
    } catch {
      // cookie cleared server-side on best-effort basis
    }
  }, []);

  return <AuthContext.Provider value={{ ...state, login, logout }}>{children}</AuthContext.Provider>;
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) throw new Error('useAuth must be used inside AuthProvider');
  return context;
}
