import type { Terminal } from '@xterm/xterm';

export const terminalInstances = new Map<string, { terminal: Terminal; ws: WebSocket }>();

export function destroyTerminalTab(tabId: string) {
  const instance = terminalInstances.get(tabId);
  if (instance) {
    instance.ws.close();
    instance.terminal.dispose();
    terminalInstances.delete(tabId);
  }
}
