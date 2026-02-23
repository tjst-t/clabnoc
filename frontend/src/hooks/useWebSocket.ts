import { useEffect, useRef, useCallback } from 'react';

interface UseWebSocketOptions {
  url: string | null;
  onMessage: (data: unknown) => void;
  onError?: (error: Event) => void;
  enabled?: boolean;
}

export function useWebSocket({ url, onMessage, onError, enabled = true }: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  const connect = useCallback(() => {
    if (!url || !enabled) return;
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onmessage = (evt) => {
      try {
        const data = JSON.parse(evt.data as string);
        onMessageRef.current(data);
      } catch {
        // ignore non-JSON (e.g., keepalive pings)
      }
    };

    ws.onerror = (evt) => onError?.(evt);

    ws.onclose = () => {
      // Reconnect after 3s
      if (enabled) {
        setTimeout(connect, 3000);
      }
    };
  }, [url, enabled, onError]);

  useEffect(() => {
    connect();
    return () => {
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [connect]);
}
