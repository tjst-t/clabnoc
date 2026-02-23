import React, { useState, useCallback, useEffect } from 'react';
import { ProjectSelector } from './components/ProjectSelector';
import { TopologyView } from './components/TopologyView';
import { NodePanel } from './components/NodePanel';
import { LinkPanel } from './components/LinkPanel';
import { FaultDialog } from './components/FaultDialog';
import { TerminalPanel } from './components/TerminalPanel';
import { useProjects } from './hooks/useProjects';
import { useTopology } from './hooks/useTopology';
import { useTerminalTabs } from './hooks/useTerminalTabs';
import { useWebSocket } from './hooks/useWebSocket';
import { getExecWSUrl, getSSHWSUrl, injectFault, nodeAction, getEventsWSUrl } from './lib/api';
import type { NodeInfo, LinkInfo, NetemConfig } from './types/topology';
import type { Terminal } from '@xterm/xterm';
import './App.css';

export function App() {
  const { projects, loading: projectsLoading, error: projectsError, refresh: refreshProjects } = useProjects();
  const [selectedProject, setSelectedProject] = useState<string | null>(null);
  const { topology, loading: topoLoading, refresh: refreshTopology } = useTopology(selectedProject);
  const [selectedNode, setSelectedNode] = useState<NodeInfo | null>(null);
  const [selectedLink, setSelectedLink] = useState<LinkInfo | null>(null);
  const [faultDialogLink, setFaultDialogLink] = useState<LinkInfo | null>(null);
  const {
    tabs, activeTabId, setActiveTabId,
    switchProject, addTab, removeTab, updateTab
  } = useTerminalTabs();

  const handleProjectSelect = useCallback((name: string) => {
    setSelectedProject(name);
    setSelectedNode(null);
    setSelectedLink(null);
    switchProject(name);
  }, [switchProject]);

  const handleNodeClick = useCallback((node: NodeInfo) => {
    setSelectedNode(node);
    setSelectedLink(null);
  }, []);

  const handleLinkClick = useCallback((link: LinkInfo) => {
    setSelectedLink(link);
    setSelectedNode(null);
  }, []);

  const handleOpenTerminal = useCallback((node: string, type: 'exec' | 'ssh') => {
    if (!selectedProject) return;
    const wsUrl = type === 'exec'
      ? getExecWSUrl(selectedProject, node)
      : getSSHWSUrl(selectedProject, node);
    addTab(selectedProject, node, type, wsUrl);
  }, [selectedProject, addTab]);

  const handleTabReady = useCallback((id: string, terminal: Terminal, socket: WebSocket) => {
    updateTab(id, { terminal, socket });
  }, [updateTab]);

  const handleFaultDown = useCallback(async (linkId: string) => {
    if (!selectedProject) return;
    try {
      await injectFault(selectedProject, linkId, 'down');
      refreshTopology();
    } catch (err) {
      console.error('fault down failed:', err);
    }
  }, [selectedProject, refreshTopology]);

  const handleFaultUp = useCallback(async (linkId: string) => {
    if (!selectedProject) return;
    try {
      await injectFault(selectedProject, linkId, 'up');
      refreshTopology();
    } catch (err) {
      console.error('fault up failed:', err);
    }
  }, [selectedProject, refreshTopology]);

  const handleFaultNetem = useCallback((link: LinkInfo) => {
    setFaultDialogLink(link);
  }, []);

  const handleApplyNetem = useCallback(async (linkId: string, params: NetemConfig) => {
    if (!selectedProject) return;
    try {
      await injectFault(selectedProject, linkId, 'netem', params);
      setFaultDialogLink(null);
      refreshTopology();
    } catch (err) {
      console.error('netem failed:', err);
    }
  }, [selectedProject, refreshTopology]);

  const handleNodeAction = useCallback(async (node: string, action: 'start' | 'stop' | 'restart') => {
    if (!selectedProject) return;
    try {
      await nodeAction(selectedProject, node, action);
      refreshTopology();
    } catch (err) {
      console.error('node action failed:', err);
    }
  }, [selectedProject, refreshTopology]);

  const handleEventMessage = useCallback((data: unknown) => {
    const event = data as { type?: string };
    if (event.type === 'node_status_changed' || event.type === 'topology_changed') {
      refreshTopology();
    } else if (event.type === 'project_changed') {
      refreshProjects();
    }
  }, [refreshTopology, refreshProjects]);

  useWebSocket({
    url: getEventsWSUrl(selectedProject || undefined),
    onMessage: handleEventMessage,
    enabled: !!selectedProject,
  });

  // Select first project on load
  useEffect(() => {
    if (projects.length > 0 && !selectedProject) {
      handleProjectSelect(projects[0].name);
    }
  }, [projects, selectedProject, handleProjectSelect]);

  return (
    <div className="app">
      <header className="app-header">
        <h1>clabnoc</h1>
        <span className="subtitle">Containerlab Network Operations Center</span>
      </header>
      <div className="app-body">
        <aside className="sidebar">
          <ProjectSelector
            projects={projects}
            selected={selectedProject}
            onSelect={handleProjectSelect}
            loading={projectsLoading}
            error={projectsError}
            onRefresh={refreshProjects}
          />
        </aside>
        <main className="main-content">
          <div className="topology-container">
            {topoLoading ? (
              <div className="loading">Loading topology...</div>
            ) : (
              <TopologyView topology={topology} onNodeClick={handleNodeClick} onLinkClick={handleLinkClick} />
            )}
          </div>
          <div className="terminal-container">
            <TerminalPanel
              tabs={tabs}
              activeTabId={activeTabId}
              onTabSelect={setActiveTabId}
              onTabClose={removeTab}
              onTabReady={handleTabReady}
            />
          </div>
        </main>
        {selectedLink ? (
          <aside className="node-panel-container">
            <LinkPanel
              link={selectedLink}
              project={selectedProject || ''}
              onFaultDown={handleFaultDown}
              onFaultUp={handleFaultUp}
              onFaultNetem={handleFaultNetem}
              onClose={() => setSelectedLink(null)}
            />
          </aside>
        ) : selectedNode ? (
          <aside className="node-panel-container">
            <NodePanel
              node={selectedNode}
              project={selectedProject || ''}
              onOpenTerminal={handleOpenTerminal}
              onClose={() => setSelectedNode(null)}
              onAction={handleNodeAction}
            />
          </aside>
        ) : null}
      </div>
      {faultDialogLink && (
        <FaultDialog
          link={faultDialogLink}
          onApply={handleApplyNetem}
          onCancel={() => setFaultDialogLink(null)}
        />
      )}
    </div>
  );
}

export default App;
