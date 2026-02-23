import type { TerminalTab as TerminalTabType } from '../types/topology';
import { TerminalTab } from './TerminalTab';

interface Props {
  project: string;
  tabs: TerminalTabType[];
  activeTabId: string | null;
  onSelectTab: (id: string) => void;
  onCloseTab: (id: string) => void;
  collapsed: boolean;
  onToggle: () => void;
}

export function TerminalPanel({
  project,
  tabs,
  activeTabId,
  onSelectTab,
  onCloseTab,
  collapsed,
  onToggle,
}: Props) {
  return (
    <div
      className={`bg-noc-panel border-t border-noc-border flex flex-col transition-all duration-200 ${
        collapsed ? 'h-8' : 'h-64'
      }`}
    >
      {/* Tab bar */}
      <div className="flex items-center border-b border-noc-border h-8 shrink-0">
        <button
          onClick={onToggle}
          className="px-3 h-full flex items-center gap-1.5 text-2xs font-mono uppercase tracking-widest
                     text-noc-text-dim hover:text-noc-text transition-colors cursor-pointer"
        >
          <svg
            width="8"
            height="8"
            viewBox="0 0 8 8"
            fill="none"
            className={`transition-transform ${collapsed ? 'rotate-180' : ''}`}
          >
            <path d="M1 2L4 5L7 2" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
          </svg>
          Terminal
          {tabs.length > 0 && (
            <span className="ml-1 text-noc-accent">{tabs.length}</span>
          )}
        </button>

        <div className="flex-1 flex items-center gap-0 overflow-x-auto">
          {tabs.map((tab) => (
            <div
              key={tab.id}
              className={`group flex items-center gap-1.5 px-3 h-full text-2xs font-mono cursor-pointer
                         border-r border-noc-border transition-colors
                         ${tab.id === activeTabId
                           ? 'bg-noc-bg text-noc-accent border-b-2 border-b-noc-accent'
                           : 'text-noc-text-dim hover:text-noc-text hover:bg-noc-surface'}`}
              onClick={() => onSelectTab(tab.id)}
            >
              <span className={`w-1 h-1 rounded-full ${
                tab.type === 'exec' ? 'bg-noc-accent' : 'bg-noc-cyan'
              }`} />
              <span className="truncate max-w-32">{tab.label}</span>
              <button
                onClick={(e) => { e.stopPropagation(); onCloseTab(tab.id); }}
                className="opacity-0 group-hover:opacity-100 transition-opacity ml-1 hover:text-noc-red"
              >
                <svg width="8" height="8" viewBox="0 0 8 8" fill="none">
                  <path d="M1 1L7 7M7 1L1 7" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Terminal content */}
      {!collapsed && (
        <div className="flex-1 relative">
          {tabs.length === 0 ? (
            <div className="flex items-center justify-center h-full">
              <span className="text-2xs font-mono text-noc-text-dim tracking-widest uppercase">
                Click a node access method to open a terminal
              </span>
            </div>
          ) : (
            tabs.map((tab) => (
              <div
                key={tab.id}
                className="absolute inset-0"
                style={{ display: tab.id === activeTabId ? 'block' : 'none' }}
              >
                <TerminalTab project={project} tab={tab} active={tab.id === activeTabId} />
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
}
