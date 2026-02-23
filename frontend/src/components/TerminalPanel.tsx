import React from 'react';
import { TerminalTab } from './TerminalTab';
import type { TerminalTab as TerminalTabType } from '../hooks/useTerminalTabs';
import type { Terminal } from '@xterm/xterm';

interface Props {
  tabs: TerminalTabType[];
  activeTabId: string | null;
  onTabSelect: (id: string) => void;
  onTabClose: (id: string) => void;
  onTabReady: (id: string, terminal: Terminal, socket: WebSocket) => void;
}

export function TerminalPanel({ tabs, activeTabId, onTabSelect, onTabClose, onTabReady }: Props) {
  if (tabs.length === 0) {
    return (
      <div className="terminal-panel empty">
        <p>Click a node and select a terminal access method</p>
      </div>
    );
  }

  return (
    <div className="terminal-panel">
      <div className="terminal-tabs">
        {tabs.map(tab => (
          <div
            key={tab.id}
            className={`terminal-tab ${activeTabId === tab.id ? 'active' : ''}`}
            onClick={() => onTabSelect(tab.id)}
          >
            <span>{tab.label}</span>
            <button
              className="close-tab"
              onClick={(e) => { e.stopPropagation(); onTabClose(tab.id); }}
            >&times;</button>
          </div>
        ))}
      </div>
      <div className="terminal-content">
        {tabs.map(tab => (
          <TerminalTab
            key={tab.id}
            tab={tab}
            active={activeTabId === tab.id}
            onReady={onTabReady}
          />
        ))}
      </div>
    </div>
  );
}
