import { useMemo, useEffect, useCallback, useState } from 'react';
import type { Topology, TopologyNode, TopologyLink, ExternalNode, ExternalNetwork } from '../types/topology';
import {
  computeLayout,
  buildExternalCablePath,
  LINK_STATE_COLORS,
  EXTERNAL_LINK_COLOR,
  SERVICES_AREA_HEIGHT,
  type CableLayout,
  type DeviceLayout,
  type PortLayout,
  type RackLayout,
} from '../lib/rack-layout';
import { useZoomPan } from '../hooks/useZoomPan';
import { DataCenter } from './topology/DataCenter';
import { Rack } from './topology/Rack';
import { DeviceFaceplate } from './topology/DeviceFaceplate';
import { CableOverlay } from './topology/CableOverlay';
import { ExternalDevice } from './topology/ExternalDevice';
import { ExternalNetworkCloud } from './topology/ExternalNetworkCloud';
import { ExternalNetworkBar } from './topology/ExternalNetworkBar';
import { ServicesArea } from './topology/ServicesArea';

interface Props {
  topology: Topology | null;
  onSelectNode: (node: TopologyNode | null) => void;
  onSelectLink: (link: TopologyLink | null) => void;
  onContextMenuLink: (link: TopologyLink, x: number, y: number) => void;
  onSelectExternalNode?: (node: ExternalNode | null) => void;
  onSelectExternalNetwork?: (network: ExternalNetwork | null) => void;
  searchQuery?: string;
}

/** Get the color for a cable based on link state */
function getCableColor(link: TopologyLink): string {
  if (link.state === 'down') return LINK_STATE_COLORS.down!;
  if (link.state === 'degraded') return LINK_STATE_COLORS.degraded!;
  return LINK_STATE_COLORS.up!;
}

interface AnnotationError {
  node: string;
  missing: string[];
}

/** Validate that all nodes have the required graph annotations for rack visualization */
function validateAnnotations(topology: Topology): AnnotationError[] {
  const errors: AnnotationError[] = [];
  for (const node of topology.nodes) {
    const missing: string[] = [];
    if (!node.graph.dc) missing.push('graph-dc');
    // DC-only nodes (Services area placement) — no rack/unit needed
    if (node.graph.dc && !node.graph.rack) continue;
    if (!node.graph.rack) missing.push('graph-rack');
    if (!node.graph.rack_unit) missing.push('graph-rack-unit');
    if (missing.length > 0) {
      errors.push({ node: node.name, missing });
    }
  }
  return errors;
}

