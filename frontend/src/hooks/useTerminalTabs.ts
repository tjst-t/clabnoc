import { useState, useRef, useCallback } from 'react';
import type { TerminalTab } from '../types/topology';

export function useTerminalTabs(project: string | null) {
  const tabsByProject = useRef<Map<string, TerminalTab[]>>(new Map());
  const [tabs, setTabs] = useState<TerminalTab[]>([]);
  const [activeTabId, setActiveTabId] = useState<string | null>(null);
  const prevProject = useRef<string | null>(null);

  // Sync tabs when project changes
  if (project !== prevProject.current) {
    // Save current tabs
    if (prevProject.current) {
      tabsByProject.current.set(prevProject.current, tabs);
    }
    // Restore tabs for new project
    const restored = project ? tabsByProject.current.get(project) ?? [] : [];
    prevProject.current = project;
    // Use immediate state update pattern
    if (restored !== tabs) {
      setTabs(restored);
      setActiveTabId(restored.length > 0 ? restored[restored.length - 1]!.id : null);
    }
  }

  const addTab = useCallback(
    (node: string, type: 'exec' | 'ssh') => {
      const id = `${node}-${type}-${Date.now()}`;
      const label = `${node} (${type})`;
      const tab: TerminalTab = { id, node, type, label };
      setTabs((prev) => {
        const next = [...prev, tab];
        if (project) {
          tabsByProject.current.set(project, next);
        }
        return next;
      });
      setActiveTabId(id);
      return tab;
    },
    [project]
  );

  const removeTab = useCallback(
    (tabId: string) => {
      setTabs((prev) => {
        const next = prev.filter((t) => t.id !== tabId);
        if (project) {
          tabsByProject.current.set(project, next);
        }
        return next;
      });
      setActiveTabId((prev) => {
        if (prev === tabId) {
          const remaining = tabs.filter((t) => t.id !== tabId);
          return remaining.length > 0 ? remaining[remaining.length - 1]!.id : null;
        }
        return prev;
      });
    },
    [project, tabs]
  );

  return {
    tabs,
    activeTabId,
    setActiveTabId,
    addTab,
    removeTab,
  };
}
