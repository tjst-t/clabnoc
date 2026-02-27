import { useState, useEffect, useRef, useCallback } from 'react';
import type { ContainerStats } from '../types/topology';
import { createStatsWebSocket } from '../lib/api';

/** Validate that a parsed object has the required ContainerStats fields. */
function isContainerStats(v: unknown): v is ContainerStats {
  if (typeof v !== 'object' || v === null) return false;
  const obj = v as Record<string, unknown>;
  return (
    typeof obj.cpu_percent === 'number' &&
    typeof obj.memory_bytes === 'number' &&
    typeof obj.memory_limit === 'number'
  );
}

export function useContainerStats(project: string | null): Map<string, ContainerStats> {
  const [stats, setStats] = useState<Map<string, ContainerStats>>(new Map());
  const wsRef = useRef<WebSocket | null>(null);

  const connect = useCallback(() => {
    if (!project) return;

    if (wsRef.current) {
      wsRef.current.close();
    }

    const ws = createStatsWebSocket(project);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'stats' && msg.stats) {
          const newStats = new Map<string, ContainerStats>();
          for (const [name, s] of Object.entries(msg.stats)) {
            if (isContainerStats(s)) {
              newStats.set(name, s);
            }
          }
          setStats(newStats);
        }
      } catch {
        // ignore parse errors
      }
    };

    ws.onclose = () => {
      // Reconnect after delay (only if this ws is still the active one)
      setTimeout(() => {
        if (wsRef.current === ws) {
          connect();
        }
      }, 5000);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [project]);

  useEffect(() => {
    if (!project) {
      setStats(new Map());
      return;
    }

    connect();

    return () => {
      const ws = wsRef.current;
      wsRef.current = null;
      ws?.close();
    };
  }, [connect, project]);

  return stats;
}
