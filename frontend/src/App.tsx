import { useState, useCallback } from 'react';
import type { TopologyNode, TopologyLink, ApiEvent, NetemParams } from './types/topology';
import { useProjects } from './hooks/useProjects';
import { useTopology } from './hooks/useTopology';
import { useWebSocket } from './hooks/useWebSocket';
import { useTerminalTabs } from './hooks/useTerminalTabs';
import { nodeAction, injectFault } from './lib/api';
import { ProjectSelector } from './components/ProjectSelector';
import { TopologyView } from './components/TopologyView';
import { NodePanel } from './components/NodePanel';
import { LinkPanel } from './components/LinkPanel';
import { TerminalPanel } from './components/TerminalPanel';
import { FaultDialog } from './components/FaultDialog';
import { destroyTerminalTab } from './components/TerminalTab';

export default function App() {
  const [selectedProject, setSelectedProject] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<TopologyNode | null>(null);
  const [selectedLink, setSelectedLink] = useState<TopologyLink | null>(null);
  const [netemDialogLink, setNetemDialogLink] = useState<TopologyLink | null>(null);
  const [terminalCollapsed, setTerminalCollapsed] = useState(false);

  const { projects, loading: projectsLoading } = useProjects();
  const { topology, refresh: refreshTopology } = useTopology(selectedProject);
  const { tabs, activeTabId, setActiveTabId, addTab, removeTab } = useTerminalTabs(selectedProject);

  // Event handler
  const handleEvent = useCallback(
    (event: ApiEvent) => {
      if (
        event.type === 'node_status_changed' ||
        event.type === 'project_changed' ||
        event.type === 'topology_changed'
      ) {
        refreshTopology();
      }
    },
    [refreshTopology]
  );

  useWebSocket(selectedProject, handleEvent);

  // Context menu for links
  const [contextMenu, setContextMenu] = useState<{
    link: TopologyLink;
    x: number;
    y: number;
  } | null>(null);

  const handleContextMenuLink = useCallback((link: TopologyLink, x: number, y: number) => {
    setContextMenu({ link, x, y });
  }, []);

  const closeContextMenu = useCallback(() => setContextMenu(null), []);

  // Handlers
  const handleOpenTerminal = useCallback(
    (node: string, type: 'exec' | 'ssh') => {
      addTab(node, type);
      setTerminalCollapsed(false);
    },
    [addTab]
  );

  const handleCloseTab = useCallback(
    (tabId: string) => {
      destroyTerminalTab(tabId);
      removeTab(tabId);
    },
    [removeTab]
  );

  const handleNodeAction = useCallback(
    async (node: string, action: 'start' | 'stop' | 'restart') => {
      if (!selectedProject) return;
      try {
        await nodeAction(selectedProject, node, { action });
        refreshTopology();
      } catch (e) {
        console.error('Node action failed:', e);
      }
    },
    [selectedProject, refreshTopology]
  );

  const handleFaultAction = useCallback(
    async (linkId: string, action: 'up' | 'down' | 'clear_netem') => {
      if (!selectedProject) return;
      try {
        await injectFault(selectedProject, linkId, { action });
        refreshTopology();
      } catch (e) {
        console.error('Fault action failed:', e);
      }
    },
    [selectedProject, refreshTopology]
  );

  const handleApplyNetem = useCallback(
    async (linkId: string, params: NetemParams) => {
      if (!selectedProject) return;
      try {
        await injectFault(selectedProject, linkId, { action: 'netem', netem: params });
        refreshTopology();
      } catch (e) {
        console.error('Netem failed:', e);
      }
    },
    [selectedProject, refreshTopology]
  );

  const hasPanel = selectedNode || selectedLink;

  return (
    <div className="h-screen flex flex-col overflow-hidden" onClick={closeContextMenu}>
      {/* Header */}
      <header className="h-11 bg-noc-panel border-b border-noc-border flex items-center px-4 shrink-0">
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2">
            <div className="w-2 h-2 bg-noc-accent rounded-full animate-pulse-slow" />
            <span className="font-display text-sm font-bold tracking-tight text-noc-text-bright">
              clabnoc
            </span>
          </div>
          <div className="w-px h-5 bg-noc-border" />
          <ProjectSelector
            projects={projects}
            selected={selectedProject}
            onSelect={setSelectedProject}
            loading={projectsLoading}
          />
        </div>
      </header>

      {/* Main content */}
      <div className="flex-1 flex overflow-hidden">
        {/* Topology view */}
        <div className="flex-1 relative">
          <TopologyView
            topology={topology}
            onSelectNode={setSelectedNode}
            onSelectLink={setSelectedLink}
            onContextMenuLink={handleContextMenuLink}
          />

          {/* Link context menu */}
          {contextMenu && (
            <div
              className="fixed z-40 bg-noc-panel border border-noc-border rounded shadow-xl py-1 min-w-40 animate-fade-in"
              style={{ left: contextMenu.x, top: contextMenu.y }}
              onClick={(e) => e.stopPropagation()}
            >
              <div className="px-3 py-1.5 text-2xs font-mono text-noc-text-dim border-b border-noc-border">
                {contextMenu.link.a.node} ↔ {contextMenu.link.z.node}
              </div>
              {contextMenu.link.state === 'up' ? (
                <>
                  <ContextMenuItem
                    label="Link Down"
                    color="text-noc-red"
                    onClick={() => {
                      handleFaultAction(contextMenu.link.id, 'down');
                      closeContextMenu();
                    }}
                  />
                  <ContextMenuItem
                    label="Apply Netem..."
                    color="text-noc-amber"
                    onClick={() => {
                      setNetemDialogLink(contextMenu.link);
                      closeContextMenu();
                    }}
                  />
                </>
              ) : contextMenu.link.state === 'down' ? (
                <ContextMenuItem
                  label="Link Up"
                  color="text-noc-green"
                  onClick={() => {
                    handleFaultAction(contextMenu.link.id, 'up');
                    closeContextMenu();
                  }}
                />
              ) : (
                <>
                  <ContextMenuItem
                    label="Clear Netem"
                    color="text-noc-green"
                    onClick={() => {
                      handleFaultAction(contextMenu.link.id, 'clear_netem');
                      closeContextMenu();
                    }}
                  />
                  <ContextMenuItem
                    label="Update Netem..."
                    color="text-noc-amber"
                    onClick={() => {
                      setNetemDialogLink(contextMenu.link);
                      closeContextMenu();
                    }}
                  />
                </>
              )}
              <ContextMenuItem
                label="View Details"
                color="text-noc-text"
                onClick={() => {
                  setSelectedLink(contextMenu.link);
                  setSelectedNode(null);
                  closeContextMenu();
                }}
              />
            </div>
          )}
        </div>

        {/* Side panel */}
        {selectedNode && (
          <NodePanel
            node={selectedNode}
            onClose={() => setSelectedNode(null)}
            onOpenTerminal={handleOpenTerminal}
            onNodeAction={handleNodeAction}
          />
        )}
        {selectedLink && !selectedNode && (
          <LinkPanel
            link={selectedLink}
            onClose={() => setSelectedLink(null)}
            onFaultAction={handleFaultAction}
            onOpenNetemDialog={setNetemDialogLink}
          />
        )}
      </div>

      {/* Terminal panel */}
      {selectedProject && (
        <TerminalPanel
          project={selectedProject}
          tabs={tabs}
          activeTabId={activeTabId}
          onSelectTab={setActiveTabId}
          onCloseTab={handleCloseTab}
          collapsed={terminalCollapsed}
          onToggle={() => setTerminalCollapsed((v) => !v)}
        />
      )}

      {/* Netem dialog */}
      {netemDialogLink && (
        <FaultDialog
          link={netemDialogLink}
          onApply={handleApplyNetem}
          onClose={() => setNetemDialogLink(null)}
        />
      )}

      {/* Bottom status bar */}
      <div className="h-5 bg-noc-bg border-t border-noc-border flex items-center px-3 shrink-0">
        <div className="flex items-center gap-3 text-2xs font-mono text-noc-text-dim">
          <span>{topology ? `${topology.nodes.length} nodes` : '—'}</span>
          <span>·</span>
          <span>{topology ? `${topology.links.length} links` : '—'}</span>
          {hasPanel && (
            <>
              <span>·</span>
              <span className="text-noc-accent">
                {selectedNode ? selectedNode.name : selectedLink ? 'link selected' : ''}
              </span>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

function ContextMenuItem({
  label,
  color,
  onClick,
}: {
  label: string;
  color: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={`w-full text-left px-3 py-1.5 text-xs font-mono hover:bg-noc-surface transition-colors cursor-pointer ${color}`}
    >
      {label}
    </button>
  );
}
