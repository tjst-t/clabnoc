import type { Terminal } from '@xterm/xterm';
import type { FitAddon } from '@xterm/addon-fit';

export const terminalInstances = new Map<string, { terminal: Terminal; ws: WebSocket; fitAddon: FitAddon }>();

export function destroyTerminalTab(tabId: string) {
  const instance = terminalInstances.get(tabId);
  if (instance) {
    instance.ws.close();
    instance.terminal.dispose();
    terminalInstances.delete(tabId);
  }
}
