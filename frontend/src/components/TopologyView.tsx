import { useMemo, useEffect, useCallback, useState } from 'react';
import type { Topology, TopologyNode, TopologyLink } from '../types/topology';
import {
  computeLayout,
  LINK_STATE_COLORS,
  type CableLayout,
  type PortLayout,
  type RackLayout,
} from '../lib/rack-layout';
import { useZoomPan } from '../hooks/useZoomPan';
import { DataCenter } from './topology/DataCenter';
import { Rack } from './topology/Rack';
import { DeviceFaceplate } from './topology/DeviceFaceplate';
import { CableOverlay } from './topology/CableOverlay';

interface Props {
  topology: Topology | null;
  onSelectNode: (node: TopologyNode | null) => void;
  onSelectLink: (link: TopologyLink | null) => void;
  onContextMenuLink: (link: TopologyLink, x: number, y: number) => void;
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
    if (!node.graph.rack) missing.push('graph-rack');
    if (!node.graph.rack_unit) missing.push('graph-rack-unit');
    if (missing.length > 0) {
      errors.push({ node: node.name, missing });
    }
  }
  return errors;
}

export function TopologyView({ topology, onSelectNode, onSelectLink, onContextMenuLink }: Props) {
  // Selection state: either a device or a specific port
  const [selectedDeviceName, setSelectedDeviceName] = useState<string | null>(null);
  const [selectedPortKey, setSelectedPortKey] = useState<string | null>(null);
  const [selectedLinkId, setSelectedLinkId] = useState<string | null>(null);

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

  const hasSelection = selectedDeviceName !== null || selectedLinkId !== null;

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

  const handleBackgroundClick = useCallback(() => {
    setSelectedDeviceName(null);
    setSelectedPortKey(null);
    setSelectedLinkId(null);
    onSelectNode(null);
    onSelectLink(null);
  }, [onSelectNode, onSelectLink]);

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
    <div className="relative w-full h-full bg-[#0a0e17] overflow-hidden">
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
              totalWidth={layout.totalWidth}
              totalHeight={layout.totalHeight}
              selectedLinkId={selectedLinkId}
              onSelectLink={handleSelectLink}
              onContextMenuLink={handleContextMenu}
            />

            {/* Devices + embedded ports */}
            {Array.from(layout.devices.values()).map((device) => {
              const name = device.node.name;
              const isDimmed = hasSelection
                && !relatedDevices.has(name)
                && !faultedDevices.has(name);
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
          </svg>
        )}
      </div>

      {/* Info panel — bottom overlay */}
      <div
        className="absolute bottom-5 left-1/2 -translate-x-1/2 z-20 font-mono text-[13px] px-6 py-3.5 max-w-[600px] transition-opacity duration-300 pointer-events-none"
        style={{
          background: '#151c2c',
          border: '1px solid #2a3a5c',
          borderRadius: '8px',
          color: infoPanelData ? '#c8d6e5' : '#8899aa',
        }}
      >
        {infoPanelData ? (
          <>
            <div className="font-semibold text-[15px]" style={{ color: '#7ec8e3' }}>
              {infoPanelData.name}
              <span className="text-[11px] font-normal ml-2" style={{ color: '#5a7a9a' }}>
                {infoPanelData.rack && `${infoPanelData.rack} U${infoPanelData.unit}`}
                {' '}
                {infoPanelData.status === 'stopped' || infoPanelData.status === 'error'
                  ? '● stopped'
                  : infoPanelData.status === 'warning'
                    ? '● warning'
                    : '● running'}
              </span>
              {infoPanelData.portHighlight && (
                <span className="text-[11px] font-normal ml-1" style={{ color: '#a0b0c0' }}>
                  / port: <b>{infoPanelData.portHighlight}</b>
                </span>
              )}
            </div>
            {infoPanelData.connections.length > 0 && (
              <div className="mt-1.5 leading-relaxed" style={{ color: '#a0b0c0' }}>
                {infoPanelData.connections.map((conn, i) => (
                  <div key={i}>
                    <span style={{ color: conn.color }}>■ {conn.myInterface}</span>
                    {' → '}
                    <b>{conn.otherDevice}</b>
                    {' : '}
                    <span style={{ color: conn.color }}>{conn.otherInterface}</span>
                    <span className="ml-1" style={{ color: '#4a5a6a' }}>[{conn.state}]</span>
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
            background: '#1c1a10',
            border: '1px solid #5a4a18',
            borderRadius: '6px',
            color: '#d4a017',
          }}
        >
          <div className="font-semibold text-[12px] mb-1" style={{ color: '#e6b422' }}>
            .clabnoc.yml: {topology.warnings.length} warning{topology.warnings.length > 1 ? 's' : ''}
          </div>
          <div className="space-y-0.5" style={{ color: '#b89a30' }}>
            {topology.warnings.map((w, i) => (
              <div key={i}>{w}</div>
            ))}
          </div>
        </div>
      )}

      {/* Annotation error overlay */}
      {topology && annotationErrors.length > 0 && (
        <div className="absolute inset-0 flex items-center justify-center z-30 bg-[#0a0e17]/90">
          <div
            className="max-w-[640px] w-full mx-4 font-mono text-sm"
            style={{
              background: '#151c2c',
              border: '1px solid #5a1828',
              borderRadius: '4px',
            }}
          >
            <div
              className="px-4 py-3 text-[13px] font-semibold"
              style={{ color: '#e74c3c', borderBottom: '1px solid #2a1828' }}
            >
              Topology cannot be displayed — missing node labels
            </div>
            <div className="px-4 py-3 text-[12px] leading-relaxed" style={{ color: '#a0b0c0' }}>
              <p className="mb-3" style={{ color: '#8899aa' }}>
                All nodes require rack placement info. Use a{' '}
                <code style={{ color: '#e67e22' }}>.clabnoc.yml</code> config file
                (recommended) or{' '}
                <code style={{ color: '#e67e22' }}>graph-dc</code>,{' '}
                <code style={{ color: '#e67e22' }}>graph-rack</code>, and{' '}
                <code style={{ color: '#e67e22' }}>graph-rack-unit</code> labels in{' '}
                the Containerlab topology file.
              </p>
              <div
                className="overflow-y-auto"
                style={{ maxHeight: '240px' }}
              >
                <table className="w-full text-left" style={{ borderCollapse: 'collapse' }}>
                  <thead>
                    <tr style={{ borderBottom: '1px solid #2a3a5c' }}>
                      <th className="pb-1.5 pr-4" style={{ color: '#5a7a9a' }}>Node</th>
                      <th className="pb-1.5" style={{ color: '#5a7a9a' }}>Missing labels</th>
                    </tr>
                  </thead>
                  <tbody>
                    {annotationErrors.map((err) => (
                      <tr key={err.node} style={{ borderBottom: '1px solid #1a2a3c' }}>
                        <td className="py-1 pr-4" style={{ color: '#7ec8e3' }}>{err.node}</td>
                        <td className="py-1">
                          {err.missing.map((label) => (
                            <code
                              key={label}
                              className="mr-2 px-1"
                              style={{
                                color: '#e74c3c',
                                background: '#2a0c10',
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
              <p className="mt-3 text-[11px]" style={{ color: '#5a6a7a' }}>
                Option 1: <code style={{ color: '#e67e22' }}>.clabnoc.yml</code> (recommended):
              </p>
              <pre
                className="mt-1 px-3 py-2 text-[11px] leading-relaxed overflow-x-auto"
                style={{ background: '#0d1219', borderRadius: '3px', color: '#8899aa' }}
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
              <p className="mt-2 text-[11px]" style={{ color: '#5a6a7a' }}>
                Option 2: Labels in <code style={{ color: '#e67e22' }}>.clab.yml</code>:
              </p>
              <pre
                className="mt-1 px-3 py-2 text-[11px] leading-relaxed overflow-x-auto"
                style={{ background: '#0d1219', borderRadius: '3px', color: '#8899aa' }}
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
          <div className="text-center text-[#6b7b8d] text-xs font-mono">
            <div>No project selected</div>
            <div className="text-[10px] mt-1" style={{ color: '#4a5a6a' }}>
              Select a project to view topology
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
