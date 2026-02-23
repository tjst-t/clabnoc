import cytoscape from 'cytoscape';
import type { TopologyData } from '../types/topology';

// Build Cytoscape elements from topology data
// - Create compound nodes for DC and Rack (parent nodes)
// - Create regular nodes for each topology node (with parent=rack or dc)
// - Create edges for links
// - Hide nodes with graph.hidden=true
export function buildElements(topo: TopologyData): cytoscape.ElementDefinition[] {
  const elements: cytoscape.ElementDefinition[] = [];

  // Add DC compound nodes
  for (const dc of (topo.groups.dcs ?? [])) {
    elements.push({ data: { id: `dc:${dc}`, label: dc, type: 'dc' } });
  }

  // Add Rack compound nodes
  for (const [dc, racks] of Object.entries(topo.groups.racks ?? {})) {
    for (const rack of (racks ?? [])) {
      elements.push({ data: { id: `rack:${dc}:${rack}`, label: rack, type: 'rack', parent: `dc:${dc}` } });
    }
  }

  // Add nodes
  for (const node of (topo.nodes ?? [])) {
    if (node.graph.hidden) continue;
    let parent: string | undefined;
    if (node.graph.dc && node.graph.rack) {
      parent = `rack:${node.graph.dc}:${node.graph.rack}`;
    } else if (node.graph.dc) {
      parent = `dc:${node.graph.dc}`;
    }
    elements.push({
      data: {
        id: node.name,
        label: node.name,
        type: 'node',
        icon: node.graph.icon,
        status: node.status,
        parent,
        nodeData: node,
      }
    });
  }

  // Add edges
  for (const link of (topo.links ?? [])) {
    elements.push({
      data: {
        id: link.id,
        source: link.a.node,
        target: link.z.node,
        state: link.state,
        linkData: link,
      }
    });
  }

  return elements;
}

// Cytoscape stylesheet
export const cytoscapeStyle: cytoscape.StylesheetStyle[] = [
  {
    selector: 'node[type="dc"]',
    style: {
      'background-color': '#e8f4f8',
      'border-color': '#2196F3',
      'border-width': 2,
      'label': 'data(label)',
      'text-valign': 'top',
      'font-size': 14,
      'font-weight': 'bold',
      'color': '#1565C0',
      'text-margin-y': -10,
      'padding': '20px',
    }
  },
  {
    selector: 'node[type="rack"]',
    style: {
      'background-color': '#f5f5f5',
      'border-color': '#9E9E9E',
      'border-width': 1,
      'border-style': 'dashed',
      'label': 'data(label)',
      'text-valign': 'top',
      'font-size': 12,
      'padding': '15px',
    }
  },
  {
    selector: 'node[type="node"]',
    style: {
      'background-color': '#fff',
      'border-color': '#555',
      'border-width': 1,
      'label': 'data(label)',
      'text-valign': 'bottom',
      'text-margin-y': 5,
      'font-size': 11,
      'width': 40,
      'height': 40,
      'shape': 'roundrectangle',
    }
  },
  {
    selector: 'node[status="running"]',
    style: { 'border-color': '#4CAF50', 'border-width': 2 }
  },
  {
    selector: 'node[status="exited"], node[status="stopped"]',
    style: { 'border-color': '#F44336', 'border-width': 2, 'opacity': 0.7 }
  },
  {
    selector: 'edge',
    style: {
      'line-color': '#4CAF50',
      'width': 2,
      'curve-style': 'bezier',
    }
  },
  {
    selector: 'edge[state="down"]',
    style: { 'line-color': '#F44336', 'line-style': 'dashed' }
  },
  {
    selector: 'edge[state="degraded"]',
    style: { 'line-color': '#FFC107' }
  },
  {
    selector: 'node:selected',
    style: { 'border-color': '#2196F3', 'border-width': 3 }
  }
];
