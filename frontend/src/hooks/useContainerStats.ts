import { useState, useEffect, useRef } from 'react';
import type { ContainerStats } from '../types/topology';
import { createStatsWebSocket } from '../lib/api';

export function useContainerStats(project: string | null): Map<string, ContainerStats> {
  const [stats, setStats] = useState<Map<string, ContainerStats>>(new Map());
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!project) {
      setStats(new Map());
      return;
    }

    const ws = createStatsWebSocket(project);
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        if (msg.type === 'stats' && msg.stats) {
          const newStats = new Map<string, ContainerStats>();
          for (const [name, s] of Object.entries(msg.stats)) {
            newStats.set(name, s as ContainerStats);
          }
          setStats(newStats);
        }
      } catch {
        // ignore parse errors
      }
    };

    ws.onerror = () => {
      // Silently handle stats connection errors
    };

    return () => {
      ws.close();
      wsRef.current = null;
    };
  }, [project]);

  return stats;
}
