import { useEffect, useRef } from 'react';
import { Terminal } from '@xterm/xterm';
import type { TerminalTab as TerminalTabType } from '../types/topology';
import { createExecWebSocket, createSSHWebSocket } from '../lib/api';

interface Props {
  project: string;
  tab: TerminalTabType;
  active: boolean;
}

// Persistent storage for terminal instances and WebSocket connections
const terminalInstances = new Map<string, { terminal: Terminal; ws: WebSocket }>();

export function TerminalTab({ project, tab, active }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const initializedRef = useRef(false);

  useEffect(() => {
    if (!containerRef.current || initializedRef.current) return;
    initializedRef.current = true;

    // Check if we already have an instance for this tab
    let instance = terminalInstances.get(tab.id);
    if (instance) {
      containerRef.current.appendChild(instance.terminal.element!);
      return;
    }

    const terminal = new Terminal({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: '"JetBrains Mono", "Fira Code", monospace',
      theme: {
        background: '#0a0e14',
        foreground: '#c5cdd8',
        cursor: '#00d4aa',
        selectionBackground: '#2a4060',
        black: '#0a0e14',
        red: '#ff3b5c',
        green: '#00e878',
        yellow: '#ffb020',
        blue: '#00c8ff',
        magenta: '#ff6b8a',
        cyan: '#00d4aa',
        white: '#c5cdd8',
        brightBlack: '#5a6a7e',
        brightRed: '#ff6b8a',
        brightGreen: '#00e878',
        brightYellow: '#ffb020',
        brightBlue: '#00c8ff',
        brightMagenta: '#ff8aaa',
        brightCyan: '#00d4aa',
        brightWhite: '#e8edf3',
      },
    });

    terminal.open(containerRef.current);

    const ws =
      tab.type === 'exec'
        ? createExecWebSocket(project, tab.node)
        : createSSHWebSocket(project, tab.node);

    ws.binaryType = 'arraybuffer';

    ws.onopen = () => {
      terminal.writeln(`\x1b[2m--- Connected (${tab.type}) ---\x1b[0m`);
    };

    ws.onmessage = (e) => {
      if (e.data instanceof ArrayBuffer) {
        terminal.write(new Uint8Array(e.data));
      } else {
        terminal.write(e.data);
      }
    };

    ws.onclose = () => {
      terminal.writeln('\r\n\x1b[2m--- Disconnected ---\x1b[0m');
    };

    ws.onerror = () => {
      terminal.writeln('\r\n\x1b[31m--- Connection error ---\x1b[0m');
    };

    terminal.onData((data) => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(data);
      }
    });

    terminalInstances.set(tab.id, { terminal, ws });

    return () => {
      // Don't destroy on unmount - preserve for project switching
    };
  }, [project, tab]);

  // Handle resize when becoming active
  useEffect(() => {
    if (!active) return;
    const instance = terminalInstances.get(tab.id);
    if (instance) {
      // Small delay to let container resize
      const timer = setTimeout(() => {
        instance.terminal.focus();
      }, 50);
      return () => clearTimeout(timer);
    }
  }, [active, tab.id]);

  return <div ref={containerRef} className="w-full h-full" />;
}

// Cleanup function for when tabs are closed
export function destroyTerminalTab(tabId: string) {
  const instance = terminalInstances.get(tabId);
  if (instance) {
    instance.ws.close();
    instance.terminal.dispose();
    terminalInstances.delete(tabId);
  }
}
