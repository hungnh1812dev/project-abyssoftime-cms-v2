import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react';
import { ConnectionOverlay } from '@/components/ConnectionOverlay';

const PING_INTERVAL_HEALTHY = 14 * 60 * 1000;
const PING_INTERVAL_UNHEALTHY = 10 * 1000;
const PING_TIMEOUT = 5 * 1000;

interface HealthContextValue {
  isApiHealthy: boolean;
}

const HealthContext = createContext<HealthContextValue | null>(null);

export function HealthProvider({ children }: { children: ReactNode }) {
  const [isApiHealthy, setIsApiHealthy] = useState(true);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const mountedRef = useRef(true);
  const pingRef = useRef<() => void>(() => {});

  function clearTimer() {
    if (timerRef.current !== null) {
      clearTimeout(timerRef.current);
      timerRef.current = null;
    }
  }

  function pingHealth() {
    clearTimer();

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), PING_TIMEOUT);
    const baseUrl = import.meta.env.VITE_API_URL || '';

    fetch(`${baseUrl}/health`, { signal: controller.signal })
      .then((response) => {
        clearTimeout(timeoutId);
        if (!mountedRef.current) return;

        if (response.ok) {
          setIsApiHealthy(true);
          timerRef.current = setTimeout(() => pingRef.current(), PING_INTERVAL_HEALTHY);
        } else {
          setIsApiHealthy(false);
          timerRef.current = setTimeout(() => pingRef.current(), PING_INTERVAL_UNHEALTHY);
        }
      })
      .catch(() => {
        clearTimeout(timeoutId);
        if (!mountedRef.current) return;

        setIsApiHealthy(false);
        timerRef.current = setTimeout(() => pingRef.current(), PING_INTERVAL_UNHEALTHY);
      });
  }

  useEffect(() => {
    pingRef.current = pingHealth;
  });

  useEffect(() => {
    mountedRef.current = true;
    pingRef.current();

    return () => {
      mountedRef.current = false;
      clearTimer();
    };
  }, []);

  useEffect(() => {
    function handleVisibilityChange() {
      if (document.visibilityState === 'hidden') {
        clearTimer();
      } else {
        pingRef.current();
      }
    }

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, []);

  return (
    <HealthContext.Provider value={{ isApiHealthy }}>
      {children}
      <ConnectionOverlay visible={!isApiHealthy} />
    </HealthContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useHealthStatus(): HealthContextValue {
  const context = useContext(HealthContext);
  if (!context) throw new Error('useHealthStatus must be used inside HealthProvider');
  return context;
}
