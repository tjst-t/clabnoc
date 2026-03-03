import type { Topology, TopologyNode, TopologyLink, ExternalNode, ExternalNetwork, ExternalLink } from '../types/topology';

// ── Layout constants ──
export const U_HEIGHT = 18;
export const RACK_WIDTH = 190;
export const RACK_INNER_LEFT = 12;
export const RACK_INNER_WIDTH = 166;
export const RACK_TOP_MARGIN = 32;
export const RACK_U_TOTAL = 42;
export const RACK_GAP = 50;
export const DC_GAP = 120;

// Port dimensions (inside device faceplate)
export const PORT_W = 11;
export const PORT_H = 8;
export const PORT_GAP = 3;
export const PORT_TOP_PAD = 3;

// Cable lane base Y (above racks for inter-rack routing)
export const CABLE_LANE_BASE_Y = 50;
export const CABLE_LANE_SPACING = 6;

// External entity constants
export const SERVICES_AREA_HEIGHT = 40;
export const NETWORK_CLOUD_HEIGHT = 32;
export const NETWORK_BAR_HEIGHT = 24;
export const EXTERNAL_NODE_WIDTH = 120;
export const EXTERNAL_NODE_HEIGHT = 28;
export const EXTERNAL_NODE_GAP = 16;

// ── Role colors — hardware-inspired palette ──
export const ROLE_COLORS: Record<string, { border: string; bg: string; fill: string }> = {
  spine:  { border: '#2980b9', bg: 'rgba(41,128,185,0.08)', fill: '#2980b9' },
  leaf:   { border: '#2980b9', bg: 'rgba(41,128,185,0.08)', fill: '#2980b9' },
  router: { border: '#2980b9', bg: 'rgba(41,128,185,0.08)', fill: '#2980b9' },
  switch: { border: '#2980b9', bg: 'rgba(41,128,185,0.08)', fill: '#2980b9' },
  server: { border: '#27ae60', bg: 'rgba(39,174,96,0.06)', fill: '#27ae60' },
  bmc:    { border: '#8e44ad', bg: 'rgba(142,68,173,0.08)', fill: '#8e44ad' },
  host:   { border: '#27ae60', bg: 'rgba(39,174,96,0.06)', fill: '#27ae60' },
};

export const EXTERNAL_NODE_COLOR = { border: '#7f8c8d', bg: 'rgba(127,140,141,0.06)', fill: '#7f8c8d' };

export const STATUS_COLORS: Record<string, string> = {
  running: '#2ecc71',
  stopped: '#e74c3c',
  error: '#e74c3c',
  warning: '#f39c12',
  ok: '#2ecc71',
};

export const LINK_STATE_COLORS: Record<string, string> = {
  up: '#3498db',
  down: '#e74c3c',
  degraded: '#e67e22',
};

export const EXTERNAL_LINK_COLOR = '#95a5a6';

// ── Layout result types ──

