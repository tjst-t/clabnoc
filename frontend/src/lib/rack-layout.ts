import type { Topology, TopologyNode, TopologyLink } from '../types/topology';

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

export interface LayoutResult {
  dcs: DCLayout[];
  racks: RackLayout[];
  uMarkers: Map<string, UMarker[]>; // key: rackId
  devices: Map<string, DeviceLayout>; // key: node name
  ports: Map<string, PortLayout>; // key: "node:interface"
  cables: CableLayout[];
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
  const topPadding = 80; // extra space above racks for cable lane routing

  // Helper to get per-rack U count (from Groups.rack_units or default)
  const getRackUnits = (rackName: string): number => {
    return topo.groups.rack_units?.[rackName] ?? RACK_U_TOTAL;
  };

  // Store per-rack U count for device layout calculation
  const rackUnitCounts = new Map<string, number>();

  for (const dc of topo.groups.dcs) {
    const dcRacks = topo.groups.racks[dc] ?? [];

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
      const rackY = topPadding + (tallestRackH - rackH);
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
    const dcHeight = tallestRackH + topPadding + 20;

    dcs.push({
      id: `dc:${dc}`,
      label: dc,
      x: dcXOffset,
      y: 0,
      width: dcWidth,
      height: dcHeight,
    });

    dcXOffset += dcWidth + DC_GAP;
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

  // ── 5. Cables ──
  for (const link of topo.links) {
    const aPort = ports.get(portKey(link.a.node, link.a.interface));
    const zPort = ports.get(portKey(link.z.node, link.z.interface));
    if (!aPort || !zPort) continue;

    const sameRack =
      nodeRack.get(link.a.node) === nodeRack.get(link.z.node) &&
      nodeRack.has(link.a.node);

    cables.push({ link, aPort, zPort, intraRack: sameRack });
  }

  // ── Total dimensions ──
  let totalWidth = 0;
  let totalHeight = 0;
  for (const dc of dcs) {
    totalWidth = Math.max(totalWidth, dc.x + dc.width);
    totalHeight = Math.max(totalHeight, dc.y + dc.height);
  }
  totalWidth += 40;
  totalHeight += 40;

  return { dcs, racks, uMarkers: uMarkersMap, devices, ports, cables, totalWidth, totalHeight };
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

/**
 * Build a bezier-curve cable path for background (always-visible) cables.
 * Uses a gentle curve from port to port — no stagger needed, cables naturally
 * overlap to form a "cable bundle" look at low opacity.
 */
export function buildDirectCablePath(cable: CableLayout): string {
  const x1 = cable.aPort.cx;
  const y1 = cable.aPort.cy;
  const x2 = cable.zPort.cx;
  const y2 = cable.zPort.cy;

  if (cable.intraRack) {
    // Intra-rack: curve out to the right side
    const bulge = 30 + Math.abs(y2 - y1) * 0.15;
    const cpx = Math.max(x1, x2) + bulge;
    return `M${x1},${y1} C${cpx},${y1} ${cpx},${y2} ${x2},${y2}`;
  }

  // Inter-rack: curve up through the cable lane area
  const midY = CABLE_LANE_BASE_Y;
  return `M${x1},${y1} C${x1},${midY} ${x2},${midY} ${x2},${y2}`;
}
