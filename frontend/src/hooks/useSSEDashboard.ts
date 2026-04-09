import { useState, useEffect, useRef } from 'react';
import { cleanDashboardData } from '../hooks/dashboardUtils';
import type { DashboardResponse } from '../types';

const SSE_URL = '/api/stream/dashboard';

/**
 * useSSEDashboard – connects to the backend SSE stream.
 * No more setInterval / 10-second polling.
 * The server pushes data each time it refreshes the cache.
 */
export function useSSEDashboard() {
  const [data, setData] = useState<DashboardResponse | undefined>(undefined);
  const [isLoading, setIsLoading] = useState(true);
  const [isError, setIsError] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const esRef = useRef<EventSource | null>(null);
  const reconnectTimeout = useRef<ReturnType<typeof setTimeout> | null>(null);

  const connect = () => {
    if (esRef.current) {
      esRef.current.close();
    }

    const es = new EventSource(SSE_URL);
    esRef.current = es;

    es.onopen = () => {
      setIsError(false);
      setError(null);
    };

    es.onmessage = (event) => {
      try {
        const raw = JSON.parse(event.data) as DashboardResponse;
        setData(cleanDashboardData(raw));
        setIsLoading(false);
      } catch (e) {
        console.error('[SSE] Failed to parse message', e);
      }
    };

    es.onerror = () => {
      es.close();
      esRef.current = null;
      setIsError(true);
      setError(new Error('SSE connection lost – reconnecting in 5s...'));
      // Auto-reconnect after 5 seconds
      reconnectTimeout.current = setTimeout(connect, 5000);
    };
  };

  useEffect(() => {
    connect();
    return () => {
      esRef.current?.close();
      if (reconnectTimeout.current) clearTimeout(reconnectTimeout.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return { data, isLoading, isError, error };
}