export interface DCLayout {
  id: string;
  label: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface RackLayout {
  id: string;
  label: string;
  dcId: string;
  x: number;
  y: number;
  width: number;
  height: number;
}

export interface UMarker {
  unit: number;
  y: number; // absolute Y
}

export interface DeviceLayout {
  node: TopologyNode;
  x: number; // absolute position of the device rect
  y: number;
  width: number;
  height: number;
  unitSize: number; // occupied U count (default 1)
  rackId: string | null;
  role: string;
}

export interface PortLayout {
  key: string; // "node:interface"
  node: TopologyNode;
  iface: string;
  cx: number; // center X (absolute)
  cy: number; // center Y (absolute)
  x: number;  // top-left X (absolute)
  y: number;  // top-left Y (absolute)
  w: number;
  h: number;
  role: string;
  state: string;
}

export interface CableLayout {
  link: TopologyLink;
  aPort: PortLayout;
  zPort: PortLayout;
  intraRack: boolean;
}

// ── External layout types ──

export interface ExternalNodeLayout {
  node: ExternalNode;
  x: number;
  y: number;
  width: number;
  height: number;
  rackId: string | null;
  servicesArea: boolean; // true = DC-level services area, false = rack-placed
}

export interface ExternalNetworkLayout {
  network: ExternalNetwork;
  x: number;
  y: number;
  width: number;
  height: number;
  position: 'top' | 'bottom';
}

export interface ExternalCableLayout {
  link: ExternalLink;
  ax: number;
  ay: number;
  zx: number;
  zy: number;
}

export interface LayoutResult {
  dcs: DCLayout[];
  racks: RackLayout[];
  uMarkers: Map<string, UMarker[]>; // key: rackId
  devices: Map<string, DeviceLayout>; // key: node name
  ports: Map<string, PortLayout>; // key: "node:interface"
  cables: CableLayout[];
  externalNodes: Map<string, ExternalNodeLayout>;
  externalNetworks: Map<string, ExternalNetworkLayout>;
  externalCables: ExternalCableLayout[];
  totalWidth: number;
  totalHeight: number;
}

/** Port key convention */
export function portKey(nodeName: string, iface: string): string {
  return `${nodeName}:${iface}`;
}

/**
 * Compute absolute layout positions for all topology elements.
 * Ports are positioned INSIDE the device faceplate as a row of small rectangles.
 */
export function computeLayout(topo: Topology): LayoutResult {
  const dcs: DCLayout[] = [];
  const racks: RackLayout[] = [];
  const uMarkersMap = new Map<string, UMarker[]>();
  const devices = new Map<string, DeviceLayout>();
  const ports = new Map<string, PortLayout>();
  const cables: CableLayout[] = [];
  const externalNodes = new Map<string, ExternalNodeLayout>();
  const externalNetworks = new Map<string, ExternalNetworkLayout>();
  const externalCables: ExternalCableLayout[] = [];

  // ── Pre-compute external entity sizes ──
  const topNetworks = (topo.external_networks ?? []).filter(n => n.position === 'top');
  const bottomNetworks = (topo.external_networks ?? []).filter(n => n.position === 'bottom');

  // DC-only external nodes (services area) per DC
  const servicesNodesByDC = new Map<string, ExternalNode[]>();
  const rackPlacedExternals: ExternalNode[] = [];
  for (const en of topo.external_nodes ?? []) {
    if (en.graph.rack) {
      rackPlacedExternals.push(en);
    } else if (en.graph.dc) {
      const list = servicesNodesByDC.get(en.graph.dc) ?? [];
      list.push(en);
      servicesNodesByDC.set(en.graph.dc, list);
    }
  }

  const hasTopNetworks = topNetworks.length > 0;
  const topNetworkPadding = hasTopNetworks ? NETWORK_CLOUD_HEIGHT + 20 : 0;

  // ── 1. Collect all interfaces per node from links ──
  const nodeInterfaces = new Map<string, string[]>();
  const portStates = new Map<string, string>();

  for (const link of topo.links) {
    for (const ep of [link.a, link.z]) {
      if (!nodeInterfaces.has(ep.node)) {
        nodeInterfaces.set(ep.node, []);
      }
      const list = nodeInterfaces.get(ep.node)!;
      if (!list.includes(ep.interface)) {
        list.push(ep.interface);
      }
    }
    portStates.set(portKey(link.a.node, link.a.interface), link.state);
    portStates.set(portKey(link.z.node, link.z.interface), link.state);
  }
  for (const [, list] of nodeInterfaces) {
    list.sort();
  }

  // ── 2. DC + rack structure ──
  let dcXOffset = 0;
  const rackPositions = new Map<string, { x: number; y: number }>();

  // We'll place racks in a top margin area for cable lanes
  const topPadding = 80 + topNetworkPadding; // extra space above racks for cable lane routing + top networks

  // Helper to get per-rack U count (from Groups.rack_units or default)
  const getRackUnits = (rackName: string): number => {
    return topo.groups.rack_units?.[rackName] ?? RACK_U_TOTAL;
  };

  // Store per-rack U count for device layout calculation
  const rackUnitCounts = new Map<string, number>();

  for (const dc of topo.groups.dcs) {
    const dcRacks = topo.groups.racks[dc] ?? [];

    // Check for services area and bottom networks/bars in this DC
    const dcServices = servicesNodesByDC.get(dc) ?? [];
    const hasServicesArea = dcServices.length > 0;
    const servicesAreaH = hasServicesArea ? SERVICES_AREA_HEIGHT + 10 : 0;

    // Bottom networks within this DC
    const dcBottomNets = bottomNetworks.filter(n => n.dc === dc || !n.dc);
    const bottomNetH = dcBottomNets.length > 0 ? (NETWORK_BAR_HEIGHT + 8) * dcBottomNets.length : 0;

    // Find tallest rack in this DC to size the DC box and bottom-align racks
    const tallestRackH = dcRacks.reduce((max, rack) => {
      const h = getRackUnits(rack) * U_HEIGHT + RACK_TOP_MARGIN + 10;
      return h > max ? h : max;
    }, 0);

    for (let ri = 0; ri < dcRacks.length; ri++) {
      const rack = dcRacks[ri]!;
      const rackId = `rack:${dc}:${rack}`;
      const rackX = dcXOffset + 30 + ri * (RACK_WIDTH + RACK_GAP);

      const rackUnits = getRackUnits(rack);
      rackUnitCounts.set(rackId, rackUnits);
      const rackH = rackUnits * U_HEIGHT + RACK_TOP_MARGIN + 10;

      // Bottom-align: shorter racks are pushed down so bottoms line up
      const rackY = topPadding + servicesAreaH + (tallestRackH - rackH);
      rackPositions.set(rackId, { x: rackX, y: rackY });

      racks.push({
        id: rackId,
        label: rack,
        dcId: `dc:${dc}`,
        x: rackX,
        y: rackY,
        width: RACK_WIDTH,
        height: rackH,
      });

      // U-position ruler markers (every 3 units)
      const markers: UMarker[] = [];
      for (let u = 1; u <= rackUnits; u += 3) {
        markers.push({
          unit: u,
          y: rackY + RACK_TOP_MARGIN + (rackUnits - u) * U_HEIGHT,
        });
      }
      uMarkersMap.set(rackId, markers);
    }

    const dcWidth = dcRacks.length > 0
      ? dcRacks.length * RACK_WIDTH + (dcRacks.length - 1) * RACK_GAP + 60
      : RACK_WIDTH + 60;
    const dcHeight = tallestRackH + topPadding + servicesAreaH + bottomNetH + 20;

    dcs.push({
      id: `dc:${dc}`,
      label: dc,
      x: dcXOffset,
      y: topNetworkPadding,
      width: dcWidth,
      height: dcHeight,
    });

    // ── Services area: DC-only external nodes ──
    if (hasServicesArea) {
      const servicesStartX = dcXOffset + 30;
      const servicesY = topPadding;

      for (let si = 0; si < dcServices.length; si++) {
        const en = dcServices[si]!;
        const enX = servicesStartX + si * (EXTERNAL_NODE_WIDTH + EXTERNAL_NODE_GAP);
        const enY = servicesY;

        externalNodes.set(en.name, {
          node: en,
          x: enX,
          y: enY,
          width: EXTERNAL_NODE_WIDTH,
          height: EXTERNAL_NODE_HEIGHT,
          rackId: null,
          servicesArea: true,
        });
      }
    }

    // ── Bottom networks within DC ──
    const bottomNetStartY = topPadding + servicesAreaH + tallestRackH + 8;
    for (let ni = 0; ni < dcBottomNets.length; ni++) {
      const net = dcBottomNets[ni]!;
      const netX = dcXOffset + 30;
      const netY = bottomNetStartY + ni * (NETWORK_BAR_HEIGHT + 8);
      const netW = dcWidth - 60;

      externalNetworks.set(net.name, {
        network: net,
        x: netX,
        y: netY,
        width: netW,
        height: NETWORK_BAR_HEIGHT,
        position: 'bottom',
      });
    }

    dcXOffset += dcWidth + DC_GAP;
  }

  // ── Top networks (above DC boxes) ──
  if (hasTopNetworks) {
    const totalDCWidth = dcs.reduce((max, dc) => Math.max(max, dc.x + dc.width), 0);
    const cloudWidth = Math.min(totalDCWidth, 250);
    for (let ni = 0; ni < topNetworks.length; ni++) {
      const net = topNetworks[ni]!;
      const netX = ni * (cloudWidth + 30);
      const netY = 4;

      externalNetworks.set(net.name, {
        network: net,
        x: netX,
        y: netY,
        width: cloudWidth,
        height: NETWORK_CLOUD_HEIGHT,
        position: 'top',
      });
    }
  }

  // ── 3. Collect nodes per rack_unit ──
  const rackUnitNodes = new Map<string, Map<number, TopologyNode[]>>();
  const ungroupedNodes: TopologyNode[] = [];
  const nodeRack = new Map<string, string>();

  for (const node of topo.nodes) {
    const g = node.graph;
    if (g.dc && g.rack) {
      const rackId = `rack:${g.dc}:${g.rack}`;
      nodeRack.set(node.name, rackId);
      if (!rackUnitNodes.has(rackId)) rackUnitNodes.set(rackId, new Map());
      const unitMap = rackUnitNodes.get(rackId)!;
      const unit = g.rack_unit || 1;
      if (!unitMap.has(unit)) unitMap.set(unit, []);
      unitMap.get(unit)!.push(node);
    } else {
      ungroupedNodes.push(node);
    }
  }

  // ── 4. Device layout + embedded ports ──
  for (const node of topo.nodes) {
    const g = node.graph;
    const role = g.role || 'host';
    let rackId: string | null = null;
    let deviceX: number;
    let deviceY: number;
    let deviceW: number;
    let deviceH: number;

    if (g.dc && g.rack) {
      rackId = `rack:${g.dc}:${g.rack}`;
      const rp = rackPositions.get(rackId);
      if (!rp) continue;

      const unit = g.rack_unit || 1;
      const unitSize = g.rack_unit_size || 1;
      const rackUnits = rackUnitCounts.get(rackId) ?? RACK_U_TOTAL;
      // Multi-U devices: height spans unitSize * U_HEIGHT
      // rack_unit is the bottom U of the device, so top = rackUnits - (unit + unitSize - 1)
      const topU = unit + unitSize - 1;
      const sameUnitNodes = rackUnitNodes.get(rackId)?.get(unit) ?? [];
      const idx = sameUnitNodes.indexOf(node);

      deviceX = rp.x + RACK_INNER_LEFT;
      deviceY = rp.y + RACK_TOP_MARGIN + (rackUnits - topU) * U_HEIGHT + idx * U_HEIGHT;
      deviceW = RACK_INNER_WIDTH;
      deviceH = U_HEIGHT * unitSize;
    } else {
      deviceX = dcXOffset + 30;
      deviceY = topPadding + ungroupedNodes.indexOf(node) * (U_HEIGHT + 4);
      deviceW = RACK_INNER_WIDTH;
      deviceH = U_HEIGHT;
    }

    const unitSize = g.rack_unit_size || 1;
    devices.set(node.name, {
      node,
      x: deviceX,
      y: deviceY,
      width: deviceW,
      height: deviceH,
      unitSize,
      rackId,
      role,
    });

    // Ports INSIDE the device (row at top-right area of faceplate)
    const interfaces = nodeInterfaces.get(node.name);
    if (interfaces && interfaces.length > 0) {
      const totalPortsWidth = interfaces.length * (PORT_W + PORT_GAP) - PORT_GAP;
      const startX = deviceX + deviceW - totalPortsWidth - 6;

      for (let i = 0; i < interfaces.length; i++) {
        const iface = interfaces[i]!;
        const pk = portKey(node.name, iface);
        const state = portStates.get(pk) ?? 'up';

        const px = startX + i * (PORT_W + PORT_GAP);
        // Center ports vertically within multi-U devices
        const py = deviceY + Math.floor((deviceH - PORT_H) / 2);

        ports.set(pk, {
          key: pk,
          node,
          iface,
          x: px,
          y: py,
          w: PORT_W,
          h: PORT_H,
          cx: px + PORT_W / 2,
          cy: py + PORT_H / 2,
          role,
          state,
        });
      }
    }
  }

  // ── Rack-placed external nodes ──
  for (const en of rackPlacedExternals) {
    const g = en.graph;
    if (!g.dc || !g.rack) continue;
    const rackId = `rack:${g.dc}:${g.rack}`;
    const rp = rackPositions.get(rackId);
    if (!rp) continue;

    const unit = g.rack_unit || 1;
    const unitSize = g.rack_unit_size || 1;
    const rackUnits = rackUnitCounts.get(rackId) ?? RACK_U_TOTAL;
    const topU = unit + unitSize - 1;

    const enX = rp.x + RACK_INNER_LEFT;
    const enY = rp.y + RACK_TOP_MARGIN + (rackUnits - topU) * U_HEIGHT;
    const enW = RACK_INNER_WIDTH;
    const enH = U_HEIGHT * unitSize;

    externalNodes.set(en.name, {
      node: en,
      x: enX,
      y: enY,
      width: enW,
      height: enH,
      rackId,
      servicesArea: false,
    });
  }

  // ── 5. Cables (clab links) ──
  for (const link of topo.links) {
    const aPort = ports.get(portKey(link.a.node, link.a.interface));
    const zPort = ports.get(portKey(link.z.node, link.z.interface));
    if (!aPort || !zPort) continue;

    const sameRack =
      nodeRack.get(link.a.node) === nodeRack.get(link.z.node) &&
      nodeRack.has(link.a.node);

    cables.push({ link, aPort, zPort, intraRack: sameRack });
  }

  // ── 6. External cables ──
  for (const el of topo.external_links ?? []) {
    const aPos = resolveEndpointPosition(el.a, devices, externalNodes, externalNetworks);
    const zPos = resolveEndpointPosition(el.z, devices, externalNodes, externalNetworks);
    if (!aPos || !zPos) continue;

    externalCables.push({
      link: el,
      ax: aPos.x,
      ay: aPos.y,
      zx: zPos.x,
      zy: zPos.y,
    });
  }

  // ── Total dimensions ──
  let totalWidth = 0;
  let totalHeight = 0;
  for (const dc of dcs) {
    totalWidth = Math.max(totalWidth, dc.x + dc.width);
    totalHeight = Math.max(totalHeight, dc.y + dc.height);
  }
  // Account for top networks
  for (const [, nl] of externalNetworks) {
    totalWidth = Math.max(totalWidth, nl.x + nl.width);
    totalHeight = Math.max(totalHeight, nl.y + nl.height);
  }
  totalWidth += 40;
  totalHeight += 40;

  return {
    dcs, racks, uMarkers: uMarkersMap, devices, ports, cables,
    externalNodes, externalNetworks, externalCables,
    totalWidth, totalHeight,
  };
}

/** Resolve the center position for an external link endpoint. */
function resolveEndpointPosition(
  ep: { node?: string; external?: string; network?: string },
  devices: Map<string, DeviceLayout>,
  extNodes: Map<string, ExternalNodeLayout>,
  extNets: Map<string, ExternalNetworkLayout>,
): { x: number; y: number } | null {
  if (ep.node) {
    const dev = devices.get(ep.node);
    if (dev) return { x: dev.x + dev.width / 2, y: dev.y + dev.height / 2 };
  }
  if (ep.external) {
    const en = extNodes.get(ep.external);
    if (en) return { x: en.x + en.width / 2, y: en.y + en.height / 2 };
  }
  if (ep.network) {
    const net = extNets.get(ep.network);
    if (net) {
      // For top networks, connect from bottom center; for bottom, top center
      if (net.position === 'top') {
        return { x: net.x + net.width / 2, y: net.y + net.height };
      }
      return { x: net.x + net.width / 2, y: net.y };
    }
  }
  return null;
}

/**
 * Build orthogonal cable path between two ports (for highlighted / faulted cables).
 * Same rack: route out right side. Different racks: go up to lane, across, down.
 * Stagger is compressed (2px intra-rack, 3px inter-rack) for dense layouts.
 */
export function buildCablePath(
  cable: CableLayout,
  cableIndex: number,
  rackMap: Map<string, RackLayout>
): string {
  const x1 = cable.aPort.cx;
  const y1 = cable.aPort.cy;
  const x2 = cable.zPort.cx;
  const y2 = cable.zPort.cy;

  if (cable.intraRack) {
    // Same rack: route out the right side and loop back
    const aRackId = Array.from(rackMap.values()).find((r) => {
      return x1 >= r.x && x1 <= r.x + r.width;
    });
    const exitX = (aRackId ? aRackId.x + aRackId.width : Math.max(x1, x2)) + 8 + cableIndex * 2;
    return `M${x1},${y1} L${exitX},${y1} L${exitX},${y2} L${x2},${y2}`;
  }

  // Different racks: go up to cable lane, across, then down
  const laneY = CABLE_LANE_BASE_Y - cableIndex * 3;
  return `M${x1},${y1} L${x1},${laneY} L${x2},${laneY} L${x2},${y2}`;
}

/** Build a simple line path for external cables (dashed). */
export function buildExternalCablePath(cable: ExternalCableLayout): string {
  return `M${cable.ax},${cable.ay} L${cable.zx},${cable.zy}`;
}
