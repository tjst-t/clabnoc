import { useEffect, useRef, useCallback } from 'react';
import type { ApiEvent } from '../types/topology';
import { createEventsWebSocket } from '../lib/api';

export function useWebSocket(
  project: string | null,
  onEvent: (event: ApiEvent) => void
) {
  const wsRef = useRef<WebSocket | null>(null);
  const onEventRef = useRef(onEvent);
  onEventRef.current = onEvent;

  const connect = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = createEventsWebSocket(project ?? undefined);
    wsRef.current = ws;

    ws.onmessage = (e) => {
      try {
        const event = JSON.parse(e.data) as ApiEvent;
        onEventRef.current(event);
      } catch {
        // ignore non-JSON messages
      }
    };

    ws.onclose = () => {
      // Reconnect after delay
      setTimeout(() => {
        if (wsRef.current === ws) {
          connect();
        }
      }, 3000);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [project]);

  useEffect(() => {
    connect();
    return () => {
      const ws = wsRef.current;
      wsRef.current = null;
      ws?.close();
    };
  }, [connect]);
}
