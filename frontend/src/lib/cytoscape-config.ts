import type { StylesheetStyle, LayoutOptions } from 'cytoscape';

type Stylesheet = StylesheetStyle;

const ROLE_COLORS: Record<string, string> = {
  spine: '#00c8ff',
  leaf: '#00d4aa',
  router: '#00c8ff',
  switch: '#00d4aa',
  server: '#ffb020',
  bmc: '#ff6b8a',
  host: '#8a8aa0',
};

const LINK_STATE_COLORS: Record<string, string> = {
  up: '#1a6a4a',
  down: '#ff3b5c',
  degraded: '#ffb020',
};

export const cytoscapeStylesheet: Stylesheet[] = [
  // --- DC compound node ---
  {
    selector: 'node[type="dc"]',
    style: {
      'background-color': '#0d1520',
      'background-opacity': 0.6,
      'border-color': '#2a3a4e',
      'border-width': 1,
      'border-opacity': 0.8,
      'label': 'data(label)',
      'text-valign': 'top',
      'text-halign': 'center',
      'font-family': '"JetBrains Mono", monospace',
      'font-size': '11px',
      'font-weight': 600,
      'color': '#5a6a7e',
      'text-margin-y': -6,
      'text-transform': 'uppercase' as never,
      'padding': '24px',
      'shape': 'rectangle',
      'compound-sizing-wrt-labels': 'exclude',
    },
  },
  // --- Rack compound node ---
  {
    selector: 'node[type="rack"]',
    style: {
      'background-color': '#111820',
      'background-opacity': 0.5,
      'border-color': '#2a3a4e',
      'border-width': 1,
      'border-style': 'dashed',
      'label': 'data(label)',
      'text-valign': 'top',
      'text-halign': 'center',
      'font-family': '"JetBrains Mono", monospace',
      'font-size': '9px',
      'font-weight': 400,
      'color': '#3d4e62',
      'text-margin-y': -4,
      'text-transform': 'uppercase' as never,
      'padding': '16px',
      'shape': 'rectangle',
      'compound-sizing-wrt-labels': 'exclude',
    },
  },
  // --- Device node ---
  {
    selector: 'node[type="device"]',
    style: {
      'background-color': '#161d27',
      'border-color': '#2a4060',
      'border-width': 2,
      'label': 'data(label)',
      'text-valign': 'bottom',
      'text-halign': 'center',
      'font-family': '"JetBrains Mono", monospace',
      'font-size': '10px',
      'font-weight': 500,
      'color': '#c5cdd8',
      'text-margin-y': 6,
      'width': 44,
      'height': 44,
      'shape': 'rectangle',
    },
  },
  // Role-specific colors
  ...Object.entries(ROLE_COLORS).map(([role, color]) => ({
    selector: `node[role="${role}"]` as never,
    style: {
      'border-color': color,
      'color': color,
    },
  })),
  // Status indicators
  {
    selector: 'node[status="running"]',
    style: {
      'background-color': '#0d1a14',
    },
  },
  {
    selector: 'node[status="stopped"]',
    style: {
      'background-color': '#1a0d10',
      'border-color': '#7a1a2e',
      'border-style': 'dashed',
    },
  },
  // Selected node
  {
    selector: 'node[type="device"]:selected',
    style: {
      'border-width': 3,
      'border-color': '#00d4aa',
      'background-color': '#0a2a20',
      'overlay-opacity': 0,
    },
  },
  // --- Edges ---
  {
    selector: 'edge',
    style: {
      'curve-style': 'bezier',
      'width': 2,
      'line-color': LINK_STATE_COLORS.up,
      'target-arrow-shape': 'none',
      'opacity': 0.7,
      'label': 'data(label)',
      'font-family': '"JetBrains Mono", monospace',
      'font-size': '7px',
      'color': '#3d4e62',
      'text-rotation': 'autorotate',
      'text-margin-y': -8,
    },
  },
  {
    selector: 'edge[state="down"]',
    style: {
      'line-color': LINK_STATE_COLORS.down,
      'line-style': 'dashed',
      'opacity': 0.9,
    },
  },
  {
    selector: 'edge[state="degraded"]',
    style: {
      'line-color': LINK_STATE_COLORS.degraded,
      'line-style': 'dotted',
      'opacity': 0.9,
    },
  },
  {
    selector: 'edge:selected',
    style: {
      'line-color': '#00d4aa',
      'opacity': 1,
      'width': 3,
      'overlay-opacity': 0,
    },
  },
];

export const defaultLayout: LayoutOptions = {
  name: 'cose',
  animate: false,
  nodeDimensionsIncludeLabels: true,
  idealEdgeLength: () => 120,
  nodeRepulsion: () => 8000,
  gravity: 0.3,
  padding: 40,
};
