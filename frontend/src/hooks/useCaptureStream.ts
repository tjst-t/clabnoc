import { useState, useEffect, useRef, useCallback } from 'react';
import type { PacketInfo } from '../types/topology';
import { createCaptureStreamWebSocket } from '../lib/api';

const MAX_PACKETS = 5000;

interface UseCaptureStreamResult {
  packets: PacketInfo[];
  connected: boolean;
  paused: boolean;
  pause: () => void;
  resume: () => void;
  clear: () => void;
}

export function useCaptureStream(
  project: string | null,
  linkId: string | null,
  bpfFilter?: string
): UseCaptureStreamResult {
  const [packets, setPackets] = useState<PacketInfo[]>([]);
  const [connected, setConnected] = useState(false);
  const [paused, setPaused] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    if (!project || !linkId) {
      setConnected(false);
      return;
    }

    const ws = createCaptureStreamWebSocket(project, linkId, bpfFilter);
    wsRef.current = ws;

    ws.onopen = () => setConnected(true);

    ws.onmessage = (e) => {
      try {
        const pkt = JSON.parse(e.data) as PacketInfo;
        if (typeof pkt.no !== 'number') return;
        setPackets((prev) => {
          const next = [...prev, pkt];
          if (next.length > MAX_PACKETS) {
            return next.slice(next.length - MAX_PACKETS);
          }
          return next;
        });
      } catch {
        // ignore non-JSON messages
      }
    };

    ws.onclose = () => {
      setConnected(false);
    };

    ws.onerror = () => {
      ws.close();
    };

    return () => {
      wsRef.current = null;
      ws.close();
    };
  }, [project, linkId, bpfFilter]);

  const pause = useCallback(() => {
    setPaused(true);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'pause' }));
    }
  }, []);

  const resume = useCallback(() => {
    setPaused(false);
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type: 'resume' }));
    }
  }, []);

  const clear = useCallback(() => {
    setPackets([]);
  }, []);

  return { packets, connected, paused, pause, resume, clear };
}
