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
      className={`bg-noc-bg tui-border-t flex flex-col transition-all duration-200 ${
        collapsed ? 'h-7' : 'h-64'
      }`}
    >
      {/* Tab bar */}
      <div className="flex items-center h-7 shrink-0 tui-border-b">
        <button
          onClick={onToggle}
          className="px-2 h-full flex items-center gap-1 text-2xs
                     text-noc-text-dim hover:text-noc-text transition-colors cursor-pointer"
        >
          <span>{collapsed ? '>' : 'v'}</span>
          <span>Terminal</span>
          {tabs.length > 0 && (
            <span className="text-noc-accent">[{tabs.length}]</span>
          )}
        </button>

        <span className="text-noc-border">|</span>

        <div className="flex-1 flex items-center gap-0 overflow-x-auto">
          {tabs.map((tab) => (
            <div
              key={tab.id}
              className={`group flex items-center gap-1 px-2 h-full text-2xs cursor-pointer
                         transition-colors
                         ${
                           tab.id === activeTabId
                             ? 'text-noc-accent bg-noc-surface'
                             : 'text-noc-text-dim hover:text-noc-text'
                         }`}
              onClick={() => onSelectTab(tab.id)}
            >
              <span
                className={
                  tab.type === 'exec' ? 'text-noc-accent' : 'text-noc-cyan'
                }
              >
                {tab.type === 'exec' ? '$' : '>'}
              </span>
              <span className="truncate max-w-32">{tab.label}</span>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onCloseTab(tab.id);
                }}
                className="opacity-0 group-hover:opacity-100 transition-opacity ml-1 hover:text-noc-red text-noc-text-dim"
              >
                x
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
              <span className="text-2xs text-noc-text-dim">
                -- Select a node and open a terminal --
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
