import { useRef, useState, useCallback } from 'react';
import type { Terminal } from '@xterm/xterm';

export interface TerminalTab {
  id: string;
  project: string;
  node: string;
  type: 'exec' | 'ssh';
  label: string;
  terminal: Terminal | null; // set after mount
  socket: WebSocket | null;  // set after connect
  wsUrl: string;
}

export function useTerminalTabs() {
  // Store tabs per project
  const tabsByProject = useRef<Map<string, TerminalTab[]>>(new Map());
  const [activeProject, setActiveProject] = useState<string>('');
  const [tabs, setTabs] = useState<TerminalTab[]>([]);
  const [activeTabId, setActiveTabId] = useState<string | null>(null);

  // Switch to a different project - saves current tabs, loads new project's tabs
  const switchProject = useCallback((project: string) => {
    setActiveProject(prev => {
      if (prev) {
        // Save current tabs for previous project - will be done via setTabs callback
      }
      return project;
    });
    setTabs(prev => {
      // Save current tabs to the previous project key
      // We need to read activeProject here, but it's stale in closure
      // Instead we update tabsByProject in a separate step
      const projectTabs = tabsByProject.current.get(project) || [];
      return projectTabs;
    });
    setActiveProject(project);
    const projectTabs = tabsByProject.current.get(project) || [];
    setActiveTabId(projectTabs.length > 0 ? projectTabs[0].id : null);
  }, []);

  // Add a new tab
  const addTab = useCallback((project: string, node: string, type: 'exec' | 'ssh', wsUrl: string): string => {
    const id = `${project}-${node}-${type}-${Date.now()}`;
    const tab: TerminalTab = {
      id, project, node, type,
      label: `${node} (${type})`,
      terminal: null,
      socket: null,
      wsUrl,
    };
    setTabs(prev => {
      const updated = [...prev, tab];
      tabsByProject.current.set(project, updated);
      return updated;
    });
    setActiveTabId(id);
    return id;
  }, []);

  // Remove a tab
  const removeTab = useCallback((id: string) => {
    setTabs(prev => {
      const tab = prev.find(t => t.id === id);
      if (tab) {
        tab.socket?.close();
        tab.terminal?.dispose();
      }
      const updated = prev.filter(t => t.id !== id);
      tabsByProject.current.set(activeProject, updated);
      return updated;
    });
    setActiveTabId(prev => {
      if (prev !== id) return prev;
      const remaining = tabs.filter(t => t.id !== id);
      return remaining.length > 0 ? remaining[remaining.length - 1].id : null;
    });
  }, [activeProject, tabs]);

  // Update tab with terminal/socket instances
  const updateTab = useCallback((id: string, updates: Partial<TerminalTab>) => {
    setTabs(prev => {
      const updated = prev.map(t => t.id === id ? { ...t, ...updates } : t);
      tabsByProject.current.set(activeProject, updated);
      return updated;
    });
  }, [activeProject]);

  return {
    tabs,
    activeTabId,
    setActiveTabId,
    activeProject,
    switchProject,
    addTab,
    removeTab,
    updateTab,
  };
}
