import { useEffect, useRef, useCallback } from 'react';
import cytoscape, { type Core, type EventObject } from 'cytoscape';
import type { Topology, TopologyNode, TopologyLink } from '../types/topology';
import { cytoscapeStylesheet, defaultLayout } from '../lib/cytoscape-config';

interface Props {
  topology: Topology | null;
  onSelectNode: (node: TopologyNode | null) => void;
  onSelectLink: (link: TopologyLink | null) => void;
  onContextMenuLink: (link: TopologyLink, x: number, y: number) => void;
}

function buildCytoscapeElements(topo: Topology) {
  const elements: cytoscape.ElementDefinition[] = [];

  // DC compound nodes
  for (const dc of topo.groups.dcs) {
    elements.push({
      data: { id: `dc:${dc}`, label: dc, type: 'dc' },
    });
    // Rack compound nodes within DC
    const racks = topo.groups.racks[dc] ?? [];
    for (const rack of racks) {
      elements.push({
        data: { id: `rack:${dc}:${rack}`, label: rack, type: 'rack', parent: `dc:${dc}` },
      });
    }
  }

  // Device nodes
  for (const node of topo.nodes) {
    const g = node.graph;
    let parent: string | undefined;
    if (g.dc && g.rack) {
      parent = `rack:${g.dc}:${g.rack}`;
    } else if (g.dc) {
      parent = `dc:${g.dc}`;
    }

    elements.push({
      data: {
        id: node.name,
        label: node.name,
        type: 'device',
        role: g.role || 'host',
        status: node.status,
        parent,
        nodeData: node,
      },
    });
  }

  // Edges
  for (const link of topo.links) {
    const sourceExists = topo.nodes.some((n) => n.name === link.a.node);
    const targetExists = topo.nodes.some((n) => n.name === link.z.node);
    if (!sourceExists || !targetExists) continue;

    elements.push({
      data: {
        id: link.id,
        source: link.a.node,
        target: link.z.node,
        label: `${link.a.interface} — ${link.z.interface}`,
        state: link.state,
        linkData: link,
      },
    });
  }

  return elements;
}

export function TopologyView({ topology, onSelectNode, onSelectLink, onContextMenuLink }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<Core | null>(null);

  const handleTap = useCallback(
    (e: EventObject) => {
      const target = e.target;
      if (target === cyRef.current) {
        onSelectNode(null);
        onSelectLink(null);
        return;
      }
      if (target.isNode() && target.data('type') === 'device') {
        onSelectNode(target.data('nodeData'));
        onSelectLink(null);
      } else if (target.isEdge()) {
        onSelectLink(target.data('linkData'));
        onSelectNode(null);
      }
    },
    [onSelectNode, onSelectLink]
  );

  const handleCxttap = useCallback(
    (e: EventObject) => {
      const target = e.target;
      if (target.isEdge()) {
        const linkData = target.data('linkData') as TopologyLink;
        const pos = e.renderedPosition ?? e.position;
        if (pos) {
          onContextMenuLink(linkData, pos.x, pos.y);
        }
      }
    },
    [onContextMenuLink]
  );

  useEffect(() => {
    if (!containerRef.current) return;

    const cy = cytoscape({
      container: containerRef.current,
      style: cytoscapeStylesheet,
      layout: { name: 'preset' },
      minZoom: 0.2,
      maxZoom: 4,
      wheelSensitivity: 0.3,
    });

    cyRef.current = cy;

    return () => {
      cy.destroy();
      cyRef.current = null;
    };
  }, []);

  useEffect(() => {
    const cy = cyRef.current;
    if (!cy) return;

    cy.on('tap', handleTap);
    cy.on('cxttap', handleCxttap);

    return () => {
      cy.off('tap', handleTap);
      cy.off('cxttap', handleCxttap);
    };
  }, [handleTap, handleCxttap]);

  useEffect(() => {
    const cy = cyRef.current;
    if (!cy || !topology) return;

    const elements = buildCytoscapeElements(topology);
    cy.elements().remove();
    cy.add(elements);
    cy.layout(defaultLayout).run();
    cy.fit(undefined, 40);
  }, [topology]);

  return (
    <div className="relative w-full h-full">
      <div ref={containerRef} className="w-full h-full" />

      {/* Grid overlay for NOC feel */}
      <div
        className="absolute inset-0 pointer-events-none animate-scan"
        style={{
          backgroundImage:
            'linear-gradient(rgba(0, 200, 255, 0.03) 1px, transparent 1px), linear-gradient(90deg, rgba(0, 200, 255, 0.03) 1px, transparent 1px)',
          backgroundSize: '40px 40px',
        }}
      />

      {!topology && (
        <div className="absolute inset-0 flex items-center justify-center">
          <div className="text-center">
            <div className="font-mono text-noc-text-dim text-sm tracking-widest uppercase">
              No project selected
            </div>
            <div className="text-2xs text-noc-text-dim mt-1 font-mono">
              Select a project to view topology
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
