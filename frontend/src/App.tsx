import { useState, useCallback, useEffect } from 'react';
import type { TopologyNode, TopologyLink, ApiEvent, NetemParams, SSHCredentials } from './types/topology';
import { useProjects } from './hooks/useProjects';
import { useTopology } from './hooks/useTopology';
import { useWebSocket } from './hooks/useWebSocket';
import { useTerminalTabs } from './hooks/useTerminalTabs';
import { useResizable } from './hooks/useResizable';
import { useContainerStats } from './hooks/useContainerStats';
import { nodeAction, injectFault, captureAction, getCaptureDownloadUrl } from './lib/api';
import { ProjectSelector } from './components/ProjectSelector';
import { TopologyView } from './components/TopologyView';
import { NodeTable } from './components/NodeTable';
import { DetailPanel } from './components/DetailPanel';
import { TerminalPanel } from './components/TerminalPanel';
import { FaultDialog } from './components/FaultDialog';
import { SSHDialog } from './components/SSHDialog';
import { destroyTerminalTab } from './lib/terminal-store';
import { ThemeProvider, useTheme } from './lib/theme';

function useIsMobile(breakpoint = 768) {
  const [isMobile, setIsMobile] = useState(() => window.innerWidth < breakpoint);
  useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${breakpoint - 1}px)`);
    const handler = (e: MediaQueryListEvent) => setIsMobile(e.matches);
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, [breakpoint]);
  return isMobile;
}

export default function App() {
  return (
    <ThemeProvider>
      <AppContent />
    </ThemeProvider>
  );
}

function AppContent() {
  const isMobile = useIsMobile();
  const { theme, toggleTheme } = useTheme();
  const [selectedProject, setSelectedProject] = useState<string | null>(() => {
    const params = new URLSearchParams(window.location.search);
    return params.get('project');
  });
  const [selectedNode, setSelectedNode] = useState<TopologyNode | null>(null);
  const [selectedLink, setSelectedLink] = useState<TopologyLink | null>(null);
  const [netemDialogLink, setNetemDialogLink] = useState<TopologyLink | null>(null);
  const [sshDialogNode, setSSHDialogNode] = useState<{ name: string; kind: string } | null>(null);
  const [terminalCollapsed, setTerminalCollapsed] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [viewMode, setViewMode] = useState<'topology' | 'table'>('topology');
  const [capturingLinks, setCapturingLinks] = useState<Set<string>>(new Set());

  // Sync selected project to URL
  useEffect(() => {
    const url = new URL(window.location.href);
    if (selectedProject) {
      url.searchParams.set('project', selectedProject);
    } else {
      url.searchParams.delete('project');
    }
    window.history.replaceState(null, '', url.toString());
  }, [selectedProject]);

  const { projects, loading: projectsLoading } = useProjects();
  const { topology, refresh: refreshTopology } = useTopology(selectedProject);
  const { tabs, activeTabId, setActiveTabId, addTab, removeTab } = useTerminalTabs(selectedProject);
  const containerStats = useContainerStats(selectedProject);

  // Sync selectedLink with refreshed topology data
  useEffect(() => {
    if (!topology || !selectedLink) return;
    const updated = topology.links.find((l) => l.id === selectedLink.id);
    if (updated && (updated.state !== selectedLink.state || JSON.stringify(updated.netem) !== JSON.stringify(selectedLink.netem))) {
      setSelectedLink(updated);
    }
  }, [topology, selectedLink]);

  const { size: detailWidth, handleMouseDown: onDetailDrag } = useResizable({
    direction: 'horizontal',
    initialSize: 288,
    minSize: 200,
    maxSize: 600,
  });
  const { size: terminalHeight, handleMouseDown: onTerminalDrag } = useResizable({
    direction: 'vertical',
    initialSize: 256,
    minSize: 80,
    maxSize: Math.round(window.innerHeight * 0.6),
  });

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
      if (type === 'ssh') {
        // Find node kind from topology for SSH dialog
        const nodeData = topology?.nodes.find((n) => n.name === node);
        setSSHDialogNode({ name: node, kind: nodeData?.kind ?? '' });
      } else {
        addTab(node, type);
        setTerminalCollapsed(false);
      }
    },
    [addTab, topology]
  );

  const handleSSHConnect = useCallback(
    (nodeName: string, credentials: SSHCredentials) => {
      addTab(nodeName, 'ssh', credentials);
      setTerminalCollapsed(false);
      setSSHDialogNode(null);
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

  const handleStartCapture = useCallback(
    async (linkId: string) => {
      if (!selectedProject) return;
      try {
        await captureAction(selectedProject, linkId, { action: 'start' });
        setCapturingLinks((prev) => new Set(prev).add(linkId));
      } catch (e) {
        console.error('Start capture failed:', e);
      }
    },
    [selectedProject]
  );

  const handleStopCapture = useCallback(
    async (linkId: string) => {
      if (!selectedProject) return;
      try {
        await captureAction(selectedProject, linkId, { action: 'stop' });
        setCapturingLinks((prev) => {
          const next = new Set(prev);
          next.delete(linkId);
          return next;
        });
      } catch (e) {
        console.error('Stop capture failed:', e);
      }
    },
    [selectedProject]
  );

  const handleDownloadCapture = useCallback(
    (linkId: string) => {
      if (!selectedProject) return;
      const url = getCaptureDownloadUrl(selectedProject, linkId);
      window.open(url, '_blank');
    },
    [selectedProject]
  );

  return (
    <div className="h-screen flex flex-col overflow-hidden bg-noc-bg text-noc-text font-mono" onClick={closeContextMenu}>
      {/* ┌─── Header ───┐ */}
      <header className="h-8 bg-noc-bg tui-border-b flex items-center px-2 shrink-0">
        <div className="flex items-center gap-3 flex-1">
          <span className="text-noc-accent text-xs font-bold">clabnoc</span>
          <span className="text-noc-border">|</span>
          <ProjectSelector
            projects={projects}
            selected={selectedProject}
            onSelect={setSelectedProject}
            loading={projectsLoading}
          />
        </div>
        <div className="flex items-center gap-2 text-2xs text-noc-text-dim">
          {topology && (
            <>
              <div className="relative">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="/ search nodes..."
                  className="bg-transparent border border-noc-border text-noc-text text-2xs px-1.5 py-0.5 w-36 focus:outline-none focus:border-noc-accent placeholder:text-noc-text-dim"
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute right-1 top-1/2 -translate-y-1/2 text-noc-text-dim hover:text-noc-text text-[9px]"
                  >
                    x
                  </button>
                )}
              </div>
              <span className="text-noc-border">|</span>
              <button
                onClick={() => setViewMode('topology')}
                className={`tui-btn ${viewMode === 'topology' ? 'tui-btn-cyan' : 'tui-btn-dim'}`}
                title="Topology view"
              >
                topo
              </button>
              <button
                onClick={() => setViewMode('table')}
                className={`tui-btn ${viewMode === 'table' ? 'tui-btn-cyan' : 'tui-btn-dim'}`}
                title="Table view"
              >
                table
              </button>
              <span className="text-noc-border">|</span>
            </>
          )}
          <span>{topology ? `${topology.nodes.length} nodes` : '--'}</span>
          <span className="text-noc-border">|</span>
          <span>{topology ? `${topology.links.length} links` : '--'}</span>
          <span className="text-noc-border">|</span>
          <button
            onClick={toggleTheme}
            className="tui-btn tui-btn-dim"
            title={`Switch to ${theme === 'dark' ? 'light' : 'dark'} mode`}
          >
            {theme === 'dark' ? '\u2600' : '\u263D'}
          </button>
        </div>
      </header>

      {/* ├─── Main ───┤ */}
      <div className={`flex-1 flex flex-col ${isMobile ? 'overflow-y-auto' : 'overflow-hidden'}`}>
        {/* Desktop: horizontal row. Mobile: stacked vertically */}
        <div className={isMobile ? '' : 'flex-1 flex overflow-hidden'}>
          {/* Topology / Table view */}
          <div className={`relative min-w-0 ${isMobile ? 'h-[60vh] shrink-0' : 'flex-1'}`}>
            {viewMode === 'table' ? (
              <NodeTable
                topology={topology}
                onSelectNode={(node) => {
                  setSelectedNode(node);
                  setSelectedLink(null);
                }}
                selectedNodeName={selectedNode?.name ?? null}
                searchQuery={searchQuery}
                containerStats={containerStats}
              />
            ) : (
              <TopologyView
                topology={topology}
                onSelectNode={setSelectedNode}
                onSelectLink={setSelectedLink}
                onContextMenuLink={handleContextMenuLink}
                searchQuery={searchQuery}
              />
            )}

            {/* Link context menu */}
            {contextMenu && (
              <div
                className="fixed z-40 bg-noc-bg tui-border py-1 min-w-44 animate-fade-in shadow-lg"
                style={{ left: contextMenu.x, top: contextMenu.y }}
                onClick={(e) => e.stopPropagation()}
              >
                <div className="px-2 py-1 text-2xs text-noc-text-dim tui-border-b">
                  {contextMenu.link.a.node} &lt;-&gt; {contextMenu.link.z.node}
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
                <div className="tui-border-t my-0.5" />
                {capturingLinks.has(contextMenu.link.id) ? (
                  <>
                    <ContextMenuItem
                      label="Stop Capture"
                      color="text-noc-red"
                      onClick={() => {
                        handleStopCapture(contextMenu.link.id);
                        closeContextMenu();
                      }}
                    />
                    <ContextMenuItem
                      label="Download Pcap"
                      color="text-noc-cyan"
                      onClick={() => {
                        handleDownloadCapture(contextMenu.link.id);
                        closeContextMenu();
                      }}
                    />
                  </>
                ) : (
                  <ContextMenuItem
                    label="Start Capture"
                    color="text-noc-cyan"
                    onClick={() => {
                      handleStartCapture(contextMenu.link.id);
                      closeContextMenu();
                    }}
                  />
                )}
              </div>
            )}
          </div>

          {/* Right panel drag handle (desktop only) */}
          {!isMobile && <div className="drag-handle-v" onMouseDown={onDetailDrag} />}

          {/* Detail panel */}
          <DetailPanel
            project={selectedProject}
            node={selectedNode}
            link={selectedLink}
            onOpenTerminal={handleOpenTerminal}
            onNodeAction={handleNodeAction}
            onFaultAction={handleFaultAction}
            onOpenNetemDialog={setNetemDialogLink}
            onStartCapture={handleStartCapture}
            onStopCapture={handleStopCapture}
            onDownloadCapture={handleDownloadCapture}
            capturingLinks={capturingLinks}
            style={isMobile ? undefined : { width: detailWidth }}
            mobile={isMobile}
            containerStats={containerStats}
          />
        </div>

        {/* ├─── Terminal ───┤ */}
        {selectedProject && (
          <>
            {!isMobile && !terminalCollapsed && (
              <div className="drag-handle-h" onMouseDown={onTerminalDrag} />
            )}
            <TerminalPanel
              project={selectedProject}
              tabs={tabs}
              activeTabId={activeTabId}
              onSelectTab={setActiveTabId}
              onCloseTab={handleCloseTab}
              collapsed={isMobile ? false : terminalCollapsed}
              onToggle={() => setTerminalCollapsed((v) => !v)}
              style={isMobile ? { height: 300 } : { height: terminalHeight }}
            />
          </>
        )}
      </div>

      {/* Netem dialog */}
      {netemDialogLink && (
        <FaultDialog
          link={netemDialogLink}
          onApply={handleApplyNetem}
          onClose={() => setNetemDialogLink(null)}
        />
      )}

      {/* SSH dialog */}
      {sshDialogNode && selectedProject && (
        <SSHDialog
          node={sshDialogNode}
          project={selectedProject}
          onConnect={(creds) => handleSSHConnect(sshDialogNode.name, creds)}
          onClose={() => setSSHDialogNode(null)}
        />
      )}
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
      className={`w-full text-left px-2 py-1 text-xs hover:bg-noc-surface transition-colors cursor-pointer ${color}`}
    >
      {label}
    </button>
  );
}
