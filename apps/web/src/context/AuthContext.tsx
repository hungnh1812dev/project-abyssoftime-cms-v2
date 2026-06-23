import { createContext, useCallback, useContext, useEffect, useRef, useState, type ReactNode } from 'react';
import { api, setAccessToken, onSessionExpired } from '@/lib/api';

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
  login: (token: string) => void;
  logout: () => void;
}

const LOGGED_OUT_STATE: AuthState = { token: null, role: null, userId: null, loading: false };

const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    token: null,
    role: null,
    userId: null,
    loading: true,
  });

  const mountedRef = useRef(true);

  useEffect(() => {
    mountedRef.current = true;

    onSessionExpired(() => {
      if (mountedRef.current) {
        setAccessToken(null);
        setState(LOGGED_OUT_STATE);
      }
    });

    api
      .post<{ accessToken: string }>(
        '/auth/refresh',
        {},
        {
          _retried: true,
          headers: { 'Cache-Control': 'no-cache, no-store' },
        } as object,
      )
      .then((response) => {
        if (!mountedRef.current) return;
        const token = response.data.accessToken;
        const { userId, role } = decodeToken(token);
        setAccessToken(token);
        setState({ token, role, userId, loading: false });
      })
      .catch(() => {
        if (mountedRef.current) {
          setState(LOGGED_OUT_STATE);
        }
      });

    return () => {
      mountedRef.current = false;
      onSessionExpired(null);
    };
  }, []);

  function login(token: string) {
    const { userId, role } = decodeToken(token);
    setAccessToken(token);
    setState({ token, role, userId, loading: false });
  }

  const logout = useCallback(async () => {
    setAccessToken(null);
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
