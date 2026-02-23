import React, { useEffect, useRef } from 'react';
import cytoscape from 'cytoscape';
import type { TopologyData, NodeInfo, LinkInfo } from '../types/topology';
import { buildElements, cytoscapeStyle } from '../lib/cytoscape-config';

interface Props {
  topology: TopologyData | null;
  onNodeClick: (node: NodeInfo) => void;
  onLinkClick?: (link: LinkInfo) => void;
}

export function TopologyView({ topology, onNodeClick, onLinkClick }: Props) {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<cytoscape.Core | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    // Initialize or update Cytoscape
    if (!cyRef.current) {
      cyRef.current = cytoscape({
        container: containerRef.current,
        elements: [],
        style: cytoscapeStyle,
        layout: { name: 'preset' },
        minZoom: 0.1,
        maxZoom: 5,
      });
    }

    const cy = cyRef.current;

    if (!topology) {
      cy.elements().remove();
      return;
    }

    // Update elements
    cy.elements().remove();
    cy.add(buildElements(topology));

    // Apply layout
    cy.layout({
      name: 'cose',
      animate: false,
      nodeOverlap: 20,
      fit: true,
      padding: 30,
    } as cytoscape.LayoutOptions).run();

    // Node click handler
    cy.on('tap', 'node[type="node"]', (evt) => {
      const nodeData = evt.target.data('nodeData') as NodeInfo;
      if (nodeData) onNodeClick(nodeData);
    });

    // Edge click handler
    cy.on('tap', 'edge', (evt) => {
      const linkData = evt.target.data('linkData') as LinkInfo;
      if (linkData && onLinkClick) onLinkClick(linkData);
    });

    return () => {
      cy.removeAllListeners();
    };
  }, [topology, onNodeClick, onLinkClick]);

  useEffect(() => {
    return () => {
      cyRef.current?.destroy();
      cyRef.current = null;
    };
  }, []);

  return (
    <div
      ref={containerRef}
      className="topology-view"
      style={{ width: '100%', height: '100%', background: '#fafafa' }}
    />
  );
}