export function TopologyView({ topology, onSelectNode, onSelectLink, onContextMenuLink, onSelectExternalNode, onSelectExternalNetwork, searchQuery }: Props) {
  // Selection state: either a device or a specific port
  const [selectedDeviceName, setSelectedDeviceName] = useState<string | null>(null);
  const [selectedPortKey, setSelectedPortKey] = useState<string | null>(null);
  const [selectedLinkId, setSelectedLinkId] = useState<string | null>(null);
  const [selectedExternalName, setSelectedExternalName] = useState<string | null>(null);
  const [selectedNetworkName, setSelectedNetworkName] = useState<string | null>(null);

  const annotationErrors = useMemo(
    () => (topology ? validateAnnotations(topology) : []),
    [topology]
  );

  const layout = useMemo(
    () => (topology && annotationErrors.length === 0 ? computeLayout(topology) : null),
    [topology, annotationErrors]
  );

  const rackMap = useMemo(() => {
    if (!layout) return new Map<string, RackLayout>();
    const m = new Map<string, RackLayout>();
    for (const r of layout.racks) m.set(r.id, r);
    return m;
  }, [layout]);

  const { containerRef, transform, onPointerDown, onPointerMove, onPointerUp, fitContent } =
    useZoomPan({ minScale: 0.15, maxScale: 6, wheelSensitivity: 0.002 });

  // Fit content when layout changes
  useEffect(() => {
    if (layout) {
      const id = requestAnimationFrame(() => {
        fitContent(layout.totalWidth, layout.totalHeight);
      });
      return () => cancelAnimationFrame(id);
    }
  }, [layout, fitContent]);

  // Faulted cable IDs + port colors — always computed (no selection needed)
  const { faultedCableIds, faultedPortColors, faultedDevices } = useMemo(() => {
    const ids = new Set<string>();
    const portColors = new Map<string, string>();
    const devs = new Set<string>();
    if (!layout) return { faultedCableIds: ids, faultedPortColors: portColors, faultedDevices: devs };

    for (const cable of layout.cables) {
      if (cable.link.state === 'down' || cable.link.state === 'degraded') {
        ids.add(cable.link.id);
        const color = getCableColor(cable.link);
        portColors.set(cable.aPort.key, color);
        portColors.set(cable.zPort.key, color);
        devs.add(cable.aPort.node.name);
        devs.add(cable.zPort.node.name);
      }
    }
    return { faultedCableIds: ids, faultedPortColors: portColors, faultedDevices: devs };
  }, [layout]);

  // Highlighted cables + related devices — selection-driven
  const { highlightedCableIds, relatedDevices, highlightedPorts, cableColorMap, visibleCables } = useMemo(() => {
    const empty = {
      highlightedCableIds: new Set<string>(),
      relatedDevices: new Set<string>(),
      highlightedPorts: new Set<string>(),
      cableColorMap: new Map<string, string>(),
      visibleCables: [] as CableLayout[],
    };
    if (!layout || (!selectedDeviceName && !selectedLinkId)) return empty;

    let related: CableLayout[];

    if (selectedLinkId) {
      related = layout.cables.filter((c) => c.link.id === selectedLinkId);
    } else if (selectedPortKey && selectedDeviceName) {
      related = layout.cables.filter(
        (c) => c.aPort.key === selectedPortKey || c.zPort.key === selectedPortKey
      );
    } else if (selectedDeviceName) {
      related = layout.cables.filter(
        (c) => c.aPort.node.name === selectedDeviceName || c.zPort.node.name === selectedDeviceName
      );
    } else {
      return empty;
    }

    const ids = new Set<string>();
    const devs = new Set<string>();
    const ports = new Set<string>();
    const colorMap = new Map<string, string>();

    if (selectedDeviceName) devs.add(selectedDeviceName);

    for (const cable of related) {
      ids.add(cable.link.id);
      devs.add(cable.aPort.node.name);
      devs.add(cable.zPort.node.name);
      ports.add(cable.aPort.key);
      ports.add(cable.zPort.key);
      const color = getCableColor(cable.link);
      colorMap.set(cable.aPort.key, color);
      colorMap.set(cable.zPort.key, color);
    }

    return {
      highlightedCableIds: ids,
      relatedDevices: devs,
      highlightedPorts: ports,
      cableColorMap: colorMap,
      visibleCables: related,
    };
  }, [layout, selectedDeviceName, selectedPortKey, selectedLinkId]);

  const hasSelection = selectedDeviceName !== null || selectedLinkId !== null || selectedExternalName !== null || selectedNetworkName !== null;

  // Search filter: compute set of matched node names (includes external nodes)
  const { searchMatchedNodes, searchMatchedExternals } = useMemo(() => {
    if (!searchQuery || !topology) return { searchMatchedNodes: null, searchMatchedExternals: null };
    const q = searchQuery.toLowerCase();
    const matched = new Set<string>();
    for (const node of topology.nodes) {
      if (
        node.name.toLowerCase().includes(q) ||
        node.kind.toLowerCase().includes(q) ||
        node.status.toLowerCase().includes(q)
      ) {
        matched.add(node.name);
      }
    }
    const matchedExt = new Set<string>();
    for (const en of topology.external_nodes ?? []) {
      if (
        en.name.toLowerCase().includes(q) ||
        en.label.toLowerCase().includes(q)
      ) {
        matchedExt.add(en.name);
      }
    }
    return { searchMatchedNodes: matched, searchMatchedExternals: matchedExt };
  }, [searchQuery, topology]);

  // ── Handlers ──

  const handleDeviceClick = useCallback(
    (node: TopologyNode) => {
      // Toggle: clicking the same device deselects
      if (selectedDeviceName === node.name && !selectedPortKey) {
        setSelectedDeviceName(null);
        setSelectedPortKey(null);
        setSelectedLinkId(null);
        onSelectNode(null);
        onSelectLink(null);
        return;
      }
      setSelectedDeviceName(node.name);
      setSelectedPortKey(null);
      setSelectedLinkId(null);
      setSelectedExternalName(null);
      setSelectedNetworkName(null);
      onSelectNode(node);
      onSelectLink(null);
    },
    [selectedDeviceName, selectedPortKey, onSelectNode, onSelectLink]
  );

  const handlePortClick = useCallback(
    (port: PortLayout) => {
      // Toggle: clicking the same port deselects
      if (selectedPortKey === port.key) {
        setSelectedDeviceName(null);
        setSelectedPortKey(null);
        setSelectedLinkId(null);
        onSelectNode(null);
        onSelectLink(null);
        return;
      }
      setSelectedDeviceName(port.node.name);
      setSelectedPortKey(port.key);
      setSelectedLinkId(null);
      setSelectedExternalName(null);
      setSelectedNetworkName(null);
      onSelectNode(port.node);
      onSelectLink(null);
    },
    [selectedPortKey, onSelectNode, onSelectLink]
  );

  const handleSelectLink = useCallback(
    (link: TopologyLink) => {
      if (selectedLinkId === link.id) {
        setSelectedLinkId(null);
        setSelectedDeviceName(null);
        setSelectedPortKey(null);
        onSelectLink(null);
        onSelectNode(null);
        return;
      }
      setSelectedLinkId(link.id);
      setSelectedDeviceName(null);
      setSelectedPortKey(null);
      setSelectedExternalName(null);
      setSelectedNetworkName(null);
      onSelectLink(link);
      onSelectNode(null);
    },
    [selectedLinkId, onSelectNode, onSelectLink]
  );

  const handleContextMenu = useCallback(
    (link: TopologyLink, x: number, y: number) => {
      onContextMenuLink(link, x, y);
    },
    [onContextMenuLink]
  );

  const handleExternalNodeClick = useCallback(
    (en: ExternalNode) => {
      if (selectedExternalName === en.name) {
        setSelectedExternalName(null);
        onSelectExternalNode?.(null);
        return;
      }
      setSelectedDeviceName(null);
      setSelectedPortKey(null);
      setSelectedLinkId(null);
      setSelectedNetworkName(null);
      setSelectedExternalName(en.name);
      onSelectNode(null);
      onSelectLink(null);
      onSelectExternalNode?.(en);
      onSelectExternalNetwork?.(null);
    },
    [selectedExternalName, onSelectNode, onSelectLink, onSelectExternalNode, onSelectExternalNetwork]
  );

  const handleNetworkClick = useCallback(
    (net: ExternalNetwork) => {
      if (selectedNetworkName === net.name) {
        setSelectedNetworkName(null);
        onSelectExternalNetwork?.(null);
        return;
      }
      setSelectedDeviceName(null);
      setSelectedPortKey(null);
      setSelectedLinkId(null);
      setSelectedExternalName(null);
      setSelectedNetworkName(net.name);
      onSelectNode(null);
      onSelectLink(null);
      onSelectExternalNode?.(null);
      onSelectExternalNetwork?.(net);
    },
    [selectedNetworkName, onSelectNode, onSelectLink, onSelectExternalNode, onSelectExternalNetwork]
  );

  const handleBackgroundClick = useCallback(() => {
    setSelectedDeviceName(null);
    setSelectedPortKey(null);
    setSelectedLinkId(null);
    setSelectedExternalName(null);
    setSelectedNetworkName(null);
    onSelectNode(null);
    onSelectLink(null);
    onSelectExternalNode?.(null);
    onSelectExternalNetwork?.(null);
  }, [onSelectNode, onSelectLink, onSelectExternalNode, onSelectExternalNetwork]);

  // Group ports by device
  const portsByDevice = useMemo(() => {
    if (!layout) return new Map<string, PortLayout[]>();
    const m = new Map<string, PortLayout[]>();
    for (const [, port] of layout.ports) {
      const name = port.node.name;
      if (!m.has(name)) m.set(name, []);
      m.get(name)!.push(port);
    }
    return m;
  }, [layout]);

  // Build info panel data
  const infoPanelData = useMemo(() => {
    if (!layout || !selectedDeviceName) return null;
    const device = layout.devices.get(selectedDeviceName);
    if (!device) return null;

    const rack = device.rackId ? rackMap.get(device.rackId) : null;

    const connections = visibleCables.map((cable) => {
      const isA = cable.aPort.node.name === selectedDeviceName;
      const myPort = isA ? cable.aPort : cable.zPort;
      const otherPort = isA ? cable.zPort : cable.aPort;
      const color = getCableColor(cable.link);
      return {
        myInterface: myPort.iface,
        otherDevice: otherPort.node.name,
        otherInterface: otherPort.iface,
        color,
        state: cable.link.state,
      };
    });

    return {
      name: device.node.name,
      rack: rack?.label ?? '',
      unit: device.node.graph.rack_unit,
      status: device.node.status,
      portHighlight: selectedPortKey
        ? layout.ports.get(selectedPortKey)?.iface ?? null
        : null,
      connections,
    };
  }, [layout, selectedDeviceName, selectedPortKey, visibleCables, rackMap]);

  return (
    <div className="relative w-full h-full overflow-hidden" style={{ background: 'var(--noc-topo-bg)' }}>
      {/* Zoomable/pannable container */}
      <div
        ref={containerRef}
        className="w-full h-full relative z-10 cursor-grab active:cursor-grabbing"
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onClick={(e) => {
          if (e.target === e.currentTarget) {
            handleBackgroundClick();
          }
        }}
      >
        {layout && (
          <svg
            style={{
              transform,
              transformOrigin: '0 0',
              position: 'absolute',
              width: layout.totalWidth,
              height: layout.totalHeight,
            }}
            viewBox={`0 0 ${layout.totalWidth} ${layout.totalHeight}`}
            xmlns="http://www.w3.org/2000/svg"
            onClick={(e) => {
              if (e.target === e.currentTarget) {
                handleBackgroundClick();
              }
            }}
          >
            {/* Data Centers */}
            {layout.dcs.map((dc) => (
              <DataCenter key={dc.id} dc={dc}>
                {layout.racks
                  .filter((r) => r.dcId === dc.id)
                  .map((rack) => (
                    <Rack
                      key={rack.id}
                      rack={rack}
                      uMarkers={layout.uMarkers.get(rack.id) ?? []}
                    />
                  ))}
              </DataCenter>
            ))}

            {/* Cables layer (behind devices on click) — drawn BEFORE devices */}
            <CableOverlay
              allCables={layout.cables}
              highlightedCableIds={highlightedCableIds}
              faultedCableIds={faultedCableIds}
              rackMap={rackMap}
              cableLaneBaseY={layout.cableLaneBaseY}
              totalWidth={layout.totalWidth}
              totalHeight={layout.totalHeight}
              selectedLinkId={selectedLinkId}
              onSelectLink={handleSelectLink}
              onContextMenuLink={handleContextMenu}
            />

            {/* External cables — dashed, grey, behind devices */}
            <g>
              {layout.externalCables.map((cable) => {
                const d = buildExternalCablePath(cable);
                const isHighlighted = selectedExternalName != null && (
                  cable.link.a.external === selectedExternalName ||
                  cable.link.z.external === selectedExternalName ||
                  cable.link.a.node === selectedExternalName ||
                  cable.link.z.node === selectedExternalName
                );
                const isNetHighlighted = selectedNetworkName != null && (
                  cable.link.a.network === selectedNetworkName ||
                  cable.link.z.network === selectedNetworkName
                );
                return (
                  <g key={cable.link.id}>
                    <path
                      d={d}
                      fill="none"
                      stroke={EXTERNAL_LINK_COLOR}
                      strokeWidth={(isHighlighted || isNetHighlighted) ? 2 : 1}
                      strokeDasharray="6,4"
                      opacity={(isHighlighted || isNetHighlighted) ? 0.8 : 0.45}
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </g>
                );
              })}
            </g>

            {/* Devices + embedded ports */}
            {Array.from(layout.devices.values()).map((device) => {
              const name = device.node.name;
              const selectionDimmed = hasSelection
                && !relatedDevices.has(name)
                && !faultedDevices.has(name);
              const searchDimmed = searchMatchedNodes !== null && !searchMatchedNodes.has(name);
              const isDimmed = selectionDimmed || searchDimmed;
              return (
                <DeviceFaceplate
                  key={name}
                  device={device}
                  ports={portsByDevice.get(name) ?? []}
                  selected={name === selectedDeviceName}
                  dimmed={isDimmed}
                  highlightedPorts={highlightedPorts}
                  cableColorMap={cableColorMap}
                  faultedPortColors={faultedPortColors}
                  onClick={() => handleDeviceClick(device.node)}
                  onPortClick={handlePortClick}
                />
              );
            })}

            {/* External networks — top (clouds) */}
            {Array.from(layout.externalNetworks.values())
              .filter((nl) => nl.position === 'top')
              .map((nl) => (
                <ExternalNetworkCloud
                  key={nl.network.name}
                  layout={nl}
                  selected={nl.network.name === selectedNetworkName}
                  dimmed={hasSelection && nl.network.name !== selectedNetworkName}
                  onClick={() => handleNetworkClick(nl.network)}
                />
              ))}

            {/* External networks — bottom (bars) */}
            {Array.from(layout.externalNetworks.values())
              .filter((nl) => nl.position === 'bottom')
              .map((nl) => (
                <ExternalNetworkBar
                  key={nl.network.name}
                  layout={nl}
                  selected={nl.network.name === selectedNetworkName}
                  dimmed={hasSelection && nl.network.name !== selectedNetworkName}
                  onClick={() => handleNetworkClick(nl.network)}
                />
              ))}

            {/* External nodes + DC-only clab nodes — services area and rack-placed */}
            {(() => {
              // Group services area external nodes by DC
              const servicesExtGroups = new Map<string, typeof layout.externalNodes extends Map<string, infer V> ? V[] : never>();
              for (const [, enl] of layout.externalNodes) {
                if (enl.servicesArea) {
                  const dc = enl.node.graph.dc || '_ungrouped';
                  if (!servicesExtGroups.has(dc)) servicesExtGroups.set(dc, []);
                  servicesExtGroups.get(dc)!.push(enl);
                }
              }

              // Group services area clab nodes (rackId === null) by DC
              const servicesClabGroups = new Map<string, DeviceLayout[]>();
              for (const [, device] of layout.devices) {
                if (device.rackId === null) {
                  const dc = device.node.graph.dc || '_ungrouped';
                  if (!servicesClabGroups.has(dc)) servicesClabGroups.set(dc, []);
                  servicesClabGroups.get(dc)!.push(device);
                }
              }

              // All DCs that have services area content
              const allServicesDCs = new Set<string>([
                ...servicesExtGroups.keys(),
                ...servicesClabGroups.keys(),
              ]);

              return (
                <>
                  {/* Services areas with grouped nodes */}
                  {Array.from(allServicesDCs).map((dc) => {
                    const extNodes = servicesExtGroups.get(dc) ?? [];
                    const clabDevices = servicesClabGroups.get(dc) ?? [];
                    const allItems = [
                      ...extNodes.map(n => ({ x: n.x, y: n.y, w: n.width })),
                      ...clabDevices.map(d => ({ x: d.x, y: d.y, w: d.width })),
                    ];
                    if (allItems.length === 0) return null;
                    const minX = Math.min(...allItems.map(n => n.x));
                    const maxX = Math.max(...allItems.map(n => n.x + n.w));
                    const minY = Math.min(...allItems.map(n => n.y));
                    return (
                      <ServicesArea
                        key={`services-${dc}`}
                        x={minX}
                        y={minY}
                        width={maxX - minX}
                        height={SERVICES_AREA_HEIGHT}
                      >
                        {extNodes.map((enl) => (
                          <ExternalDevice
                            key={enl.node.name}
                            layout={enl}
                            selected={enl.node.name === selectedExternalName}
                            dimmed={hasSelection && enl.node.name !== selectedExternalName}
                            onClick={() => handleExternalNodeClick(enl.node)}
                          />
                        ))}
                        {clabDevices.map((device) => {
                          const name = device.node.name;
                          const selectionDimmed = hasSelection
                            && !relatedDevices.has(name)
                            && !faultedDevices.has(name);
                          const searchDimmedDev = searchMatchedNodes !== null && !searchMatchedNodes.has(name);
                          return (
                            <DeviceFaceplate
                              key={name}
                              device={device}
                              ports={portsByDevice.get(name) ?? []}
                              selected={name === selectedDeviceName}
                              dimmed={selectionDimmed || searchDimmedDev}
                              highlightedPorts={highlightedPorts}
                              cableColorMap={cableColorMap}
                              faultedPortColors={faultedPortColors}
                              onClick={() => handleDeviceClick(device.node)}
                              onPortClick={handlePortClick}
                            />
                          );
                        })}
                      </ServicesArea>
                    );
                  })}

                  {/* Rack-placed external nodes */}
                  {Array.from(layout.externalNodes.values())
                    .filter((enl) => !enl.servicesArea)
                    .map((enl) => {
                      const searchDimmed = searchMatchedExternals !== null && !searchMatchedExternals.has(enl.node.name);
                      return (
                        <ExternalDevice
                          key={enl.node.name}
                          layout={enl}
                          selected={enl.node.name === selectedExternalName}
                          dimmed={(hasSelection && enl.node.name !== selectedExternalName) || searchDimmed}
                          onClick={() => handleExternalNodeClick(enl.node)}
                        />
                      );
                    })}
                </>
              );
            })()}
          </svg>
        )}
      </div>

      {/* Info panel — bottom overlay */}
      <div
        className="absolute bottom-5 left-1/2 -translate-x-1/2 z-20 font-mono text-[13px] px-6 py-3.5 max-w-[600px] transition-opacity duration-300 pointer-events-none"
        style={{
          background: 'var(--noc-info-bg)',
          border: '1px solid var(--noc-info-border)',
          borderRadius: '8px',
          color: infoPanelData ? 'var(--noc-info-text)' : 'var(--noc-device-name)',
        }}
      >
        {infoPanelData ? (
          <>
            <div className="font-semibold text-[15px]" style={{ color: 'var(--noc-info-heading)' }}>
              {infoPanelData.name}
              <span className="text-[11px] font-normal ml-2" style={{ color: 'var(--noc-info-muted)' }}>
                {infoPanelData.rack && `${infoPanelData.rack} U${infoPanelData.unit}`}
                {' '}
                {infoPanelData.status === 'stopped' || infoPanelData.status === 'error'
                  ? '● stopped'
                  : infoPanelData.status === 'warning'
                    ? '● warning'
                    : '● running'}
              </span>
              {infoPanelData.portHighlight && (
                <span className="text-[11px] font-normal ml-1" style={{ color: 'var(--noc-info-conn)' }}>
                  / port: <b>{infoPanelData.portHighlight}</b>
                </span>
              )}
            </div>
            {infoPanelData.connections.length > 0 && (
              <div className="mt-1.5 leading-relaxed" style={{ color: 'var(--noc-info-conn)' }}>
                {infoPanelData.connections.map((conn, i) => (
                  <div key={i}>
                    <span style={{ color: conn.color }}>■ {conn.myInterface}</span>
                    {' → '}
                    <b>{conn.otherDevice}</b>
                    {' : '}
                    <span style={{ color: conn.color }}>{conn.otherInterface}</span>
                    <span className="ml-1" style={{ color: 'var(--noc-info-badge)' }}>[{conn.state}]</span>
                  </div>
                ))}
              </div>
            )}
          </>
        ) : (
          'Click a device or cable to highlight'
        )}
      </div>

      {/* Layout warnings banner */}
      {topology && topology.warnings && topology.warnings.length > 0 && annotationErrors.length === 0 && (
        <div
          className="absolute top-3 left-1/2 -translate-x-1/2 z-20 font-mono text-[11px] px-4 py-2.5 max-w-[600px]"
          style={{
            background: 'var(--noc-warn-bg)',
            border: '1px solid var(--noc-warn-border)',
            borderRadius: '6px',
            color: 'var(--noc-warn-text)',
          }}
        >
          <div className="font-semibold text-[12px] mb-1" style={{ color: 'var(--noc-warn-title)' }}>
            .clabnoc.yml: {topology.warnings.length} warning{topology.warnings.length > 1 ? 's' : ''}
          </div>
          <div className="space-y-0.5" style={{ color: 'var(--noc-warn-muted)' }}>
            {topology.warnings.map((w, i) => (
              <div key={i}>{w}</div>
            ))}
          </div>
        </div>
      )}

      {/* Annotation error overlay */}
      {topology && annotationErrors.length > 0 && (
        <div className="absolute inset-0 flex items-center justify-center z-30" style={{ background: 'var(--noc-overlay-bg)' }}>
          <div
            className="max-w-[640px] w-full mx-4 font-mono text-sm"
            style={{
              background: 'var(--noc-info-bg)',
              border: '1px solid var(--noc-error-border)',
              borderRadius: '4px',
            }}
          >
            <div
              className="px-4 py-3 text-[13px] font-semibold"
              style={{ color: '#e74c3c', borderBottom: '1px solid var(--noc-error-divider)' }}
            >
              Topology cannot be displayed — missing node labels
            </div>
            <div className="px-4 py-3 text-[12px] leading-relaxed" style={{ color: 'var(--noc-info-conn)' }}>
              <p className="mb-3" style={{ color: 'var(--noc-device-name)' }}>
                All nodes require rack placement info. Use a{' '}
                <code style={{ color: 'var(--noc-error-highlight)' }}>.clabnoc.yml</code> config file
                (recommended) or{' '}
                <code style={{ color: 'var(--noc-error-highlight)' }}>graph-dc</code>,{' '}
                <code style={{ color: 'var(--noc-error-highlight)' }}>graph-rack</code>, and{' '}
                <code style={{ color: 'var(--noc-error-highlight)' }}>graph-rack-unit</code> labels in{' '}
                the Containerlab topology file.
              </p>
              <div
                className="overflow-y-auto"
                style={{ maxHeight: '240px' }}
              >
                <table className="w-full text-left" style={{ borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid var(--noc-table-border)' }}>
                      <th className="pb-1.5 pr-4" style={{ color: 'var(--noc-info-muted)' }}>Node</th>
                      <th className="pb-1.5" style={{ color: 'var(--noc-info-muted)' }}>Missing labels</th>
                    </tr>
                  </thead>
                  <tbody>
                    {annotationErrors.map((err) => (
                      <tr key={err.node} style={{ borderBottom: '1px solid var(--noc-table-divider)' }}>
                        <td className="py-1 pr-4" style={{ color: 'var(--noc-info-heading)' }}>{err.node}</td>
                        <td className="py-1">
                          {err.missing.map((label) => (
                            <code
                              key={label}
                              className="mr-2 px-1"
                              style={{
                                color: '#e74c3c',
                                background: 'var(--noc-error-code-bg)',
                                borderRadius: '2px',
                                fontSize: '11px',
                              }}
                            >
                              {label}
                            </code>
                          ))}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <p className="mt-3 text-[11px]" style={{ color: 'var(--noc-error-muted)' }}>
                Option 1: <code style={{ color: 'var(--noc-error-highlight)' }}>.clabnoc.yml</code> (recommended):
              </p>
              <pre
                className="mt-1 px-3 py-2 text-[11px] leading-relaxed overflow-x-auto"
                style={{ background: 'var(--noc-code-bg)', borderRadius: '3px', color: 'var(--noc-code-fg)' }}
              >
{`racks:
  rack-a01:
    dc: dc1
    units: 42
nodes:
  my-node:
    rack: rack-a01
    unit: 42
    size: 1`}
              </pre>
              <p className="mt-2 text-[11px]" style={{ color: 'var(--noc-error-muted)' }}>
                Option 2: Labels in <code style={{ color: 'var(--noc-error-highlight)' }}>.clab.yml</code>:
              </p>
              <pre
                className="mt-1 px-3 py-2 text-[11px] leading-relaxed overflow-x-auto"
                style={{ background: 'var(--noc-code-bg)', borderRadius: '3px', color: 'var(--noc-code-fg)' }}
              >
{`topology:
  nodes:
    my-node:
      labels:
        graph-dc: "dc1"
        graph-rack: "rack-a01"
        graph-rack-unit: "42"`}
              </pre>
            </div>
          </div>
        </div>
      )}

      {/* Empty state */}
      {!topology && (
        <div className="absolute inset-0 flex items-center justify-center z-20">
          <div className="text-center text-xs font-mono" style={{ color: 'var(--noc-empty-text)' }}>
            <div>No project selected</div>
            <div className="text-[10px] mt-1" style={{ color: 'var(--noc-empty-muted)' }}>
              Select a project to view topology
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
